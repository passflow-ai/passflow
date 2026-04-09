package tools

import (
	"strings"
	"testing"
)

func TestSanitizeGitError(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantSub string // substring that must be present in output
		badSub  string // substring that must NOT be present in output
	}{
		{
			name:    "token in https URL is redacted",
			input:   "fatal: repository 'https://x-access-token:ghp_SECRETTOKEN123@github.com/org/repo.git/' not found",
			wantSub: "<redacted>@github.com",
			badSub:  "ghp_SECRETTOKEN123",
		},
		{
			name:    "token-only user-info is redacted",
			input:   "error: https://ghp_TOKEN@github.com/org/repo.git did not send all necessary objects",
			wantSub: "<redacted>@github.com",
			badSub:  "ghp_TOKEN",
		},
		{
			name:    "message without URL is unchanged",
			input:   "fatal: not a git repository (or any of the parent directories): .git",
			wantSub: "fatal: not a git repository",
			badSub:  "",
		},
		{
			name:    "empty string is unchanged",
			input:   "",
			wantSub: "",
			badSub:  "",
		},
		{
			name:    "multiple token URLs in same message are all redacted",
			input:   "clone from https://tok:abc123@github.com/a/b and https://tok:xyz789@github.com/c/d failed",
			wantSub: "<redacted>@github.com",
			badSub:  "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeGitError(tt.input)
			if tt.wantSub != "" && !strings.Contains(got, tt.wantSub) {
				t.Errorf("sanitizeGitError(%q) = %q; want substring %q", tt.input, got, tt.wantSub)
			}
			if tt.badSub != "" && strings.Contains(got, tt.badSub) {
				t.Errorf("sanitizeGitError(%q) = %q; must not contain %q", tt.input, got, tt.badSub)
			}
		})
	}
}

func TestValidateRepoPattern(t *testing.T) {
	tests := []struct {
		repo    string
		isValid bool
	}{
		{"owner/repo", true},
		{"my-org/my-repo", true},
		{"org_name/repo.name", true},
		{"owner/repo; rm -rf /", false},
		{"../../../etc/passwd", false},
		{"owner/repo\nmalicious", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.repo, func(t *testing.T) {
			err := validateRepo(tt.repo)
			if tt.isValid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.isValid && err == nil {
				t.Errorf("expected invalid, got nil error")
			}
		})
	}
}

func TestValidateRefPattern(t *testing.T) {
	tests := []struct {
		ref     string
		isValid bool
	}{
		{"main", true},
		{"feature/new-feature", true},
		{"v1.0.0", true},
		{"refs/heads/main", true},
		{"main; echo pwned", false},
		{"main\n", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			err := validateRef(tt.ref)
			if tt.isValid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.isValid && err == nil {
				t.Errorf("expected invalid, got nil error")
			}
		})
	}
}

// TestAskpassScriptDoesNotEmbedToken verifies that the GIT_ASKPASS script
// never contains the literal token value. A token with a single-quote
// character used to break out of the shell quoting and could enable arbitrary
// command execution. The fix passes the token through the GIT_TOKEN
// environment variable instead of embedding it in the script body.
func TestAskpassScriptDoesNotEmbedToken(t *testing.T) {
	dangerousTokens := []string{
		"ghp_normaltoken123",
		"tok'en;withquote",
		"tok'en'; echo pwned; echo '",
		"token`whoami`",
		"token$(rm -rf /)",
	}

	for _, token := range dangerousTokens {
		t.Run(token, func(t *testing.T) {
			script := buildAskpassScript()
			if strings.Contains(script, token) {
				t.Errorf("askpass script must not embed token literal, but found token in script:\n%s", script)
			}
			// The script must reference $GIT_TOKEN rather than the literal value.
			if !strings.Contains(script, "$GIT_TOKEN") {
				t.Errorf("askpass script must reference $GIT_TOKEN env var, got:\n%s", script)
			}
		})
	}
}

// TestAskpassScriptIsStaticShell verifies the generated script is a fixed,
// token-free shell script regardless of input. Any variation means the token
// is being interpolated into the script, which is the vulnerability.
func TestAskpassScriptIsStatic(t *testing.T) {
	script1 := buildAskpassScript()
	script2 := buildAskpassScript()
	if script1 != script2 {
		t.Errorf("buildAskpassScript must return a static string, got two different values:\n%q\n%q", script1, script2)
	}
	// Ensure the script header is correct shell.
	if !strings.HasPrefix(script1, "#!/bin/sh") {
		t.Errorf("askpass script must start with #!/bin/sh, got: %q", script1)
	}
}

func TestValidatePathPattern(t *testing.T) {
	tests := []struct {
		path    string
		isValid bool
	}{
		{"/tmp/passflow-abc12345", true},
		{"/tmp/passflow-deadbeef", true},
		{"/etc/passwd", false},
		{"/tmp/passflow-abc; rm -rf /", false},
		{"../../../etc/passwd", false},
		{"/tmp/other-dir", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := validatePath(tt.path)
			if tt.isValid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.isValid && err == nil {
				t.Errorf("expected invalid, got nil error")
			}
		})
	}
}
