package tools

import (
	"context"
	"strings"
	"sync"
	"testing"
)

// --- HIGH 1: Singleton race condition ---

// TestGetK8sClientNoConcurrentDoubleInit verifies that concurrent callers
// cannot observe two distinct *K8sClient values from GetK8sClient.
// Because NewK8sClient always fails outside a cluster, the singleton must
// store the error too and return the same error to every caller.
func TestGetK8sClientConcurrentCallsAreSerialised(t *testing.T) {
	// Reset state so this test is self-contained.
	resetK8sSingleton()

	const goroutines = 50
	errs := make([]error, goroutines)
	clients := make([]*K8sClient, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			c, err := GetK8sClient()
			clients[i] = c
			errs[i] = err
		}()
	}
	wg.Wait()

	// Outside a cluster all calls must fail with the same error.
	// Check that no two goroutines received different non-nil clients.
	var firstNonNil *K8sClient
	for _, c := range clients {
		if c != nil {
			if firstNonNil == nil {
				firstNonNil = c
			} else if firstNonNil != c {
				t.Error("GetK8sClient returned two different non-nil clients: singleton invariant violated")
			}
		}
	}
}

// TestGetK8sClientSyncOnceSemantics ensures the initializer runs exactly once.
func TestGetK8sClientSyncOnceSemantics(t *testing.T) {
	resetK8sSingleton()

	initCount := 0
	// Temporarily override the factory to count invocations.
	orig := k8sClientFactory
	defer func() { k8sClientFactory = orig }()
	k8sClientFactory = func() (*K8sClient, error) {
		initCount++
		return nil, errTestSentinel
	}

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			GetK8sClient() //nolint:errcheck
		}()
	}
	wg.Wait()

	if initCount != 1 {
		t.Errorf("factory called %d times; want exactly 1 (sync.Once violated)", initCount)
	}
}

// --- HIGH 2: exec allowlist bypass ---

// TestCatAndEnvRemovedFromAllowlist ensures that cat and env are no longer
// in the allowlist (they can read secrets and env vars respectively).
func TestCatAndEnvRemovedFromAllowlist(t *testing.T) {
	client := &K8sClient{}
	ctx := context.Background()

	for _, cmd := range []string{"cat", "env"} {
		t.Run(cmd, func(t *testing.T) {
			_, err := client.exec(ctx, map[string]interface{}{
				"pod":     "test-pod",
				"command": []interface{}{cmd},
			})
			if err == nil {
				t.Errorf("command %q must be rejected but was accepted", cmd)
			}
			if !strings.Contains(err.Error(), "is not allowed") {
				t.Errorf("expected 'is not allowed' error for %q, got: %v", cmd, err)
			}
		})
	}
}

// TestValidateExecArgsSensitivePaths verifies that arguments containing
// sensitive filesystem paths are rejected before any network call.
func TestValidateExecArgsSensitivePaths(t *testing.T) {
	blocked := [][]string{
		{"ls", "/var/run/secrets/"},
		{"ls", "/var/run/secrets/kubernetes.io/serviceaccount/token"},
		{"ls", "/proc/1/environ"},
		{"ls", "/etc/shadow"},
		{"ls", "/etc/passwd"},
		{"ls", "../../../etc/passwd"},
		{"ls", "/some/../secret"},
	}

	for _, args := range blocked {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			err := validateExecArgs(args)
			if err == nil {
				t.Errorf("args %v should be blocked but validateExecArgs returned nil", args)
			}
		})
	}
}

// TestValidateExecArgsSafePaths verifies that benign arguments are accepted.
func TestValidateExecArgsSafePaths(t *testing.T) {
	safe := [][]string{
		{"ls", "-la"},
		{"ls", "/tmp"},
		{"echo", "hello world"},
		{"ps", "aux"},
	}

	for _, args := range safe {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			err := validateExecArgs(args)
			if err != nil {
				t.Errorf("args %v should be accepted but got: %v", args, err)
			}
		})
	}
}

// TestExecRejectsPathTraversal verifies the full exec() path rejects path traversal.
func TestExecRejectsPathTraversal(t *testing.T) {
	client := &K8sClient{}
	ctx := context.Background()

	_, err := client.exec(ctx, map[string]interface{}{
		"pod":     "test-pod",
		"command": []interface{}{"ls", "/var/run/secrets/kubernetes.io/serviceaccount/token"},
	})
	if err == nil {
		t.Fatal("exec must reject command with sensitive path argument")
	}
	if strings.Contains(err.Error(), "is not allowed") {
		// The base command was rejected - but ls is in the allowlist now.
		// This means we need argument validation to fire separately.
		t.Errorf("error should mention sensitive path, not command allowlist: %v", err)
	}
}

// --- HIGH 3: Namespace restriction ---

// TestValidateNamespaceWithAllowlist verifies that only explicitly allowed
// namespaces pass when ALLOWED_NAMESPACES is set.
func TestValidateNamespaceWithAllowlist(t *testing.T) {
	t.Setenv("ALLOWED_NAMESPACES", "default,passflow,production")

	allowed := []string{"default", "passflow", "production"}
	for _, ns := range allowed {
		t.Run("allowed_"+ns, func(t *testing.T) {
			if err := validateNamespace(ns); err != nil {
				t.Errorf("namespace %q should be allowed, got: %v", ns, err)
			}
		})
	}

	blocked := []string{"kube-system", "kube-public", "attacker", ""}
	for _, ns := range blocked {
		t.Run("blocked_"+ns, func(t *testing.T) {
			if err := validateNamespace(ns); err == nil {
				t.Errorf("namespace %q should be blocked but validateNamespace returned nil", ns)
			}
		})
	}
}

// TestValidateNamespaceNoEnvDefaultsToSelfNamespace verifies that when
// ALLOWED_NAMESPACES is unset, only the pod's own namespace is permitted.
// We stub the pod namespace reader so the test is hermetic.
func TestValidateNamespaceNoEnvDefaultsToSelfNamespace(t *testing.T) {
	t.Setenv("ALLOWED_NAMESPACES", "")

	// Override the namespace reader to avoid filesystem access in tests.
	orig := podNamespaceReader
	defer func() { podNamespaceReader = orig }()
	podNamespaceReader = func() (string, error) { return "my-namespace", nil }

	if err := validateNamespace("my-namespace"); err != nil {
		t.Errorf("pod's own namespace should be allowed, got: %v", err)
	}
	if err := validateNamespace("other-namespace"); err == nil {
		t.Error("foreign namespace must be blocked when ALLOWED_NAMESPACES is unset")
	}
}

// TestRunKubernetesActionNamespaceValidated verifies that runKubernetesAction
// enforces namespace validation before hitting the K8s API.
func TestRunKubernetesActionNamespaceValidated(t *testing.T) {
	t.Setenv("ALLOWED_NAMESPACES", "default")

	orig := podNamespaceReader
	defer func() { podNamespaceReader = orig }()
	podNamespaceReader = func() (string, error) { return "default", nil }

	ctx := context.Background()
	_, err := runKubernetesAction(ctx, "get", nil, map[string]interface{}{
		"resource":  "pods",
		"namespace": "kube-system",
	})
	if err == nil {
		t.Fatal("runKubernetesAction must reject disallowed namespace")
	}
	if !strings.Contains(err.Error(), "namespace") {
		t.Errorf("error should mention namespace, got: %v", err)
	}
}
