package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveIntegrationAction_UsesArgsWhenConfigIsEmpty(t *testing.T) {
	action := resolveIntegrationAction("", map[string]interface{}{"action": "logs"})
	if action != "logs" {
		t.Fatalf("expected action logs, got %q", action)
	}
}

func TestBuildK8sGPTCompletionURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "base url", in: "http://gateway:8080", want: "http://gateway:8080/v1/completions"},
		{name: "trailing slash", in: "http://gateway:8080/", want: "http://gateway:8080/v1/completions"},
		{name: "already completion url", in: "http://gateway:8080/v1/completions", want: "http://gateway:8080/v1/completions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildK8sGPTCompletionURL(tt.in); got != tt.want {
				t.Fatalf("buildK8sGPTCompletionURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidateK8sGPTCommandRejectsManagedBackendOverride(t *testing.T) {
	err := validateK8sGPTCommand([]string{"analyze", "--backend", "openai", "--explain"})
	if err == nil {
		t.Fatal("expected backend override to be rejected")
	}
}

func TestValidateK8sGPTCommandRejectsUnsupportedTopLevel(t *testing.T) {
	err := validateK8sGPTCommand([]string{"serve", "--mcp"})
	if err == nil {
		t.Fatal("expected serve to be rejected")
	}
}

func TestAugmentK8sGPTAnalyzeCommandAddsGatewayHeaders(t *testing.T) {
	command := []string{"analyze", "--explain"}
	cfg := k8sgptBackendConfig{
		BaseURL: "http://gateway:8080/v1/completions",
		Model:   "auto",
		CustomHeader: []string{
			"Authorization:Bearer test-token",
			"X-Workspace-ID:ws-123",
		},
	}

	prepared := augmentK8sGPTAnalyzeCommand(command, cfg)

	if !hasFlag(prepared, "--backend") {
		t.Fatalf("expected analyze command to include backend, got %v", prepared)
	}
	if !hasCustomHeader(prepared, "Authorization:Bearer test-token") {
		t.Fatalf("expected Authorization header, got %v", prepared)
	}
	if !hasCustomHeader(prepared, "X-Workspace-ID:ws-123") {
		t.Fatalf("expected workspace header, got %v", prepared)
	}
}

func TestBuildK8sGPTHomeDirSanitizesIdentifiers(t *testing.T) {
	home := buildK8sGPTHomeDir(RuntimeContext{
		WorkspaceID: "ws/123",
		AgentID:     "agent:prod",
		ExecutionID: "exec value",
	})

	want := filepath.Join("/tmp/passflow", "k8sgpt", "ws_123", "agent_prod", "exec_value")
	if home != want {
		t.Fatalf("buildK8sGPTHomeDir() = %q, want %q", home, want)
	}
}

func TestLatestK8sGPTDumpReturnsNewestFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dump_20240101000000.json"), []byte(`{"old":true}`), 0o644); err != nil {
		t.Fatalf("write old dump: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dump_20240102000000.json"), []byte(`{"new":true}`), 0o644); err != nil {
		t.Fatalf("write new dump: %v", err)
	}

	path, data, ok := latestK8sGPTDump(dir)
	if !ok {
		t.Fatal("expected dump file to be detected")
	}
	if filepath.Base(path) != "dump_20240102000000.json" {
		t.Fatalf("expected latest dump file, got %s", filepath.Base(path))
	}
	if data != `{"new":true}` {
		t.Fatalf("expected latest dump data, got %s", data)
	}
}
