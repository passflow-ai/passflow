package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// allowedExecCommands is the allowlist of safe executable commands inside pods.
// Only the base command (first element) is checked.
// cat and env are intentionally excluded: cat can read arbitrary files
// (including service-account tokens) and env exposes environment variables
// that may contain API keys or other secrets.
var allowedExecCommands = map[string]bool{
	"ls":     true,
	"echo":   true,
	"whoami": true,
	"date":   true,
	"uname":  true,
	"df":     true,
	"free":   true,
	"top":    true,
	"ps":     true,
	"pwd":    true,
	"id":     true,
	"uptime": true,
	"which":  true,
}

// sensitivePaths is the set of filesystem path prefixes and exact names that
// must never appear as arguments to exec commands.
var sensitivePaths = []string{
	"/var/run/secrets/",
	"/proc/",
	"/etc/shadow",
	"/etc/passwd",
}

// validateExecArgs checks every argument in a command slice for sensitive
// path access or path-traversal sequences.
func validateExecArgs(command []string) error {
	for _, arg := range command[1:] { // skip the base command itself
		if strings.Contains(arg, "..") {
			return fmt.Errorf("exec: argument %q contains path traversal sequence (..)", arg)
		}
		for _, prefix := range sensitivePaths {
			if strings.HasPrefix(arg, prefix) || arg == strings.TrimSuffix(prefix, "/") {
				return fmt.Errorf("exec: argument %q accesses a sensitive path", arg)
			}
		}
	}
	return nil
}

// podNamespaceReader is a variable so tests can override it without touching
// the filesystem.
var podNamespaceReader = func() (string, error) {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", fmt.Errorf("could not determine pod namespace: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// validateNamespace ensures the requested namespace is within the operator's
// explicit allowlist (ALLOWED_NAMESPACES env var, comma-separated). When the
// variable is unset or empty the only permitted namespace is the pod's own
// namespace (read from the service-account projection).
func validateNamespace(ns string) error {
	allowed := os.Getenv("ALLOWED_NAMESPACES")
	if allowed != "" {
		for _, permitted := range strings.Split(allowed, ",") {
			if strings.TrimSpace(permitted) == ns {
				return nil
			}
		}
		return fmt.Errorf("namespace %q is not in ALLOWED_NAMESPACES", ns)
	}

	// No explicit allowlist — fall back to the pod's own namespace.
	selfNS, err := podNamespaceReader()
	if err != nil {
		return fmt.Errorf("namespace validation failed: %w", err)
	}
	if ns != selfNS {
		return fmt.Errorf("namespace %q is not allowed; only the pod's own namespace %q is permitted when ALLOWED_NAMESPACES is unset", ns, selfNS)
	}
	return nil
}

// K8sClient wraps kubernetes clientsets
type K8sClient struct {
	clientset  *kubernetes.Clientset
	dynamic    dynamic.Interface
	restConfig *rest.Config
}

// k8sClientFactory is a variable so tests can inject a fake constructor
// without rebuilding the binary.
var k8sClientFactory = NewK8sClient

// singleton state — protected by k8sClientOnce so that concurrent callers
// from the ExecuteAll goroutine pool can never double-initialise.
var (
	k8sClient     *K8sClient
	k8sClientOnce sync.Once
	k8sClientErr  error
)

// errTestSentinel is returned by the overridable factory in unit tests.
var errTestSentinel = errors.New("test sentinel: not in cluster")

// resetK8sSingleton resets the singleton for tests that need a clean slate.
// It must only be called from test code.
func resetK8sSingleton() {
	k8sClient = nil
	k8sClientErr = nil
	k8sClientOnce = sync.Once{}
}

// NewK8sClient creates a new kubernetes client using in-cluster config
func NewK8sClient() (*K8sClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &K8sClient{
		clientset:  clientset,
		dynamic:    dynamicClient,
		restConfig: config,
	}, nil
}

// GetK8sClient returns the singleton k8s client. Initialisation is guaranteed
// to run exactly once even when called concurrently from multiple goroutines.
func GetK8sClient() (*K8sClient, error) {
	k8sClientOnce.Do(func() {
		k8sClient, k8sClientErr = k8sClientFactory()
	})
	return k8sClient, k8sClientErr
}

func validateK8sAction(action string, args map[string]interface{}) error {
	switch action {
	case "get", "delete":
		resource, _ := args["resource"].(string)
		if resource == "" {
			return fmt.Errorf("resource is required")
		}
	case "apply":
		manifest, _ := args["manifest"].(string)
		if manifest == "" {
			return fmt.Errorf("manifest is required")
		}
	case "rollout_status", "rollout_undo":
		deployment, _ := args["deployment"].(string)
		if deployment == "" {
			return fmt.Errorf("deployment is required")
		}
	case "logs":
		pod, _ := args["pod"].(string)
		if pod == "" {
			return fmt.Errorf("pod is required")
		}
	case "exec":
		pod, _ := args["pod"].(string)
		command, _ := args["command"].([]interface{})
		if pod == "" {
			return fmt.Errorf("pod is required")
		}
		if len(command) == 0 {
			return fmt.Errorf("command is required")
		}
	case "diff":
		// diff can work with manifest or without
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}
	return nil
}

// runKubernetesAction executes a kubernetes action
func runKubernetesAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	if err := validateK8sAction(action, args); err != nil {
		return "", fmt.Errorf("kubernetes: %w", err)
	}

	// Validate namespace before making any API call so LLM-controlled input
	// cannot reach arbitrary namespaces.
	ns, _ := args["namespace"].(string)
	if ns == "" {
		ns = "default"
	}
	if err := validateNamespace(ns); err != nil {
		return "", fmt.Errorf("kubernetes: namespace validation: %w", err)
	}

	client, err := GetK8sClient()
	if err != nil {
		return "", fmt.Errorf("kubernetes: %w", err)
	}

	switch action {
	case "apply":
		return client.apply(ctx, args)
	case "get":
		return client.get(ctx, args)
	case "delete":
		return client.delete(ctx, args)
	case "diff":
		return client.diff(ctx, args)
	case "rollout_status":
		return client.rolloutStatus(ctx, args)
	case "rollout_undo":
		return client.rolloutUndo(ctx, args)
	case "logs":
		return client.logs(ctx, args)
	case "exec":
		return client.exec(ctx, args)
	default:
		return "", fmt.Errorf("kubernetes: unsupported action %q", action)
	}
}

func (c *K8sClient) apply(ctx context.Context, args map[string]interface{}) (string, error) {
	manifest, _ := args["manifest"].(string)
	namespace, _ := args["namespace"].(string)
	if namespace == "" {
		namespace = "default"
	}

	applied := []string{}
	decoder := yamlutil.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 4096)

	for {
		var obj unstructured.Unstructured
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to decode manifest: %w", err)
		}

		if obj.GetNamespace() == "" {
			obj.SetNamespace(namespace)
		}

		gvk := obj.GroupVersionKind()
		gvr := schema.GroupVersionResource{
			Group:    gvk.Group,
			Version:  gvk.Version,
			Resource: strings.ToLower(gvk.Kind) + "s",
		}

		_, err := c.dynamic.Resource(gvr).Namespace(obj.GetNamespace()).Apply(
			ctx,
			obj.GetName(),
			&obj,
			metav1.ApplyOptions{FieldManager: "passflow-gitops"},
		)
		if err != nil {
			return "", fmt.Errorf("failed to apply %s/%s: %w", gvk.Kind, obj.GetName(), err)
		}

		applied = append(applied, fmt.Sprintf("%s/%s", gvk.Kind, obj.GetName()))
	}

	result := map[string]interface{}{
		"applied": applied,
		"status":  "success",
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}

func (c *K8sClient) get(ctx context.Context, args map[string]interface{}) (string, error) {
	resource, _ := args["resource"].(string)
	namespace, _ := args["namespace"].(string)
	name, _ := args["name"].(string)

	if namespace == "" {
		namespace = "default"
	}

	var result interface{}

	switch resource {
	case "pods", "pod":
		if name != "" {
			pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			result = pod
		} else {
			pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				return "", err
			}
			result = pods.Items
		}
	case "deployments", "deployment":
		if name != "" {
			deploy, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			result = deploy
		} else {
			deploys, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				return "", err
			}
			result = deploys.Items
		}
	case "services", "service":
		if name != "" {
			svc, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			result = svc
		} else {
			svcs, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				return "", err
			}
			result = svcs.Items
		}
	case "configmaps", "configmap":
		if name != "" {
			cm, err := c.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			result = cm
		} else {
			cms, err := c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				return "", err
			}
			result = cms.Items
		}
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resource)
	}

	out, _ := json.Marshal(map[string]interface{}{"items": result})
	return string(out), nil
}

func (c *K8sClient) delete(ctx context.Context, args map[string]interface{}) (string, error) {
	resource, _ := args["resource"].(string)
	namespace, _ := args["namespace"].(string)
	name, _ := args["name"].(string)

	if namespace == "" {
		namespace = "default"
	}
	if name == "" {
		return "", fmt.Errorf("name is required for delete")
	}

	var err error
	switch resource {
	case "pods", "pod":
		err = c.clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "deployments", "deployment":
		err = c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "services", "service":
		err = c.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "configmaps", "configmap":
		err = c.clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resource)
	}

	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(map[string]interface{}{"deleted": true, "resource": resource, "name": name})
	return string(out), nil
}

func (c *K8sClient) diff(ctx context.Context, args map[string]interface{}) (string, error) {
	manifest, _ := args["manifest"].(string)
	namespace, _ := args["namespace"].(string)
	if namespace == "" {
		namespace = "default"
	}

	changes := []map[string]interface{}{}
	hasChanges := false

	decoder := yamlutil.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 4096)

	for {
		var obj unstructured.Unstructured
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to decode manifest: %w", err)
		}

		if obj.GetNamespace() == "" {
			obj.SetNamespace(namespace)
		}

		gvk := obj.GroupVersionKind()
		gvr := schema.GroupVersionResource{
			Group:    gvk.Group,
			Version:  gvk.Version,
			Resource: strings.ToLower(gvk.Kind) + "s",
		}

		existing, err := c.dynamic.Resource(gvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), metav1.GetOptions{})

		change := map[string]interface{}{
			"kind": gvk.Kind,
			"name": obj.GetName(),
		}

		if err != nil {
			change["action"] = "create"
			hasChanges = true
		} else {
			if existing.GetResourceVersion() != "" {
				change["action"] = "update"
				hasChanges = true
			} else {
				change["action"] = "unchanged"
			}
		}

		changes = append(changes, change)
	}

	result := map[string]interface{}{
		"changes":    changes,
		"hasChanges": hasChanges,
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}

// maxRolloutTimeoutSec is the upper bound on the user-supplied rollout_status
// timeout. Values above this cap (or zero/negative) are silently reduced to
// prevent a single LLM tool call from tying up a worker for an arbitrary
// amount of time.
const maxRolloutTimeoutSec = 120.0

func (c *K8sClient) rolloutStatus(ctx context.Context, args map[string]interface{}) (string, error) {
	deployment, _ := args["deployment"].(string)
	namespace, _ := args["namespace"].(string)
	timeoutSec, _ := args["timeout"].(float64)

	if namespace == "" {
		namespace = "default"
	}

	// Cap the timeout: reject zero/negative values and enforce the maximum.
	if timeoutSec <= 0 || timeoutSec > maxRolloutTimeoutSec {
		timeoutSec = maxRolloutTimeoutSec
	}

	timeout := time.Duration(timeoutSec) * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Respect context cancellation on every poll iteration so the worker
		// is not blocked past the job's deadline.
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("rollout_status cancelled: %w", ctx.Err())
		default:
		}

		deploy, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deployment, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		if deploy.Status.ReadyReplicas == *deploy.Spec.Replicas &&
			deploy.Status.UpdatedReplicas == *deploy.Spec.Replicas {
			result := map[string]interface{}{
				"ready":    true,
				"replicas": deploy.Status.ReadyReplicas,
			}
			out, _ := json.Marshal(result)
			return string(out), nil
		}

		time.Sleep(2 * time.Second)
	}

	return "", fmt.Errorf("timeout waiting for rollout")
}

func (c *K8sClient) rolloutUndo(ctx context.Context, args map[string]interface{}) (string, error) {
	deployment, _ := args["deployment"].(string)
	namespace, _ := args["namespace"].(string)

	if namespace == "" {
		namespace = "default"
	}

	// Fetch the current deployment to read its revision annotation and selector.
	deploy, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deployment, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get deployment %q: %w", deployment, err)
	}

	currentRevision := deploy.Annotations["deployment.kubernetes.io/revision"]

	// Build a label selector from the deployment's pod template labels so we
	// list only the ReplicaSets that belong to this deployment.
	selector := metav1.FormatLabelSelector(deploy.Spec.Selector)
	rsList, err := c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list replicasets for deployment %q: %w", deployment, err)
	}

	// Filter to ReplicaSets owned by this deployment and collect their
	// revision numbers. We want the one with the highest revision that is
	// strictly less than the current revision (i.e. the previous revision).
	type rsEntry struct {
		revision int
		name     string
		replicas int32
	}

	parseRevision := func(rev string) int {
		n := 0
		for _, ch := range rev {
			if ch >= '0' && ch <= '9' {
				n = n*10 + int(ch-'0')
			} else {
				return -1
			}
		}
		return n
	}

	currentRev := parseRevision(currentRevision)

	var candidates []rsEntry
	for _, rs := range rsList.Items {
		// Confirm ownership.
		owned := false
		for _, ref := range rs.OwnerReferences {
			if ref.Kind == "Deployment" && ref.Name == deployment {
				owned = true
				break
			}
		}
		if !owned {
			continue
		}

		rev := parseRevision(rs.Annotations["deployment.kubernetes.io/revision"])
		if rev < 0 {
			continue
		}
		// Only consider revisions that are older than the current one.
		if currentRev > 0 && rev >= currentRev {
			continue
		}

		replicas := int32(0)
		if rs.Spec.Replicas != nil {
			replicas = *rs.Spec.Replicas
		}
		candidates = append(candidates, rsEntry{revision: rev, name: rs.Name, replicas: replicas})
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no previous replicaset found for deployment %q to roll back to", deployment)
	}

	// Pick the candidate with the highest revision number (most recent prior revision).
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].revision > candidates[j].revision
	})
	previous := candidates[0]

	// Retrieve the full previous ReplicaSet to copy its pod template spec.
	prevRS, err := c.clientset.AppsV1().ReplicaSets(namespace).Get(ctx, previous.name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get previous replicaset %q: %w", previous.name, err)
	}

	// Patch the deployment's pod template spec with the one from the previous
	// ReplicaSet and update its revision annotation so Kubernetes treats this
	// as a new rollout to the old revision.
	deploy.Spec.Template.Spec = prevRS.Spec.Template.Spec
	deploy.Spec.Template.Labels = prevRS.Spec.Template.Labels
	if deploy.Annotations == nil {
		deploy.Annotations = make(map[string]string)
	}

	_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to update deployment for rollback: %w", err)
	}

	result := map[string]interface{}{
		"rolledBack":         true,
		"fromRevision":       currentRevision,
		"toRevision":         previous.revision,
		"previousReplicaSet": previous.name,
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}

func (c *K8sClient) logs(ctx context.Context, args map[string]interface{}) (string, error) {
	pod, _ := args["pod"].(string)
	namespace, _ := args["namespace"].(string)
	container, _ := args["container"].(string)
	tailLines, _ := args["tail"].(float64)

	if namespace == "" {
		namespace = "default"
	}
	if tailLines == 0 {
		tailLines = 100
	}

	tailLinesInt := int64(tailLines)
	opts := &corev1.PodLogOptions{
		TailLines: &tailLinesInt,
	}
	if container != "" {
		opts.Container = container
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(pod, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	// Cap the log read at 1 MB to prevent a runaway stream from exhausting
	// worker memory. Logs beyond this limit are silently dropped.
	const maxLogBytes = 1 << 20 // 1 MB
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, io.LimitReader(stream, maxLogBytes))
	if err != nil {
		return "", err
	}

	result := map[string]interface{}{
		"logs": buf.String(),
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}

func (c *K8sClient) exec(ctx context.Context, args map[string]interface{}) (string, error) {
	pod, _ := args["pod"].(string)
	namespace, _ := args["namespace"].(string)
	container, _ := args["container"].(string)
	commandRaw, _ := args["command"].([]interface{})

	if namespace == "" {
		namespace = "default"
	}

	command := make([]string, len(commandRaw))
	for i, c := range commandRaw {
		command[i], _ = c.(string)
	}

	if len(command) == 0 {
		return "", fmt.Errorf("exec: command is required")
	}
	baseCmd := command[0]
	if !allowedExecCommands[baseCmd] {
		allowed := make([]string, 0, len(allowedExecCommands))
		for cmd := range allowedExecCommands {
			allowed = append(allowed, cmd)
		}
		sort.Strings(allowed)
		return "", fmt.Errorf("exec: command %q is not allowed; permitted commands: %s", baseCmd, strings.Join(allowed, ", "))
	}

	// Validate arguments for sensitive path access even when the base command
	// itself is on the allowlist.
	if err := validateExecArgs(command); err != nil {
		return "", err
	}

	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.restConfig, "POST", req.URL())
	if err != nil {
		return "", err
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	exitCode := 0
	if err != nil {
		exitCode = 1
	}

	result := map[string]interface{}{
		"stdout":   stdout.String(),
		"stderr":   stderr.String(),
		"exitCode": exitCode,
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}
