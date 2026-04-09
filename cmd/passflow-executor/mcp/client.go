package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Client communicates with the MCP Gateway.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new MCP client. The bearer token used to authenticate
// against the MCP Gateway is read from the MCP_GATEWAY_TOKEN environment
// variable. Requests will be sent without authentication when the variable is
// unset, which allows unit tests to operate against stub servers.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   os.Getenv("MCP_GATEWAY_TOKEN"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// newRequest builds an *http.Request with the provided context and attaches
// the Authorization header when a token is configured. The context is
// propagated so callers can cancel or time-out in-flight MCP requests.
func (c *Client) newRequest(ctx context.Context, method, url string, body *bytes.Reader) (*http.Request, error) {
	var bodyReader interface {
		Read(p []byte) (n int, err error)
	}
	if body != nil {
		bodyReader = body
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

// ListTools fetches available tools from the gateway.
// The context is forwarded to the underlying HTTP request so the caller can
// cancel or impose a deadline on the network call.
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.baseURL+"/tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("build list tools request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list tools request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list tools returned status %d", resp.StatusCode)
	}

	var result ListToolsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode list tools response: %w", err)
	}

	return result.Tools, nil
}

// CallTool invokes a tool via the gateway.
// The context is forwarded to the underlying HTTP request so the caller can
// cancel or impose a deadline on the network call.
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*CallResponse, error) {
	callReq := CallRequest{
		Name:      name,
		Arguments: arguments,
	}

	body, err := json.Marshal(callReq)
	if err != nil {
		return nil, fmt.Errorf("marshal call request: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, c.baseURL+"/tools/call", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build call tool request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call tool request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("call tool returned status %d", resp.StatusCode)
	}

	var result CallResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode call tool response: %w", err)
	}

	return &result, nil
}
