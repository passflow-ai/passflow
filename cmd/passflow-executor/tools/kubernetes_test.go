package tools

import (
	"context"
	"strings"
	"testing"
)

func TestNewK8sClientOutOfCluster(t *testing.T) {
	// When running outside cluster, should return error
	_, err := NewK8sClient()
	if err == nil {
		t.Skip("Running inside cluster, skipping out-of-cluster test")
	}
	// Expected to fail when not in cluster
	t.Logf("Expected error when out of cluster: %v", err)
}

func TestK8sActionValidation(t *testing.T) {
	tests := []struct {
		action    string
		args      map[string]interface{}
		expectErr bool
	}{
		{
			action:    "get",
			args:      map[string]interface{}{"resource": "pods", "namespace": "default"},
			expectErr: false,
		},
		{
			action:    "get",
			args:      map[string]interface{}{"resource": "", "namespace": "default"},
			expectErr: true,
		},
		{
			action:    "apply",
			args:      map[string]interface{}{"manifest": "", "namespace": "default"},
			expectErr: true,
		},
		{
			action:    "invalid_action",
			args:      map[string]interface{}{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			err := validateK8sAction(tt.action, tt.args)
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestExecCommandAllowlist verifies that only permitted commands are accepted
// and that disallowed commands are rejected before any network call is made.
func TestExecCommandAllowlist(t *testing.T) {
	// We cannot reach a real cluster in unit tests, so we exercise the
	// allowlist guard through a stub K8sClient whose clientset is nil.
	// The allowlist check runs before the network call, so a nil clientset
	// is safe for these cases.
	client := &K8sClient{}
	ctx := context.Background()

	// cat and env are intentionally removed from the allowlist (HIGH 2 security fix):
	// cat can read arbitrary files (service-account tokens, /etc/shadow, etc.)
	// env exposes environment variables that may contain API keys.
	allowedCases := []string{"ls", "echo", "whoami", "date", "uname", "df", "free", "ps", "pwd", "id", "uptime", "which"}
	for _, cmd := range allowedCases {
		t.Run("allowed_"+cmd, func(t *testing.T) {
			defer func() {
				// Allowed commands pass the allowlist but panic on nil clientset
				// (no real cluster). A panic here means the allowlist accepted
				// the command — which is what we're testing.
				recover()
			}()
			_, err := client.exec(ctx, map[string]interface{}{
				"pod":     "test-pod",
				"command": []interface{}{cmd, "-l"},
			})
			// The command passed the allowlist. Any error here is a network
			// error (no real cluster), not a rejection.
			if err != nil && strings.Contains(err.Error(), "is not allowed") {
				t.Errorf("command %q should be in allowlist but was rejected: %v", cmd, err)
			}
		})
	}

	// cat and env are now explicitly blocked (HIGH 2 security fix).
	disallowedCases := []string{"sh", "bash", "curl", "wget", "nc", "python", "python3", "perl", "ruby", "rm", "chmod", "chown", "cat", "env"}
	for _, cmd := range disallowedCases {
		t.Run("disallowed_"+cmd, func(t *testing.T) {
			_, err := client.exec(ctx, map[string]interface{}{
				"pod":     "test-pod",
				"command": []interface{}{cmd},
			})
			if err == nil {
				t.Errorf("command %q should have been rejected but was accepted", cmd)
			}
			if !strings.Contains(err.Error(), "is not allowed") {
				t.Errorf("expected 'is not allowed' error for %q, got: %v", cmd, err)
			}
		})
	}
}

// TestExecCommandAllowlistEmptyCommand verifies that an empty command slice
// returns an error rather than panicking.
func TestExecCommandAllowlistEmptyCommand(t *testing.T) {
	client := &K8sClient{}
	ctx := context.Background()

	_, err := client.exec(ctx, map[string]interface{}{
		"pod":     "test-pod",
		"command": []interface{}{},
	})
	if err == nil {
		t.Fatal("expected error for empty command, got nil")
	}
}

// TestGetSecretsBlocked verifies that the "secrets" resource type is no longer
// accessible through the get action.
func TestGetSecretsBlocked(t *testing.T) {
	client := &K8sClient{}
	ctx := context.Background()

	for _, resource := range []string{"secrets", "secret"} {
		t.Run(resource, func(t *testing.T) {
			_, err := client.get(ctx, map[string]interface{}{
				"resource":  resource,
				"namespace": "default",
			})
			if err == nil {
				t.Errorf("expected error when getting resource %q (secrets must be blocked)", resource)
			}
			if !strings.Contains(err.Error(), "unsupported resource type") {
				t.Errorf("expected 'unsupported resource type' error for %q, got: %v", resource, err)
			}
		})
	}
}
