package acf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SystemToolProvider implements Claude Code built-in tools (Bash, Read, Write, Edit, Grep, Glob, LS)
// inside the passflow-agent-executor pod sandbox. All file operations are restricted to workDir,
// and all tool calls are gated through the workspace Policy.
type SystemToolProvider struct {
	workDir string
	policy  *Policy
}

// NewSystemToolProvider creates a provider scoped to workDir.
// Pass nil for policy to allow all tools with default limits.
func NewSystemToolProvider(workDir string, policy *Policy) *SystemToolProvider {
	return &SystemToolProvider{workDir: workDir, policy: policy}
}

// Execute runs a tool call and always returns a non-nil *ToolResult.
// Errors (policy violations, missing params, path traversal) are surfaced
// via ToolResult.Error rather than the Go error return so the ReAct loop
// can feed them back to the LLM as observation text.
func (s *SystemToolProvider) Execute(ctx context.Context, call *ToolCall) (*ToolResult, error) {
	if err := s.validatePolicy(call); err != nil {
		return &ToolResult{CallID: call.CallID, Error: err.Error()}, nil
	}

	var result *ToolResult
	switch call.Tool {
	case "Bash":
		result = s.execBash(ctx, call)
	case "Read":
		result = s.execRead(call)
	case "Write":
		result = s.execWrite(call)
	case "Edit":
		result = s.execEdit(call)
	case "Grep":
		result = s.execGrep(call)
	case "Glob":
		result = s.execGlob(call)
	case "LS":
		result = s.execLS(call)
	default:
		result = &ToolResult{Error: fmt.Sprintf("unknown system tool: %q", call.Tool)}
	}

	result.CallID = call.CallID
	return result, nil
}

// validatePolicy returns an error if the tool call is denied by policy.
func (s *SystemToolProvider) validatePolicy(call *ToolCall) error {
	if s.policy == nil {
		return nil
	}
	for _, denied := range s.policy.DeniedTools {
		if denied == call.Tool {
			return fmt.Errorf("tool %q denied by workspace policy", call.Tool)
		}
	}
	if len(s.policy.AllowedTools) > 0 {
		for _, a := range s.policy.AllowedTools {
			if a == call.Tool {
				return nil
			}
		}
		return fmt.Errorf("tool %q not in allowed list", call.Tool)
	}
	return nil
}

// validatePath ensures path resolves to a location inside workDir.
func (s *SystemToolProvider) validatePath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	// Ensure the resolved path is under workDir (guard against symlink / .. tricks).
	workDir := filepath.Clean(s.workDir)
	if !strings.HasPrefix(abs, workDir+string(filepath.Separator)) && abs != workDir {
		return fmt.Errorf("path %q is outside work directory", path)
	}
	return nil
}

func (s *SystemToolProvider) getStringParam(call *ToolCall, key string) string {
	if call.Params == nil {
		return ""
	}
	v, _ := call.Params[key].(string)
	return v
}

// execBash runs an arbitrary shell command inside workDir with an optional timeout.
func (s *SystemToolProvider) execBash(ctx context.Context, call *ToolCall) *ToolResult {
	command := s.getStringParam(call, "command")
	if command == "" {
		return &ToolResult{Error: "command parameter required"}
	}

	timeout := 120 * time.Second
	if s.policy != nil && s.policy.MaxBashTimeout > 0 {
		timeout = time.Duration(s.policy.MaxBashTimeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command) //nolint:gosec
	cmd.Dir = s.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR:\n" + stderr.String()
	}
	if err != nil {
		return &ToolResult{Output: output, Error: err.Error()}
	}
	return &ToolResult{Output: output}
}

// execRead reads a file and returns its contents.
func (s *SystemToolProvider) execRead(call *ToolCall) *ToolResult {
	path := s.getStringParam(call, "file_path")
	if path == "" {
		return &ToolResult{Error: "file_path parameter required"}
	}
	if err := s.validatePath(path); err != nil {
		return &ToolResult{Error: err.Error()}
	}
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return &ToolResult{Error: err.Error()}
	}
	return &ToolResult{Output: string(data)}
}

// execWrite writes content to a file, creating parent directories as needed.
func (s *SystemToolProvider) execWrite(call *ToolCall) *ToolResult {
	path := s.getStringParam(call, "file_path")
	content := s.getStringParam(call, "content")
	if path == "" {
		return &ToolResult{Error: "file_path parameter required"}
	}
	if err := s.validatePath(path); err != nil {
		return &ToolResult{Error: err.Error()}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return &ToolResult{Error: err.Error()}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil { //nolint:gosec
		return &ToolResult{Error: err.Error()}
	}
	return &ToolResult{Output: fmt.Sprintf("Written %d bytes to %s", len(content), path)}
}

// execEdit performs a single exact-string replacement in a file.
func (s *SystemToolProvider) execEdit(call *ToolCall) *ToolResult {
	path := s.getStringParam(call, "file_path")
	oldStr := s.getStringParam(call, "old_string")
	newStr := s.getStringParam(call, "new_string")
	if path == "" || oldStr == "" {
		return &ToolResult{Error: "file_path and old_string parameters required"}
	}
	if err := s.validatePath(path); err != nil {
		return &ToolResult{Error: err.Error()}
	}
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return &ToolResult{Error: err.Error()}
	}
	content := string(data)
	if !strings.Contains(content, oldStr) {
		return &ToolResult{Error: "old_string not found in file"}
	}
	newContent := strings.Replace(content, oldStr, newStr, 1)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil { //nolint:gosec
		return &ToolResult{Error: err.Error()}
	}
	return &ToolResult{Output: "Edit applied successfully"}
}

// execGrep searches all files under workDir for lines matching a regex pattern.
func (s *SystemToolProvider) execGrep(call *ToolCall) *ToolResult {
	pattern := s.getStringParam(call, "pattern")
	if pattern == "" {
		return &ToolResult{Error: "pattern parameter required"}
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("invalid regex: %s", err)}
	}

	var matches []string
	_ = filepath.Walk(s.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Size() > 1024*1024 {
			return nil
		}
		data, err := os.ReadFile(path) //nolint:gosec
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(s.workDir, path)
		for i, line := range strings.Split(string(data), "\n") {
			if re.MatchString(line) {
				matches = append(matches, fmt.Sprintf("%s:%d:%s", rel, i+1, line))
			}
		}
		return nil
	})

	if len(matches) == 0 {
		return &ToolResult{Output: "No matches found"}
	}
	return &ToolResult{Output: strings.Join(matches, "\n")}
}

// execGlob expands a glob pattern relative to workDir and returns matching paths.
func (s *SystemToolProvider) execGlob(call *ToolCall) *ToolResult {
	pattern := s.getStringParam(call, "pattern")
	if pattern == "" {
		return &ToolResult{Error: "pattern parameter required"}
	}
	fullPattern := filepath.Join(s.workDir, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return &ToolResult{Error: err.Error()}
	}
	var relPaths []string
	for _, m := range matches {
		rel, _ := filepath.Rel(s.workDir, m)
		relPaths = append(relPaths, rel)
	}
	if len(relPaths) == 0 {
		return &ToolResult{Output: "No matches found"}
	}
	return &ToolResult{Output: strings.Join(relPaths, "\n")}
}

// execLS lists directory contents, defaulting to workDir.
func (s *SystemToolProvider) execLS(call *ToolCall) *ToolResult {
	path := s.getStringParam(call, "path")
	if path == "" {
		path = s.workDir
	}
	if err := s.validatePath(path); err != nil {
		return &ToolResult{Error: err.Error()}
	}
	entries, err := os.ReadDir(path) //nolint:gosec
	if err != nil {
		return &ToolResult{Error: err.Error()}
	}
	var lines []string
	for _, e := range entries {
		prefix := "  "
		if e.IsDir() {
			prefix = "d "
		}
		lines = append(lines, prefix+e.Name())
	}
	return &ToolResult{Output: strings.Join(lines, "\n")}
}
