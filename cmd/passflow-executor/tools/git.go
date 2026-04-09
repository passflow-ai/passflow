package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Validation patterns for git inputs
var (
	validRepoPattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`)
	validRefPattern  = regexp.MustCompile(`^[a-zA-Z0-9_./-]+$`)
	validPathPattern = regexp.MustCompile(`^/tmp/passflow-[a-f0-9]{8}$`)

	// tokenInURLPattern matches authentication tokens embedded in URLs of the
	// form https://anything:TOKEN@host or https://TOKEN@host and replaces the
	// sensitive credential portion so it is never surfaced in error messages.
	tokenInURLPattern = regexp.MustCompile(`(https?://)[^@\s]+@`)
)

// sanitizeGitError removes any embedded authentication tokens from a git
// error string before it is returned to callers.  Git occasionally echoes the
// remote URL (including credentials) in its stderr output; this function
// replaces the user-info portion with a placeholder so tokens are not leaked.
func sanitizeGitError(msg string) string {
	return tokenInURLPattern.ReplaceAllString(msg, "${1}<redacted>@")
}

func validateRepo(repo string) error {
	if repo == "" {
		return fmt.Errorf("repo cannot be empty")
	}
	if !validRepoPattern.MatchString(repo) {
		return fmt.Errorf("invalid repo format: %q", repo)
	}
	return nil
}

func validateRef(ref string) error {
	if ref == "" {
		return fmt.Errorf("ref cannot be empty")
	}
	if strings.ContainsAny(ref, ";\n\r`$") {
		return fmt.Errorf("ref contains invalid characters: %q", ref)
	}
	if !validRefPattern.MatchString(ref) {
		return fmt.Errorf("invalid ref format: %q", ref)
	}
	return nil
}

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("path cannot contain '..': %q", path)
	}
	if !validPathPattern.MatchString(path) {
		return fmt.Errorf("invalid path format: %q (must be /tmp/passflow-<hex8>)", path)
	}
	return nil
}

// buildAskpassScript returns a fixed, token-free shell script that reads the
// access token from the GIT_TOKEN environment variable and prints it to
// stdout. Git invokes this script as GIT_ASKPASS to obtain the password.
//
// The token is intentionally never embedded in the script body: embedding it
// with shell quoting (e.g. echo 'TOKEN') allows injection when the token
// contains single-quote characters or other shell metacharacters. By
// referencing $GIT_TOKEN instead, the script is a static string regardless of
// the token value.
func buildAskpassScript() string {
	return "#!/bin/sh\necho \"$GIT_TOKEN\"\n"
}

// CreateWorkspace creates a temporary workspace directory
func CreateWorkspace() string {
	id := uuid.New().String()[:8]
	return fmt.Sprintf("/tmp/passflow-%s", id)
}

// runGitAction executes a git action
func runGitAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	switch action {
	case "clone":
		return gitClone(ctx, creds, args)
	case "pull":
		return gitPull(ctx, args)
	case "diff":
		return gitDiff(ctx, args)
	case "checkout":
		return gitCheckout(ctx, args)
	default:
		return "", fmt.Errorf("git: unsupported action %q", action)
	}
}

func gitClone(ctx context.Context, creds map[string]string, args map[string]interface{}) (string, error) {
	repo, _ := args["repo"].(string)
	branch, _ := args["branch"].(string)
	path, _ := args["path"].(string)

	if err := validateRepo(repo); err != nil {
		return "", fmt.Errorf("git.clone: %w", err)
	}

	if branch == "" {
		branch = "main"
	}
	if err := validateRef(branch); err != nil {
		return "", fmt.Errorf("git.clone: %w", err)
	}

	if path == "" {
		path = CreateWorkspace()
	}
	if err := validatePath(path); err != nil {
		return "", fmt.Errorf("git.clone: %w", err)
	}

	// The clone URL never contains the token.  Credentials are supplied via
	// GIT_ASKPASS so they stay out of process arguments, logs, and git's own
	// error messages.
	cloneURL := fmt.Sprintf("https://github.com/%s.git", repo)

	cmd := exec.CommandContext(ctx, "git", "clone", "--branch", branch, "--single-branch", "--depth", "1", cloneURL, path)

	// Provide credentials through GIT_ASKPASS when a token is available.
	// GIT_ASKPASS is an executable that git invokes to obtain the password;
	// using it keeps the token out of argv and out of git's stderr output.
	token := creds["access_token"]
	if token != "" {
		// Write a minimal askpass script to a temporary file.  The script
		// ignores its argument (the prompt) and simply prints the token, which
		// git uses as the password for the configured username.
		askpassFile, err := os.CreateTemp("", "git-askpass-*.sh")
		if err != nil {
			return "", fmt.Errorf("git.clone: failed to create askpass script: %w", err)
		}
		askpassPath := askpassFile.Name()
		// Clean up the script once the command has finished.
		defer os.Remove(askpassPath)

		// Script outputs the token as the password via the GIT_TOKEN env var.
		// The token is never embedded in the script body — this prevents shell
		// injection when the token contains single-quotes or other shell
		// metacharacters.
		scriptContent := buildAskpassScript()
		if _, err := askpassFile.WriteString(scriptContent); err != nil {
			askpassFile.Close()
			return "", fmt.Errorf("git.clone: failed to write askpass script: %w", err)
		}
		askpassFile.Close()

		if err := os.Chmod(askpassPath, 0700); err != nil {
			return "", fmt.Errorf("git.clone: failed to chmod askpass script: %w", err)
		}

		cmd.Env = append(os.Environ(),
			"GIT_ASKPASS="+askpassPath,
			"GIT_USERNAME=x-access-token",
			// Pass the token through an environment variable so it is never
			// part of the script source and cannot be injected via shell quoting.
			"GIT_TOKEN="+token,
			// Disable the system credential helper so git does not attempt to
			// store or retrieve credentials from any external keychain.
			"GIT_TERMINAL_PROMPT=0",
		)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git.clone failed: %s", sanitizeGitError(stderr.String()))
	}

	// Get commit hash
	commit, err := getHeadCommit(ctx, path)
	if err != nil {
		commit = "unknown"
	}

	result := map[string]interface{}{
		"path":   path,
		"commit": commit,
		"branch": branch,
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}

func gitPull(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)

	if err := validatePath(path); err != nil {
		return "", fmt.Errorf("git.pull: %w", err)
	}

	// Check directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("git.pull: path does not exist: %s", path)
	}

	oldCommit, _ := getHeadCommit(ctx, path)

	cmd := exec.CommandContext(ctx, "git", "-C", path, "pull", "--ff-only")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git.pull failed: %s", sanitizeGitError(stderr.String()))
	}

	newCommit, _ := getHeadCommit(ctx, path)

	result := map[string]interface{}{
		"updated": oldCommit != newCommit,
		"commit":  newCommit,
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}

func gitDiff(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	base, _ := args["base"].(string)

	if err := validatePath(path); err != nil {
		return "", fmt.Errorf("git.diff: %w", err)
	}

	if base == "" {
		base = "HEAD~1"
	}
	if err := validateRef(base); err != nil {
		return "", fmt.Errorf("git.diff: %w", err)
	}

	// Get changed files
	cmdFiles := exec.CommandContext(ctx, "git", "-C", path, "diff", "--name-only", base)
	filesOut, err := cmdFiles.Output()
	if err != nil {
		return "", fmt.Errorf("git.diff: failed to get changed files")
	}

	files := strings.Split(strings.TrimSpace(string(filesOut)), "\n")
	if len(files) == 1 && files[0] == "" {
		files = []string{}
	}

	// Get diff content (limited)
	cmdDiff := exec.CommandContext(ctx, "git", "-C", path, "diff", base)
	diffOut, _ := cmdDiff.Output()
	diffStr := string(diffOut)
	if len(diffStr) > 10000 {
		diffStr = diffStr[:10000] + "\n... (truncated)"
	}

	result := map[string]interface{}{
		"files": files,
		"diff":  diffStr,
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}

func gitCheckout(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	ref, _ := args["ref"].(string)

	if err := validatePath(path); err != nil {
		return "", fmt.Errorf("git.checkout: %w", err)
	}
	if err := validateRef(ref); err != nil {
		return "", fmt.Errorf("git.checkout: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "-C", path, "checkout", ref)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git.checkout failed: %s", sanitizeGitError(stderr.String()))
	}

	commit, _ := getHeadCommit(ctx, path)

	result := map[string]interface{}{
		"ref":    ref,
		"commit": commit,
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}

func getHeadCommit(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
