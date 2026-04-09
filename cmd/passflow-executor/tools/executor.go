package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jaak-ai/passflow-agent-executor/job"
	"github.com/jaak-ai/passflow-agent-executor/llm"
)

// Result is the outcome of a tool execution.
type Result struct {
	ToolCallID string
	ToolName   string
	Output     string // JSON string returned to the LLM
	Error      string // non-empty on failure
	Mocked     bool   // true when the output came from a shadow-mode mock
}

// RuntimeContext carries execution-scoped identifiers that tools can use for
// isolated temp state or downstream correlation.
type RuntimeContext struct {
	WorkspaceID string
	AgentID     string
	ExecutionID string
}

// Executor runs tool calls on behalf of the agent.
type Executor struct {
	tools             map[string]job.ToolConfig
	httpClient        *http.Client
	shadowInterceptor *ShadowInterceptor
	runtime           RuntimeContext
}

// New creates a new Executor with the tools defined in the job spec.
func New(toolConfigs []job.ToolConfig) *Executor {
	return NewWithRuntime(toolConfigs, RuntimeContext{})
}

// NewWithRuntime creates a new Executor with execution-scoped runtime context.
func NewWithRuntime(toolConfigs []job.ToolConfig, runtime RuntimeContext) *Executor {
	m := make(map[string]job.ToolConfig, len(toolConfigs))
	for _, t := range toolConfigs {
		m[t.Name] = t
	}
	return &Executor{
		tools:      m,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		runtime:    runtime,
	}
}

// SetShadowMode attaches a ShadowInterceptor to this executor. When set,
// ExecuteAll routes each tool call through the interceptor before deciding
// whether to mock, passthrough, or abort. The engine code is unaffected
// because it still calls ExecuteAll on the same *Executor.
func (e *Executor) SetShadowMode(interceptor *ShadowInterceptor) {
	e.shadowInterceptor = interceptor
}

// ShadowStats returns the interception statistics. Returns a zero-value
// ShadowStats if no interceptor is set.
func (e *Executor) ShadowStats() ShadowStats {
	if e.shadowInterceptor == nil {
		return ShadowStats{}
	}
	return e.shadowInterceptor.Stats()
}

// Definitions returns tool definitions for the LLM (function calling schema).
// The slice is sorted by tool name to guarantee a deterministic ordering on
// every call, since the underlying storage is a map whose iteration order is
// intentionally random in Go.
func (e *Executor) Definitions() []llm.ToolDefinition {
	defs := make([]llm.ToolDefinition, 0, len(e.tools))
	for _, t := range e.tools {
		params := t.Parameters
		if params == nil {
			params = map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			}
		}
		defs = append(defs, llm.ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  params,
		})
	}
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].Name < defs[j].Name
	})
	return defs
}

// CanHandle reports whether the executor knows about the tool with the given
// name.  The ReAct loop uses this to decide whether a call should be handled
// by the regular executor or delegated to an alternative router (e.g. ACF).
func (e *Executor) CanHandle(name string) bool {
	_, ok := e.tools[name]
	return ok
}

// Execute runs a single tool call and returns the result.
func (e *Executor) Execute(ctx context.Context, call llm.ToolCall) Result {
	toolCfg, ok := e.tools[call.Name]
	if !ok {
		return Result{
			ToolCallID: call.ID,
			ToolName:   call.Name,
			Error:      fmt.Sprintf("tool %q not found", call.Name),
			Output:     fmt.Sprintf(`{"error":"tool %q not found"}`, call.Name),
		}
	}

	// Parse arguments from LLM
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return Result{
			ToolCallID: call.ID,
			ToolName:   call.Name,
			Error:      "invalid arguments: " + err.Error(),
			Output:     `{"error":"invalid arguments"}`,
		}
	}

	switch toolCfg.Type {
	case "http":
		return e.runHTTP(ctx, call.ID, call.Name, toolCfg.HTTPConfig, args)
	default:
		return e.runIntegration(ctx, call.ID, call.Name, toolCfg.IntegrationConfig, args)
	}
}

// ExecuteAll runs all tool calls in a response concurrently and returns results.
// When a shadow interceptor is attached via SetShadowMode, each call is routed
// through the interceptor: mocked calls return the canned output, passthrough
// calls execute normally, and aborted calls return an error.
func (e *Executor) ExecuteAll(ctx context.Context, calls []llm.ToolCall) []Result {
	if e.shadowInterceptor != nil {
		return e.executeAllShadow(ctx, calls)
	}
	results := make([]Result, len(calls))
	var wg sync.WaitGroup
	for i, call := range calls {
		wg.Add(1)
		go func(idx int, c llm.ToolCall) {
			defer wg.Done()
			results[idx] = e.Execute(ctx, c)
		}(i, call)
	}
	wg.Wait()
	return results
}

// executeAllShadow runs tool calls through the shadow interceptor sequentially.
// Sequential execution is required because an abort on one tool must prevent
// execution of subsequent tools in the same batch.
func (e *Executor) executeAllShadow(ctx context.Context, calls []llm.ToolCall) []Result {
	results := make([]Result, len(calls))
	for i, call := range calls {
		results[i] = e.executeShadowCall(ctx, call)
		// If this call caused an abort, fill remaining results with abort errors.
		if results[i].Error != "" && strings.Contains(results[i].Error, "shadow mode") {
			for j := i + 1; j < len(calls); j++ {
				results[j] = Result{
					ToolCallID: calls[j].ID,
					ToolName:   calls[j].Name,
					Error:      "shadow mode aborted",
					Output:     `{"error":"shadow mode aborted"}`,
				}
			}
			break
		}
	}
	return results
}

// executeShadowCall handles a single tool call in shadow mode.
func (e *Executor) executeShadowCall(ctx context.Context, call llm.ToolCall) Result {
	// Determine the HTTP method if this is an HTTP tool (needed for classification).
	var httpMethod string
	if cfg, ok := e.tools[call.Name]; ok && cfg.HTTPConfig != nil {
		httpMethod = strings.ToUpper(cfg.HTTPConfig.Method)
	}

	mockOutput, executeReal, abort := e.shadowInterceptor.Intercept(call.Name, httpMethod)

	if abort {
		return Result{
			ToolCallID: call.ID,
			ToolName:   call.Name,
			Error:      fmt.Sprintf("shadow mode: tool %q aborted (no mock available)", call.Name),
			Output:     fmt.Sprintf(`{"error":"shadow mode: tool %q cannot execute without mock"}`, call.Name),
		}
	}

	if !executeReal {
		// Return the mocked output.
		return Result{
			ToolCallID: call.ID,
			ToolName:   call.Name,
			Output:     mockOutput,
			Mocked:     true,
		}
	}

	// Passthrough: execute the tool for real.
	return e.Execute(ctx, call)
}

// privateIPNets contains CIDR blocks that must never be targeted by HTTP tools.
// This guards against SSRF attacks where an LLM (via prompt injection) directs
// the executor to reach internal services such as the K8s API server, cloud
// instance-metadata endpoints, Redis, or other cluster-internal resources.
var privateIPNets = func() []*net.IPNet {
	cidrs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16", // link-local / AWS+GCP metadata
		"::1/128",        // IPv6 loopback
		"fd00::/8",       // IPv6 unique local
		"fc00::/7",       // IPv6 unique local (full range)
	}
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			nets = append(nets, ipNet)
		}
	}
	return nets
}()

// blockedHostnames lists hostnames that must never be contacted regardless of
// the IP they resolve to. This catches Kubernetes service-discovery names.
var blockedHostnames = []string{
	"kubernetes.default.svc",
	"kubernetes.default.svc.cluster.local",
	"localhost",
}

// validateURL checks that u is a safe, externally-reachable URL.
// It rejects:
//   - non-http(s) schemes (file://, ftp://, gopher://, etc.)
//   - hostnames on the blocklist (kubernetes.default.svc, localhost, …)
//   - IP addresses in private / link-local / loopback CIDR ranges
//
// This is the primary defence against SSRF via the HTTP tool type.
func validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL scheme %q is not allowed; only http and https are permitted", parsed.Scheme)
	}

	host := parsed.Hostname() // strips port

	// Check against blocked hostnames first (fast path, no DNS).
	for _, blocked := range blockedHostnames {
		if strings.EqualFold(host, blocked) {
			return fmt.Errorf("URL host %q is not allowed (SSRF block)", host)
		}
	}

	// If the host is already an IP literal, validate it directly.
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return fmt.Errorf("URL resolves to a private/reserved IP address %s (SSRF block)", ip)
		}
		return nil
	}

	// For hostnames, resolve and check every returned address.
	addrs, err := net.LookupHost(host)
	if err != nil {
		// If DNS fails, block the request — we cannot verify safety.
		return fmt.Errorf("could not resolve host %q: %w", host, err)
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip != nil && isPrivateIP(ip) {
			return fmt.Errorf("URL host %q resolves to a private/reserved IP address %s (SSRF block)", host, ip)
		}
	}

	return nil
}

// isPrivateIP reports whether ip falls within any of the blocked CIDR ranges.
func isPrivateIP(ip net.IP) bool {
	for _, network := range privateIPNets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// renderBodyTemplate performs simple key-value substitution on tmpl.
// Placeholders have the form {{.key}} and are replaced with the corresponding
// value from args. Unknown keys and any other {{...}} expressions are replaced
// with an empty string.
//
// This deliberately avoids text/template to prevent template injection: a
// text/template body template sourced from job spec config could call built-in
// functions (printf, call, etc.) or methods on the args map, giving an
// attacker who controls the job spec arbitrary code execution within the
// template engine context. By using strings.NewReplacer, we guarantee the
// substitution is a pure string operation with no evaluation semantics.
func renderBodyTemplate(tmpl string, args map[string]interface{}) (string, error) {
	pairs := make([]string, 0, len(args)*2)
	for k, v := range args {
		placeholder := "{{." + k + "}}"
		pairs = append(pairs, placeholder, fmt.Sprintf("%v", v))
	}
	r := strings.NewReplacer(pairs...)
	result := r.Replace(tmpl)

	// Strip any remaining {{...}} blocks — these are either unreferenced
	// {{.unknownKey}} placeholders or non-key expressions like
	// {{printf ...}}, {{call .}}, etc. None of them should ever be executed;
	// we simply discard them. This is the second line of defence against
	// template injection: even if a malicious expression reaches this function,
	// it is removed as text rather than evaluated.
	for strings.Contains(result, "{{") {
		start := strings.Index(result, "{{")
		end := strings.Index(result[start:], "}}")
		if end == -1 {
			// Unmatched opening delimiter — remove it and stop scanning.
			result = result[:start] + result[start+2:]
			break
		}
		result = result[:start] + result[start+end+2:]
	}

	return result, nil
}

// runHTTP executes a generic HTTP tool call.
func (e *Executor) runHTTP(ctx context.Context, callID, toolName string, cfg *job.HTTPToolConfig, args map[string]interface{}) Result {
	if cfg == nil {
		return Result{ToolCallID: callID, ToolName: toolName, Error: "missing http_config", Output: `{"error":"missing http_config"}`}
	}

	// Validate the target URL before any network activity to prevent SSRF.
	if err := validateURL(cfg.URL); err != nil {
		return Result{ToolCallID: callID, ToolName: toolName, Error: "URL not allowed: " + err.Error(), Output: `{"error":"URL not allowed"}`}
	}

	method := strings.ToUpper(cfg.Method)
	if method == "" {
		method = "GET"
	}

	// Render body template using safe string interpolation (not text/template).
	// text/template allows calling built-in functions and methods on the data
	// context, which enables template injection when the template comes from
	// untrusted job spec config.
	var bodyReader io.Reader
	if cfg.BodyTemplate != "" {
		rendered, err := renderBodyTemplate(cfg.BodyTemplate, args)
		if err != nil {
			return Result{ToolCallID: callID, ToolName: toolName, Error: "body render failed: " + err.Error(), Output: `{"error":"body render failed"}`}
		}
		bodyReader = bytes.NewBufferString(rendered)
	} else if method == "POST" || method == "PUT" || method == "PATCH" {
		// Default: send args as JSON body
		bodyBytes, _ := json.Marshal(args)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, cfg.URL, bodyReader)
	if err != nil {
		return Result{ToolCallID: callID, ToolName: toolName, Error: "failed to create request: " + err.Error(), Output: `{"error":"request creation failed"}`}
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return Result{ToolCallID: callID, ToolName: toolName, Error: "request failed: " + err.Error(), Output: fmt.Sprintf(`{"error":"request failed: %s"}`, err.Error())}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024)) // 64KB limit
	if err != nil {
		return Result{ToolCallID: callID, ToolName: toolName, Error: "failed to read response", Output: `{"error":"failed to read response"}`}
	}

	output := fmt.Sprintf(`{"status":%d,"body":%s}`, resp.StatusCode, sanitizeJSON(string(respBody)))
	return Result{ToolCallID: callID, ToolName: toolName, Output: output}
}

// runIntegration executes an integration-based tool call.
func (e *Executor) runIntegration(ctx context.Context, callID, toolName string, cfg *job.IntegrationToolConfig, args map[string]interface{}) Result {
	if cfg == nil {
		return Result{ToolCallID: callID, ToolName: toolName, Error: "missing integration_config", Output: `{"error":"missing integration_config"}`}
	}

	// Dispatch to integration-specific handler
	var output string
	var execErr error
	action := resolveIntegrationAction(cfg.Action, args)

	switch cfg.Provider {
	case "slack":
		output, execErr = runSlackAction(ctx, action, cfg.Credentials, args)
	case "github":
		output, execErr = runGitHubAction(ctx, action, cfg.Credentials, args)
	case "linear":
		output, execErr = runLinearAction(ctx, action, cfg.Credentials, args)
	case "notion":
		output, execErr = runNotionAction(ctx, action, cfg.Credentials, args)
	case "kubernetes":
		output, execErr = runKubernetesAction(ctx, action, cfg.Credentials, args)
	case "k8sgpt":
		output, execErr = e.runK8sGPTAction(ctx, action, cfg.Credentials, args)
	case "git":
		output, execErr = runGitAction(ctx, action, cfg.Credentials, args)
	case "google_drive":
		output, execErr = runGoogleDriveAction(ctx, action, cfg.Credentials, args)
	case "prometheus":
		output, execErr = runPrometheusAction(ctx, action, cfg.Credentials, args)
	case "loki":
		output, execErr = runLokiAction(ctx, action, cfg.Credentials, args)
	default:
		// Generic fallback: call the integration's API URL if available
		execErr = fmt.Errorf("unsupported integration provider: %q", cfg.Provider)
	}

	if execErr != nil {
		errMsg := execErr.Error()
		return Result{
			ToolCallID: callID,
			ToolName:   toolName,
			Error:      errMsg,
			Output:     fmt.Sprintf(`{"error":%q}`, errMsg),
		}
	}

	return Result{ToolCallID: callID, ToolName: toolName, Output: output}
}

func resolveIntegrationAction(configured string, args map[string]interface{}) string {
	if strings.TrimSpace(configured) != "" {
		return configured
	}
	if args == nil {
		return ""
	}
	action, _ := args["action"].(string)
	return strings.TrimSpace(action)
}

// sanitizeJSON ensures a string is valid JSON (wraps in quotes if not).
func sanitizeJSON(s string) string {
	var js json.RawMessage
	if json.Unmarshal([]byte(s), &js) == nil {
		return s
	}
	quoted, _ := json.Marshal(s)
	return string(quoted)
}
