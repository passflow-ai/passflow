package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// --- OCI Metrics ---
//
// OCI Monitoring API integration for querying OCI metrics:
//   - query: Execute an MQL (Monitoring Query Language) query. Requires
//     "compartment_id" (string) and "query" (string). Optional "start_time"
//     and "end_time" (RFC3339 strings), defaults to last hour.
//   - list_metrics: List available metric definitions. Requires "compartment_id".
//     Optional "namespace" (string) to filter by metric namespace.
//   - summarize: Summarize metrics data. Requires "compartment_id", "namespace",
//     "query", "start_time", "end_time". Optional "resolution" (string, e.g. "1m").
//
// Credentials (OCI API signing):
//   - tenancy_ocid: OCID of the tenancy
//   - user_ocid: OCID of the user
//   - fingerprint: API key fingerprint
//   - private_key: Private key PEM content
//   - region: OCI region (e.g., sa-saopaulo-1)
//
// For simplicity, this implementation uses HTTP with pre-signed requests.
// In production, use the OCI Go SDK for proper request signing.

func runOCIMetricsAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	region := creds["region"]
	if region == "" {
		return "", fmt.Errorf("oci_metrics: missing region credential")
	}

	baseURL := fmt.Sprintf("https://telemetry.%s.oraclecloud.com", region)
	headers := ociHeaders(creds)

	switch action {
	case "query":
		return ociMetricsQuery(ctx, baseURL, headers, args)
	case "list_metrics":
		return ociListMetrics(ctx, baseURL, headers, args)
	case "summarize":
		return ociSummarizeMetrics(ctx, baseURL, headers, args)
	default:
		return "", fmt.Errorf("oci_metrics: unsupported action %q", action)
	}
}

func ociHeaders(creds map[string]string) map[string]string {
	// Note: In production, these headers should be properly signed using OCI request signing.
	// This simplified version expects a pre-authenticated token or uses instance principal.
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	// If using instance principal or resource principal, token is auto-obtained
	if token := creds["token"]; token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	return headers
}

func ociMetricsQuery(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	compartmentID, _ := args["compartment_id"].(string)
	query, _ := args["query"].(string)

	if compartmentID == "" || query == "" {
		return "", fmt.Errorf("oci_metrics.query: requires compartment_id, query")
	}

	// Default time range: last hour
	endTime := time.Now().UTC()
	startTime := endTime.Add(-1 * time.Hour)

	if st, ok := args["start_time"].(string); ok && st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			startTime = t
		}
	}
	if et, ok := args["end_time"].(string); ok && et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			endTime = t
		}
	}

	payload := map[string]interface{}{
		"compartmentId": compartmentID,
		"query":         query,
		"startTime":     startTime.Format(time.RFC3339),
		"endTime":       endTime.Format(time.RFC3339),
	}

	url := fmt.Sprintf("%s/20180401/metrics/actions/summarizeMetricsData", baseURL)
	return integrationRequest(ctx, http.MethodPost, url, headers, payload)
}

func ociListMetrics(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	compartmentID, _ := args["compartment_id"].(string)
	if compartmentID == "" {
		return "", fmt.Errorf("oci_metrics.list_metrics: requires compartment_id")
	}

	payload := map[string]interface{}{
		"compartmentId": compartmentID,
	}

	if ns, ok := args["namespace"].(string); ok && ns != "" {
		payload["namespace"] = ns
	}

	url := fmt.Sprintf("%s/20180401/metrics/actions/listMetrics", baseURL)
	return integrationRequest(ctx, http.MethodPost, url, headers, payload)
}

func ociSummarizeMetrics(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	compartmentID, _ := args["compartment_id"].(string)
	namespace, _ := args["namespace"].(string)
	query, _ := args["query"].(string)
	startTimeStr, _ := args["start_time"].(string)
	endTimeStr, _ := args["end_time"].(string)

	if compartmentID == "" || namespace == "" || query == "" || startTimeStr == "" || endTimeStr == "" {
		return "", fmt.Errorf("oci_metrics.summarize: requires compartment_id, namespace, query, start_time, end_time")
	}

	payload := map[string]interface{}{
		"compartmentId": compartmentID,
		"namespace":     namespace,
		"query":         query,
		"startTime":     startTimeStr,
		"endTime":       endTimeStr,
	}

	if res, ok := args["resolution"].(string); ok && res != "" {
		payload["resolution"] = res
	}

	url := fmt.Sprintf("%s/20180401/metrics/actions/summarizeMetricsData", baseURL)
	return integrationRequest(ctx, http.MethodPost, url, headers, payload)
}

// --- OCI Compute ---
//
// OCI Compute API integration for node autoscaling:
//   - list_instance_pools: List instance pools in a compartment. Requires "compartment_id".
//   - get_instance_pool: Get details of an instance pool. Requires "instance_pool_id".
//   - update_pool_size: Update the size of an instance pool. Requires "instance_pool_id"
//     and "size" (number).
//   - list_instances: List instances in a pool. Requires "compartment_id" and
//     "instance_pool_id".
//   - terminate_instance: Terminate an instance. Requires "instance_id".
//     Optional "preserve_boot_volume" (bool, default false).
//
// Credentials: Same as OCI Metrics

func runOCIComputeAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	region := creds["region"]
	if region == "" {
		return "", fmt.Errorf("oci_compute: missing region credential")
	}

	baseURL := fmt.Sprintf("https://iaas.%s.oraclecloud.com", region)
	headers := ociHeaders(creds)

	switch action {
	case "list_instance_pools":
		return ociListInstancePools(ctx, baseURL, headers, args)
	case "get_instance_pool":
		return ociGetInstancePool(ctx, baseURL, headers, args)
	case "update_pool_size":
		return ociUpdatePoolSize(ctx, baseURL, headers, args)
	case "list_instances":
		return ociListInstances(ctx, baseURL, headers, args)
	case "terminate_instance":
		return ociTerminateInstance(ctx, baseURL, headers, args)
	default:
		return "", fmt.Errorf("oci_compute: unsupported action %q", action)
	}
}

func ociListInstancePools(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	compartmentID, _ := args["compartment_id"].(string)
	if compartmentID == "" {
		return "", fmt.Errorf("oci_compute.list_instance_pools: requires compartment_id")
	}

	url := fmt.Sprintf("%s/20160918/instancePools?compartmentId=%s", baseURL, urlEncode(compartmentID))
	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func ociGetInstancePool(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	poolID, _ := args["instance_pool_id"].(string)
	if poolID == "" {
		return "", fmt.Errorf("oci_compute.get_instance_pool: requires instance_pool_id")
	}

	url := fmt.Sprintf("%s/20160918/instancePools/%s", baseURL, urlEncode(poolID))
	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func ociUpdatePoolSize(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	poolID, _ := args["instance_pool_id"].(string)
	size, ok := args["size"].(float64)

	if poolID == "" || !ok {
		return "", fmt.Errorf("oci_compute.update_pool_size: requires instance_pool_id, size")
	}

	payload := map[string]interface{}{
		"size": int(size),
	}

	url := fmt.Sprintf("%s/20160918/instancePools/%s", baseURL, urlEncode(poolID))
	return integrationRequest(ctx, http.MethodPut, url, headers, payload)
}

func ociListInstances(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	compartmentID, _ := args["compartment_id"].(string)
	poolID, _ := args["instance_pool_id"].(string)

	if compartmentID == "" {
		return "", fmt.Errorf("oci_compute.list_instances: requires compartment_id")
	}

	url := fmt.Sprintf("%s/20160918/instances?compartmentId=%s", baseURL, urlEncode(compartmentID))

	if poolID != "" {
		// Filter by pool - need to get pool instances separately
		url = fmt.Sprintf("%s/20160918/instancePools/%s/instances?compartmentId=%s",
			baseURL, urlEncode(poolID), urlEncode(compartmentID))
	}

	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func ociTerminateInstance(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	instanceID, _ := args["instance_id"].(string)
	if instanceID == "" {
		return "", fmt.Errorf("oci_compute.terminate_instance: requires instance_id")
	}

	preserveBootVolume := false
	if pbv, ok := args["preserve_boot_volume"].(bool); ok {
		preserveBootVolume = pbv
	}

	url := fmt.Sprintf("%s/20160918/instances/%s?preserveBootVolume=%t",
		baseURL, urlEncode(instanceID), preserveBootVolume)
	return integrationRequest(ctx, http.MethodDelete, url, headers, nil)
}

// --- OCI Registry (Artifacts) ---
//
// OCI Container Registry API integration for image management:
//   - list_repositories: List container repositories. Requires "compartment_id".
//   - list_images: List images in a repository. Requires "compartment_id" and
//     "repository_id". Optional "sort_by" (TIMECREATED, DISPLAYNAME),
//     "sort_order" (ASC, DESC).
//   - get_image: Get image details. Requires "image_id".
//   - delete_image: Delete a container image. Requires "image_id".
//
// Credentials: Same as OCI Metrics

func runOCIRegistryAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	region := creds["region"]
	if region == "" {
		return "", fmt.Errorf("oci_registry: missing region credential")
	}

	baseURL := fmt.Sprintf("https://artifacts.%s.oraclecloud.com", region)
	headers := ociHeaders(creds)

	switch action {
	case "list_repositories":
		return ociListRepositories(ctx, baseURL, headers, args)
	case "list_images":
		return ociListImages(ctx, baseURL, headers, args)
	case "get_image":
		return ociGetImage(ctx, baseURL, headers, args)
	case "delete_image":
		return ociDeleteImage(ctx, baseURL, headers, args)
	default:
		return "", fmt.Errorf("oci_registry: unsupported action %q", action)
	}
}

func ociListRepositories(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	compartmentID, _ := args["compartment_id"].(string)
	if compartmentID == "" {
		return "", fmt.Errorf("oci_registry.list_repositories: requires compartment_id")
	}

	url := fmt.Sprintf("%s/20160918/container/repositories?compartmentId=%s", baseURL, urlEncode(compartmentID))
	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func ociListImages(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	compartmentID, _ := args["compartment_id"].(string)
	repositoryID, _ := args["repository_id"].(string)

	if compartmentID == "" || repositoryID == "" {
		return "", fmt.Errorf("oci_registry.list_images: requires compartment_id, repository_id")
	}

	url := fmt.Sprintf("%s/20160918/container/images?compartmentId=%s&repositoryId=%s",
		baseURL, urlEncode(compartmentID), urlEncode(repositoryID))

	if sortBy, ok := args["sort_by"].(string); ok && sortBy != "" {
		url += "&sortBy=" + urlEncode(strings.ToUpper(sortBy))
	}
	if sortOrder, ok := args["sort_order"].(string); ok && sortOrder != "" {
		url += "&sortOrder=" + urlEncode(strings.ToUpper(sortOrder))
	}

	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func ociGetImage(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	imageID, _ := args["image_id"].(string)
	if imageID == "" {
		return "", fmt.Errorf("oci_registry.get_image: requires image_id")
	}

	url := fmt.Sprintf("%s/20160918/container/images/%s", baseURL, urlEncode(imageID))
	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func ociDeleteImage(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	imageID, _ := args["image_id"].(string)
	if imageID == "" {
		return "", fmt.Errorf("oci_registry.delete_image: requires image_id")
	}

	url := fmt.Sprintf("%s/20160918/container/images/%s", baseURL, urlEncode(imageID))
	return integrationRequest(ctx, http.MethodDelete, url, headers, nil)
}

// --- OCI Jaeger/Tempo (APM Trace) ---
//
// OCI Application Performance Monitoring - Trace API:
//   - query_traces: Query traces. Requires "compartment_id", "apm_domain_id".
//     Optional "time_span_started_greater_than_or_equal_to", "time_span_started_less_than",
//     "query" (AQL query string).
//   - get_trace: Get trace details. Requires "apm_domain_id", "trace_key".
//   - get_span: Get span details. Requires "apm_domain_id", "trace_key", "span_key".
//
// Credentials: Same as OCI Metrics

func runOCITraceAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	region := creds["region"]
	if region == "" {
		return "", fmt.Errorf("oci_trace: missing region credential")
	}

	baseURL := fmt.Sprintf("https://apm-trace.%s.oraclecloud.com", region)
	headers := ociHeaders(creds)

	switch action {
	case "query_traces":
		return ociQueryTraces(ctx, baseURL, headers, args)
	case "get_trace":
		return ociGetTrace(ctx, baseURL, headers, args)
	case "get_span":
		return ociGetSpan(ctx, baseURL, headers, args)
	default:
		return "", fmt.Errorf("oci_trace: unsupported action %q", action)
	}
}

func ociQueryTraces(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	apmDomainID, _ := args["apm_domain_id"].(string)
	if apmDomainID == "" {
		return "", fmt.Errorf("oci_trace.query_traces: requires apm_domain_id")
	}

	// Default time range: last hour
	endTime := time.Now().UTC()
	startTime := endTime.Add(-1 * time.Hour)

	if st, ok := args["start_time"].(string); ok && st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			startTime = t
		}
	}
	if et, ok := args["end_time"].(string); ok && et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			endTime = t
		}
	}

	url := fmt.Sprintf("%s/20200630/queries/results?apmDomainId=%s&timeSpanStartedGreaterThanOrEqualTo=%s&timeSpanStartedLessThan=%s",
		baseURL, urlEncode(apmDomainID), urlEncode(startTime.Format(time.RFC3339)), urlEncode(endTime.Format(time.RFC3339)))

	if query, ok := args["query"].(string); ok && query != "" {
		url += "&query=" + urlEncode(query)
	}

	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func ociGetTrace(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	apmDomainID, _ := args["apm_domain_id"].(string)
	traceKey, _ := args["trace_key"].(string)

	if apmDomainID == "" || traceKey == "" {
		return "", fmt.Errorf("oci_trace.get_trace: requires apm_domain_id, trace_key")
	}

	url := fmt.Sprintf("%s/20200630/traces/%s?apmDomainId=%s",
		baseURL, urlEncode(traceKey), urlEncode(apmDomainID))
	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func ociGetSpan(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	apmDomainID, _ := args["apm_domain_id"].(string)
	traceKey, _ := args["trace_key"].(string)
	spanKey, _ := args["span_key"].(string)

	if apmDomainID == "" || traceKey == "" || spanKey == "" {
		return "", fmt.Errorf("oci_trace.get_span: requires apm_domain_id, trace_key, span_key")
	}

	url := fmt.Sprintf("%s/20200630/traces/%s/spans/%s?apmDomainId=%s",
		baseURL, urlEncode(traceKey), urlEncode(spanKey), urlEncode(apmDomainID))
	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

// ProviderDispatcher routes OCI actions to the appropriate handler
func RunOCIAction(ctx context.Context, provider, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	switch provider {
	case "oci_metrics":
		return runOCIMetricsAction(ctx, action, creds, args)
	case "oci_compute":
		return runOCIComputeAction(ctx, action, creds, args)
	case "oci_registry":
		return runOCIRegistryAction(ctx, action, creds, args)
	case "oci_trace":
		return runOCITraceAction(ctx, action, creds, args)
	default:
		return "", fmt.Errorf("unknown OCI provider: %s", provider)
	}
}

// OCICredentials holds the required credentials for OCI API calls
type OCICredentials struct {
	TenancyOCID string `json:"tenancy_ocid"`
	UserOCID    string `json:"user_ocid"`
	Fingerprint string `json:"fingerprint"`
	PrivateKey  string `json:"private_key"`
	Region      string `json:"region"`
	Token       string `json:"token,omitempty"` // For instance/resource principal
}

// ValidateOCICredentials checks that required OCI credentials are present
func ValidateOCICredentials(creds map[string]string) error {
	// If using instance principal, only region is required
	if creds["token"] != "" {
		if creds["region"] == "" {
			return fmt.Errorf("oci: missing region credential")
		}
		return nil
	}

	// For API key authentication, all fields are required
	required := []string{"tenancy_ocid", "user_ocid", "fingerprint", "private_key", "region"}
	var missing []string

	for _, field := range required {
		if creds[field] == "" {
			missing = append(missing, field)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("oci: missing credentials: %s", strings.Join(missing, ", "))
	}

	return nil
}

// OCIMetricQuery represents a metric query request
type OCIMetricQuery struct {
	CompartmentID string `json:"compartmentId"`
	Namespace     string `json:"namespace,omitempty"`
	Query         string `json:"query"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	Resolution    string `json:"resolution,omitempty"`
}

// ToJSON converts the query to JSON
func (q *OCIMetricQuery) ToJSON() ([]byte, error) {
	return json.Marshal(q)
}
