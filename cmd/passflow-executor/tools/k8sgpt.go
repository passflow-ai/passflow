package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultK8sGPTTimeoutSeconds = 120
	maxK8sGPTTimeoutSeconds     = 300
	defaultK8sGPTModel          = "auto"
)

var (
	k8sgptLookPath       = exec.LookPath
	k8sgptCommandContext = exec.CommandContext
)

type k8sgptBackendConfig struct {
	BaseURL      string
	Model        string
	ServiceToken string
	WorkspaceID  string
	CustomHeader []string
}

func (e *Executor) runK8sGPTAction(ctx context.Context, _ string, creds map[string]string, args map[string]interface{}) (string, error) {
	command, timeoutSeconds, err := parseK8sGPTArgs(args)
	if err != nil {
		return "", fmt.Errorf("k8sgpt: %w", err)
	}
	if err := validateK8sGPTCommand(command); err != nil {
		return "", fmt.Errorf("k8sgpt: %w", err)
	}

	binPath, err := k8sgptLookPath("k8sgpt")
	if err != nil {
		return "", fmt.Errorf("k8sgpt: binary not found in PATH: %w", err)
	}

	homeDir := buildK8sGPTHomeDir(e.runtime)
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		return "", fmt.Errorf("k8sgpt: failed to create config directory: %w", err)
	}

	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	env := append(os.Environ(),
		"HOME="+homeDir,
		"NO_COLOR=1",
	)

	prepared := append([]string(nil), command...)
	if k8sgptPrimaryCommand(prepared) == "analyze" {
		if k8sgptNeedsGateway(prepared) {
			backendCfg, err := resolveK8sGPTBackendConfig(e.runtime, creds)
			if err != nil {
				return "", fmt.Errorf("k8sgpt: %w", err)
			}
			if err := ensureK8sGPTBackend(cmdCtx, binPath, homeDir, env, backendCfg); err != nil {
				return "", err
			}
			prepared = augmentK8sGPTAnalyzeCommand(prepared, backendCfg)
		}
		if !hasFlag(prepared, "--output") {
			prepared = append(prepared, "--output=json")
		}
	}

	stdout, stderr, err := executeK8sGPT(cmdCtx, binPath, homeDir, env, prepared)
	if err != nil {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = strings.TrimSpace(stdout)
		}
		if msg != "" {
			return "", fmt.Errorf("k8sgpt command failed: %w: %s", err, msg)
		}
		return "", fmt.Errorf("k8sgpt command failed: %w", err)
	}

	result := map[string]interface{}{
		"command":     append([]string{"k8sgpt"}, prepared...),
		"config_home": homeDir,
	}

	if trimmed := strings.TrimSpace(stderr); trimmed != "" {
		result["stderr"] = trimmed
	}

	if k8sgptPrimaryCommand(prepared) == "dump" {
		if dumpPath, dumpData, ok := latestK8sGPTDump(homeDir); ok {
			result["dump_file"] = dumpPath
			if parsed, ok := decodeJSONPayload(dumpData); ok {
				result["result"] = parsed
			} else {
				result["stdout"] = dumpData
			}
		}
	}

	if _, hasResult := result["result"]; !hasResult {
		if parsed, ok := decodeJSONPayload(stdout); ok {
			result["result"] = parsed
		} else if trimmed := strings.TrimSpace(stdout); trimmed != "" {
			result["stdout"] = trimmed
		}
	}

	out, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return "", fmt.Errorf("k8sgpt: failed to encode result: %w", marshalErr)
	}
	return string(out), nil
}

func parseK8sGPTArgs(args map[string]interface{}) ([]string, int, error) {
	rawCommand, ok := args["command"]
	if !ok {
		return nil, 0, fmt.Errorf("command is required")
	}

	command, err := interfaceSliceToStrings(rawCommand)
	if err != nil {
		return nil, 0, fmt.Errorf("command: %w", err)
	}
	if len(command) == 0 {
		return nil, 0, fmt.Errorf("command cannot be empty")
	}

	timeoutSeconds := defaultK8sGPTTimeoutSeconds
	if raw, ok := args["timeout_seconds"]; ok {
		switch v := raw.(type) {
		case float64:
			timeoutSeconds = int(v)
		case int:
			timeoutSeconds = v
		}
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultK8sGPTTimeoutSeconds
	}
	if timeoutSeconds > maxK8sGPTTimeoutSeconds {
		timeoutSeconds = maxK8sGPTTimeoutSeconds
	}

	return command, timeoutSeconds, nil
}

func interfaceSliceToStrings(v interface{}) ([]string, error) {
	switch values := v.(type) {
	case []string:
		out := make([]string, 0, len(values))
		for _, item := range values {
			if strings.TrimSpace(item) == "" {
				return nil, fmt.Errorf("contains empty argument")
			}
			out = append(out, item)
		}
		return out, nil
	case []interface{}:
		out := make([]string, 0, len(values))
		for _, item := range values {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("must be an array of strings")
			}
			if strings.TrimSpace(s) == "" {
				return nil, fmt.Errorf("contains empty argument")
			}
			out = append(out, s)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("must be an array of strings")
	}
}

func validateK8sGPTCommand(command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("command is required")
	}
	if strings.HasPrefix(command[0], "-") {
		return fmt.Errorf("top-level subcommand is required")
	}

	allowed := map[string]bool{
		"analyze":         true,
		"filters":         true,
		"integrations":    true,
		"dump":            true,
		"version":         true,
		"cache":           true,
		"custom-analyzer": true,
	}
	if !allowed[command[0]] {
		return fmt.Errorf("unsupported top-level command %q", command[0])
	}

	switch command[0] {
	case "serve", "auth", "generate":
		return fmt.Errorf("command %q is not available through the agent tool", command[0])
	}

	if command[0] == "analyze" && (hasFlag(command, "--backend") || hasFlag(command, "-b")) {
		return fmt.Errorf("do not pass --backend; the Passflow gateway backend is managed automatically")
	}

	return nil
}

func buildK8sGPTHomeDir(runtime RuntimeContext) string {
	workspace := sanitizePathSegment(runtime.WorkspaceID)
	if workspace == "" {
		workspace = "default-workspace"
	}
	agent := sanitizePathSegment(runtime.AgentID)
	if agent == "" {
		agent = "default-agent"
	}
	execution := sanitizePathSegment(runtime.ExecutionID)
	if execution == "" {
		execution = "default-execution"
	}
	return filepath.Join("/tmp/passflow", "k8sgpt", workspace, agent, execution)
}

func sanitizePathSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

func resolveK8sGPTBackendConfig(runtime RuntimeContext, creds map[string]string) (k8sgptBackendConfig, error) {
	baseURL := firstNonEmpty(
		creds["base_url"],
		creds["gateway_url"],
		os.Getenv("K8SGPT_BACKEND_URL"),
	)
	if baseURL == "" {
		baseURL = buildK8sGPTCompletionURL(os.Getenv("LLM_GATEWAY_URL"))
	} else {
		baseURL = buildK8sGPTCompletionURL(baseURL)
	}
	if baseURL == "" {
		return k8sgptBackendConfig{}, fmt.Errorf("LLM gateway URL is not configured")
	}

	serviceToken := firstNonEmpty(
		creds["service_token"],
		os.Getenv("SERVICE_TOKEN"),
		os.Getenv("CHRONOS_SERVICE_TOKEN"),
	)
	if serviceToken == "" {
		return k8sgptBackendConfig{}, fmt.Errorf("service token is not configured")
	}

	workspaceID := firstNonEmpty(
		creds["workspace_id"],
		runtime.WorkspaceID,
	)

	headers := []string{
		"Authorization:Bearer " + serviceToken,
	}
	if workspaceID != "" {
		headers = append(headers, "X-Workspace-ID:"+workspaceID)
	}

	return k8sgptBackendConfig{
		BaseURL:      baseURL,
		Model:        firstNonEmpty(creds["model"], os.Getenv("K8SGPT_MODEL"), defaultK8sGPTModel),
		ServiceToken: serviceToken,
		WorkspaceID:  workspaceID,
		CustomHeader: headers,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func buildK8sGPTCompletionURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.TrimRight(raw, "/")
	if strings.HasSuffix(raw, "/v1/completions") {
		return raw
	}
	return raw + "/v1/completions"
}

func k8sgptPrimaryCommand(command []string) string {
	if len(command) == 0 {
		return ""
	}
	return command[0]
}

func k8sgptNeedsGateway(command []string) bool {
	if k8sgptPrimaryCommand(command) != "analyze" {
		return false
	}
	return hasFlag(command, "--explain") || hasFlag(command, "--with-doc")
}

func hasFlag(command []string, flag string) bool {
	for i, arg := range command {
		if arg == flag {
			return true
		}
		if strings.HasPrefix(flag, "--") && strings.HasPrefix(arg, flag+"=") {
			return true
		}
		if (flag == "--output" || flag == "--backend" || flag == "--custom-headers") && i+1 < len(command) && command[i] == flag {
			return true
		}
	}
	return false
}

func ensureK8sGPTBackend(ctx context.Context, binPath, homeDir string, env []string, cfg k8sgptBackendConfig) error {
	marker := filepath.Join(homeDir, ".passflow-k8sgpt-backend.ready")
	if _, err := os.Stat(marker); err == nil {
		return nil
	}

	args := []string{
		"auth", "add",
		"--backend", "customrest",
		"--baseurl", cfg.BaseURL,
		"--model", cfg.Model,
	}
	if cfg.ServiceToken != "" {
		args = append(args, "--password", cfg.ServiceToken)
	}

	stdout, stderr, err := executeK8sGPT(ctx, binPath, homeDir, env, args)
	if err != nil && !strings.Contains(strings.ToLower(stderr), "already") {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = strings.TrimSpace(stdout)
		}
		if msg != "" {
			return fmt.Errorf("k8sgpt: failed to configure customrest backend: %w: %s", err, msg)
		}
		return fmt.Errorf("k8sgpt: failed to configure customrest backend: %w", err)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"base_url": cfg.BaseURL,
		"model":    cfg.Model,
	})
	if writeErr := os.WriteFile(marker, payload, 0o644); writeErr != nil {
		return fmt.Errorf("k8sgpt: failed to persist backend marker: %w", writeErr)
	}
	return nil
}

func augmentK8sGPTAnalyzeCommand(command []string, cfg k8sgptBackendConfig) []string {
	prepared := append([]string(nil), command...)
	prepared = append(prepared, "--backend=customrest")
	for _, header := range cfg.CustomHeader {
		if !hasCustomHeader(prepared, header) {
			prepared = append(prepared, "--custom-headers", header)
		}
	}
	return prepared
}

func hasCustomHeader(command []string, header string) bool {
	for i, arg := range command {
		if arg == "--custom-headers" && i+1 < len(command) && command[i+1] == header {
			return true
		}
		if strings.HasPrefix(arg, "--custom-headers=") && strings.TrimPrefix(arg, "--custom-headers=") == header {
			return true
		}
	}
	return false
}

func executeK8sGPT(ctx context.Context, binPath, homeDir string, env, args []string) (string, string, error) {
	cmd := k8sgptCommandContext(ctx, binPath, args...)
	cmd.Dir = homeDir
	cmd.Env = env

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func latestK8sGPTDump(homeDir string) (string, string, bool) {
	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return "", "", false
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "dump_") && strings.HasSuffix(name, ".json") {
			files = append(files, name)
		}
	}
	if len(files) == 0 {
		return "", "", false
	}
	sort.Strings(files)
	path := filepath.Join(homeDir, files[len(files)-1])
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", false
	}
	return path, strings.TrimSpace(string(data)), true
}

func decodeJSONPayload(raw string) (interface{}, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, false
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil, false
	}
	return parsed, true
}
