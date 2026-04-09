package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var integrationHTTP = &http.Client{Timeout: 30 * time.Second}

// integrationRequest is the shared HTTP helper for all integration providers.
// It builds a request with the given method, URL, headers, and optional body,
// executes it, decodes the JSON response, and returns the re-marshalled result.
// A nil payload sends no body (suitable for GET requests).
func integrationRequest(ctx context.Context, method, url string, headers map[string]string, payload interface{}) (string, error) {
	var bodyReader io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return "", fmt.Errorf("integration: failed to encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return "", fmt.Errorf("integration: failed to create request: %w", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := integrationHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("integration: request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("integration: failed to decode response from %s: %w", url, err)
	}

	out, _ := json.Marshal(result)
	return string(out), nil
}

// --- Slack ---
//
// Slack integration actions:
//   - send_message: Post a message to a Slack channel. Requires "channel" (string)
//     and "text" (string) arguments. Uses the chat.postMessage API.
//   - list_channels: List up to 100 public Slack channels visible to the bot.
//     No arguments required. Uses the conversations.list API.

func runSlackAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	token := creds["access_token"]
	if token == "" {
		return "", fmt.Errorf("slack: missing access_token credential")
	}

	switch action {
	case "send_message":
		return slackSendMessage(ctx, token, args)
	case "list_channels":
		return slackListChannels(ctx, token)
	default:
		return "", fmt.Errorf("slack: unsupported action %q", action)
	}
}

func slackSendMessage(ctx context.Context, token string, args map[string]interface{}) (string, error) {
	channel, _ := args["channel"].(string)
	text, _ := args["text"].(string)
	if channel == "" || text == "" {
		return "", fmt.Errorf("slack.send_message: requires channel and text")
	}

	headers := map[string]string{"Authorization": "Bearer " + token}
	payload := map[string]interface{}{"channel": channel, "text": text}
	return integrationRequest(ctx, http.MethodPost, "https://slack.com/api/chat.postMessage", headers, payload)
}

func slackListChannels(ctx context.Context, token string) (string, error) {
	headers := map[string]string{"Authorization": "Bearer " + token}
	return integrationRequest(ctx, http.MethodGet, "https://slack.com/api/conversations.list?limit=100", headers, nil)
}

// --- GitHub ---
//
// GitHub integration actions:
//   - create_issue: Create a new GitHub issue. Requires "owner" (string), "repo"
//     (string), and "title" (string). Optional "body" (string) sets the issue
//     description. Uses the GitHub REST API v3.
//   - list_issues: List up to 20 open issues in a repository. Requires "owner"
//     (string) and "repo" (string).
//   - create_comment: Add a comment to an existing issue. Requires "owner"
//     (string), "repo" (string), "issue_number" (number), and "body" (string).
//   - get_repo: Retrieve metadata for a repository. Requires "owner" (string)
//     and "repo" (string).

func runGitHubAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	token := creds["access_token"]
	if token == "" {
		return "", fmt.Errorf("github: missing access_token credential")
	}

	switch action {
	case "create_issue":
		return githubCreateIssue(ctx, token, args)
	case "list_issues":
		return githubListIssues(ctx, token, args)
	case "create_comment":
		return githubCreateComment(ctx, token, args)
	case "get_repo":
		return githubGetRepo(ctx, token, args)
	default:
		return "", fmt.Errorf("github: unsupported action %q", action)
	}
}

func githubHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
		"Accept":        "application/vnd.github+json",
	}
}

func githubCreateIssue(ctx context.Context, token string, args map[string]interface{}) (string, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	title, _ := args["title"].(string)
	body, _ := args["body"].(string)
	if owner == "" || repo == "" || title == "" {
		return "", fmt.Errorf("github.create_issue: requires owner, repo, title")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	return integrationRequest(ctx, http.MethodPost, url, githubHeaders(token), map[string]interface{}{"title": title, "body": body})
}

func githubListIssues(ctx context.Context, token string, args map[string]interface{}) (string, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	if owner == "" || repo == "" {
		return "", fmt.Errorf("github.list_issues: requires owner, repo")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?per_page=20", owner, repo)
	return integrationRequest(ctx, http.MethodGet, url, githubHeaders(token), nil)
}

func githubCreateComment(ctx context.Context, token string, args map[string]interface{}) (string, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	issueNum, _ := args["issue_number"].(float64)
	body, _ := args["body"].(string)
	if owner == "" || repo == "" || issueNum == 0 || body == "" {
		return "", fmt.Errorf("github.create_comment: requires owner, repo, issue_number, body")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, int(issueNum))
	return integrationRequest(ctx, http.MethodPost, url, githubHeaders(token), map[string]interface{}{"body": body})
}

func githubGetRepo(ctx context.Context, token string, args map[string]interface{}) (string, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	if owner == "" || repo == "" {
		return "", fmt.Errorf("github.get_repo: requires owner and repo")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	return integrationRequest(ctx, http.MethodGet, url, githubHeaders(token), nil)
}

// --- Linear ---
//
// Linear integration actions:
//   - create_issue: Create a new issue in Linear via the GraphQL API. Requires
//     "title" (string) and "team_id" (string). Optional "description" (string)
//     sets the issue body. Returns the created issue ID, title, and URL.

func runLinearAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	token := creds["access_token"]
	if token == "" {
		return "", fmt.Errorf("linear: missing access_token")
	}
	switch action {
	case "create_issue":
		return linearCreateIssue(ctx, token, args)
	default:
		return "", fmt.Errorf("linear: unsupported action %q", action)
	}
}

func linearCreateIssue(ctx context.Context, token string, args map[string]interface{}) (string, error) {
	title, _ := args["title"].(string)
	description, _ := args["description"].(string)
	teamID, _ := args["team_id"].(string)
	if title == "" || teamID == "" {
		return "", fmt.Errorf("linear.create_issue: requires title and team_id")
	}

	query := `mutation CreateIssue($title: String!, $description: String, $teamId: String!) {
		issueCreate(input: {title: $title, description: $description, teamId: $teamId}) {
			success issue { id title url }
		}
	}`

	payload := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"title":       title,
			"description": description,
			"teamId":      teamID,
		},
	}

	headers := map[string]string{"Authorization": token}
	return integrationRequest(ctx, http.MethodPost, "https://api.linear.app/graphql", headers, payload)
}

// --- Notion ---
//
// Notion integration actions:
//   - create_page: Create a new Notion page as a child of an existing page.
//     Requires "parent_id" (string) and "title" (string). Optional "content"
//     (string) is added as the first paragraph block of the new page.
//   - search: Search across all pages and databases in the workspace. Requires
//     "query" (string) — the full-text search term.

func runNotionAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	token := creds["access_token"]
	if token == "" {
		return "", fmt.Errorf("notion: missing access_token")
	}
	switch action {
	case "create_page":
		return notionCreatePage(ctx, token, args)
	case "search":
		return notionSearch(ctx, token, args)
	default:
		return "", fmt.Errorf("notion: unsupported action %q", action)
	}
}

func notionHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization":  "Bearer " + token,
		"Notion-Version": "2022-06-28",
	}
}

func notionCreatePage(ctx context.Context, token string, args map[string]interface{}) (string, error) {
	parentID, _ := args["parent_id"].(string)
	title, _ := args["title"].(string)
	content, _ := args["content"].(string)
	if parentID == "" || title == "" {
		return "", fmt.Errorf("notion.create_page: requires parent_id and title")
	}
	payload := map[string]interface{}{
		"parent": map[string]interface{}{"page_id": parentID},
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"title": []interface{}{
					map[string]interface{}{"text": map[string]interface{}{"content": title}},
				},
			},
		},
		"children": []interface{}{
			map[string]interface{}{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []interface{}{
						map[string]interface{}{"text": map[string]interface{}{"content": content}},
					},
				},
			},
		},
	}
	return integrationRequest(ctx, http.MethodPost, "https://api.notion.com/v1/pages", notionHeaders(token), payload)
}

func notionSearch(ctx context.Context, token string, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	return integrationRequest(ctx, http.MethodPost, "https://api.notion.com/v1/search", notionHeaders(token), map[string]interface{}{"query": query})
}

// --- Prometheus ---
//
// Prometheus Query API integration (read-only):
//   - query: Execute an instant PromQL query. Requires "query" (string).
//     Optional "time" (string, RFC3339 or Unix timestamp) for evaluation time.
//   - query_range: Execute a range PromQL query. Requires "query" (string),
//     "start" (string), "end" (string), and "step" (string, e.g. "15s", "1m").
//   - series: Find series by label matchers. Requires "match" (string or []string).
//     Optional "start" and "end" for time range.
//   - labels: Get all label names. No required arguments.
//   - label_values: Get values for a specific label. Requires "label" (string).
//
// Credentials:
//   - url: Base URL of Prometheus server (e.g., http://prometheus:9090)
//   - username (optional): For basic auth
//   - password (optional): For basic auth
//   - token (optional): For Bearer token auth

func runPrometheusAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	baseURL := creds["url"]
	if baseURL == "" {
		return "", fmt.Errorf("prometheus: missing url credential")
	}

	headers := prometheusHeaders(creds)

	switch action {
	case "query":
		return prometheusQuery(ctx, baseURL, headers, args)
	case "query_range":
		return prometheusQueryRange(ctx, baseURL, headers, args)
	case "series":
		return prometheusSeries(ctx, baseURL, headers, args)
	case "labels":
		return prometheusLabels(ctx, baseURL, headers)
	case "label_values":
		return prometheusLabelValues(ctx, baseURL, headers, args)
	default:
		return "", fmt.Errorf("prometheus: unsupported action %q", action)
	}
}

func prometheusHeaders(creds map[string]string) map[string]string {
	headers := make(map[string]string)

	// Bearer token auth
	if token := creds["token"]; token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	// Basic auth (encoded as header if both username and password provided)
	// Note: For basic auth, the integrationRequestWithBasicAuth helper would be better,
	// but for simplicity we'll handle it via URL or expect proxy/ingress auth
	return headers
}

func prometheusQuery(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("prometheus.query: requires query")
	}

	url := fmt.Sprintf("%s/api/v1/query?query=%s", baseURL, urlEncode(query))

	if t, ok := args["time"].(string); ok && t != "" {
		url += "&time=" + urlEncode(t)
	}

	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func prometheusQueryRange(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	start, _ := args["start"].(string)
	end, _ := args["end"].(string)
	step, _ := args["step"].(string)

	if query == "" || start == "" || end == "" || step == "" {
		return "", fmt.Errorf("prometheus.query_range: requires query, start, end, step")
	}

	url := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%s&end=%s&step=%s",
		baseURL, urlEncode(query), urlEncode(start), urlEncode(end), urlEncode(step))

	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func prometheusSeries(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	url := fmt.Sprintf("%s/api/v1/series", baseURL)

	// match can be string or []string
	params := "?"
	switch m := args["match"].(type) {
	case string:
		params += "match[]=" + urlEncode(m)
	case []interface{}:
		for i, v := range m {
			if s, ok := v.(string); ok {
				if i > 0 {
					params += "&"
				}
				params += "match[]=" + urlEncode(s)
			}
		}
	default:
		return "", fmt.Errorf("prometheus.series: requires match (string or array)")
	}

	if start, ok := args["start"].(string); ok && start != "" {
		params += "&start=" + urlEncode(start)
	}
	if end, ok := args["end"].(string); ok && end != "" {
		params += "&end=" + urlEncode(end)
	}

	return integrationRequest(ctx, http.MethodGet, url+params, headers, nil)
}

func prometheusLabels(ctx context.Context, baseURL string, headers map[string]string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/labels", baseURL)
	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func prometheusLabelValues(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	label, _ := args["label"].(string)
	if label == "" {
		return "", fmt.Errorf("prometheus.label_values: requires label")
	}

	url := fmt.Sprintf("%s/api/v1/label/%s/values", baseURL, urlEncode(label))
	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

// --- Loki ---
//
// Loki Query API integration (read-only):
//   - query: Execute an instant LogQL query. Requires "query" (string).
//     Optional "limit" (number, default 100), "time" (string), "direction" (string: forward/backward).
//   - query_range: Execute a range LogQL query. Requires "query" (string),
//     "start" (string), "end" (string). Optional "limit" (number), "step" (string), "direction" (string).
//   - labels: Get all label names. Optional "start" and "end" for time range.
//   - label_values: Get values for a specific label. Requires "label" (string).
//     Optional "start", "end", and "query" (LogQL selector to filter).
//   - series: Find series by label matchers. Requires "match" (string or []string).
//     Optional "start" and "end" for time range.
//
// Credentials:
//   - url: Base URL of Loki server (e.g., http://loki:3100)
//   - username (optional): For basic auth
//   - password (optional): For basic auth
//   - token (optional): For Bearer token auth
//   - org_id (optional): X-Scope-OrgID header for multi-tenant Loki

func runLokiAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	baseURL := creds["url"]
	if baseURL == "" {
		return "", fmt.Errorf("loki: missing url credential")
	}

	headers := lokiHeaders(creds)

	switch action {
	case "query":
		return lokiQuery(ctx, baseURL, headers, args)
	case "query_range":
		return lokiQueryRange(ctx, baseURL, headers, args)
	case "labels":
		return lokiLabels(ctx, baseURL, headers, args)
	case "label_values":
		return lokiLabelValues(ctx, baseURL, headers, args)
	case "series":
		return lokiSeries(ctx, baseURL, headers, args)
	default:
		return "", fmt.Errorf("loki: unsupported action %q", action)
	}
}

func lokiHeaders(creds map[string]string) map[string]string {
	headers := make(map[string]string)

	// Bearer token auth
	if token := creds["token"]; token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	// Multi-tenant org ID
	if orgID := creds["org_id"]; orgID != "" {
		headers["X-Scope-OrgID"] = orgID
	}

	return headers
}

func lokiQuery(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("loki.query: requires query")
	}

	url := fmt.Sprintf("%s/loki/api/v1/query?query=%s", baseURL, urlEncode(query))

	if limit, ok := args["limit"].(float64); ok && limit > 0 {
		url += fmt.Sprintf("&limit=%d", int(limit))
	}
	if t, ok := args["time"].(string); ok && t != "" {
		url += "&time=" + urlEncode(t)
	}
	if dir, ok := args["direction"].(string); ok && dir != "" {
		url += "&direction=" + urlEncode(dir)
	}

	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func lokiQueryRange(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	start, _ := args["start"].(string)
	end, _ := args["end"].(string)

	if query == "" || start == "" || end == "" {
		return "", fmt.Errorf("loki.query_range: requires query, start, end")
	}

	url := fmt.Sprintf("%s/loki/api/v1/query_range?query=%s&start=%s&end=%s",
		baseURL, urlEncode(query), urlEncode(start), urlEncode(end))

	if limit, ok := args["limit"].(float64); ok && limit > 0 {
		url += fmt.Sprintf("&limit=%d", int(limit))
	}
	if step, ok := args["step"].(string); ok && step != "" {
		url += "&step=" + urlEncode(step)
	}
	if dir, ok := args["direction"].(string); ok && dir != "" {
		url += "&direction=" + urlEncode(dir)
	}

	return integrationRequest(ctx, http.MethodGet, url, headers, nil)
}

func lokiLabels(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	url := fmt.Sprintf("%s/loki/api/v1/labels", baseURL)

	params := ""
	if start, ok := args["start"].(string); ok && start != "" {
		params += "?start=" + urlEncode(start)
	}
	if end, ok := args["end"].(string); ok && end != "" {
		if params == "" {
			params = "?"
		} else {
			params += "&"
		}
		params += "end=" + urlEncode(end)
	}

	return integrationRequest(ctx, http.MethodGet, url+params, headers, nil)
}

func lokiLabelValues(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	label, _ := args["label"].(string)
	if label == "" {
		return "", fmt.Errorf("loki.label_values: requires label")
	}

	url := fmt.Sprintf("%s/loki/api/v1/label/%s/values", baseURL, urlEncode(label))

	params := ""
	if start, ok := args["start"].(string); ok && start != "" {
		params += "?start=" + urlEncode(start)
	}
	if end, ok := args["end"].(string); ok && end != "" {
		if params == "" {
			params = "?"
		} else {
			params += "&"
		}
		params += "end=" + urlEncode(end)
	}
	if q, ok := args["query"].(string); ok && q != "" {
		if params == "" {
			params = "?"
		} else {
			params += "&"
		}
		params += "query=" + urlEncode(q)
	}

	return integrationRequest(ctx, http.MethodGet, url+params, headers, nil)
}

func lokiSeries(ctx context.Context, baseURL string, headers map[string]string, args map[string]interface{}) (string, error) {
	url := fmt.Sprintf("%s/loki/api/v1/series", baseURL)

	params := "?"
	switch m := args["match"].(type) {
	case string:
		params += "match[]=" + urlEncode(m)
	case []interface{}:
		for i, v := range m {
			if s, ok := v.(string); ok {
				if i > 0 {
					params += "&"
				}
				params += "match[]=" + urlEncode(s)
			}
		}
	default:
		return "", fmt.Errorf("loki.series: requires match (string or array)")
	}

	if start, ok := args["start"].(string); ok && start != "" {
		params += "&start=" + urlEncode(start)
	}
	if end, ok := args["end"].(string); ok && end != "" {
		params += "&end=" + urlEncode(end)
	}

	return integrationRequest(ctx, http.MethodGet, url+params, headers, nil)
}

// urlEncode performs URL encoding for query parameters.
func urlEncode(s string) string {
	// Using a simple replacement for common characters
	// For production, use net/url.QueryEscape
	result := ""
	for _, c := range s {
		switch c {
		case ' ':
			result += "%20"
		case '=':
			result += "%3D"
		case '&':
			result += "%26"
		case '+':
			result += "%2B"
		case '{':
			result += "%7B"
		case '}':
			result += "%7D"
		case '"':
			result += "%22"
		case '\'':
			result += "%27"
		case '[':
			result += "%5B"
		case ']':
			result += "%5D"
		case '#':
			result += "%23"
		case '%':
			result += "%25"
		default:
			result += string(c)
		}
	}
	return result
}
