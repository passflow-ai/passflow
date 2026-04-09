package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
)

const apiURL = "https://api.anthropic.com/v1/messages"
const apiVersion = "2023-06-01"

// Client implements llm.Client using the Anthropic Messages API.
type Client struct {
	apiKey     string
	modelID    string
	httpClient *http.Client
}

// New creates a new Anthropic client.
func New(apiKey, modelID string) *Client {
	return &Client{
		apiKey:     apiKey,
		modelID:    modelID,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

// Complete sends messages to Anthropic and returns the response.
func (c *Client) Complete(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition) (*llm.Response, error) {
	anthropicMsgs, system := convertMessages(messages)

	reqBody := anthropicRequest{
		Model:     c.modelID,
		MaxTokens: 4096,
		System:    system,
		Messages:  anthropicMsgs,
	}

	if len(tools) > 0 {
		reqBody.Tools = convertTools(tools)
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic: request failed: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic: API returned status %d: %s", resp.StatusCode, truncateBody(string(rawBody)))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(rawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("anthropic: failed to parse response: %w", err)
	}

	return parseResponse(&apiResp), nil
}

// Close releases any resources held by the client.
// For Anthropic, this is a no-op as the HTTP client doesn't require cleanup.
func (c *Client) Close() error {
	return nil
}

// convertMessages extracts system prompt and converts llm.Message to Anthropic format.
func convertMessages(messages []llm.Message) ([]anthropicMessage, string) {
	var system string
	var anthropicMsgs []anthropicMessage

	for _, m := range messages {
		switch m.Role {
		case "system":
			system = m.Content
		case "user":
			anthropicMsgs = append(anthropicMsgs, anthropicMessage{
				Role:    "user",
				Content: []anthropicContent{{Type: "text", Text: m.Content}},
			})
		case "assistant":
			if len(m.ToolCalls) > 0 {
				var content []anthropicContent
				if m.Content != "" {
					content = append(content, anthropicContent{Type: "text", Text: m.Content})
				}
				for _, tc := range m.ToolCalls {
					content = append(content, anthropicContent{
						Type:  "tool_use",
						ID:    tc.ID,
						Name:  tc.Name,
						Input: json.RawMessage(tc.Arguments),
					})
				}
				anthropicMsgs = append(anthropicMsgs, anthropicMessage{Role: "assistant", Content: content})
			} else {
				anthropicMsgs = append(anthropicMsgs, anthropicMessage{
					Role:    "assistant",
					Content: []anthropicContent{{Type: "text", Text: m.Content}},
				})
			}
		case "tool":
			// Tool results go as a user message with type=tool_result
			anthropicMsgs = append(anthropicMsgs, anthropicMessage{
				Role: "user",
				Content: []anthropicContent{{
					Type:      "tool_result",
					ToolUseID: m.ToolCallID,
					Content:   m.Content,
				}},
			})
		}
	}

	return anthropicMsgs, system
}

// convertTools converts llm.ToolDefinition to Anthropic tool format.
func convertTools(tools []llm.ToolDefinition) []anthropicTool {
	result := make([]anthropicTool, 0, len(tools))
	for _, t := range tools {
		result = append(result, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}
	return result
}

// parseResponse converts an Anthropic response to llm.Response.
func parseResponse(resp *anthropicResponse) *llm.Response {
	result := &llm.Response{
		Usage: llm.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			result.Content += block.Text
		case "tool_use":
			result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: string(block.Input),
			})
		}
	}

	return result
}

// truncateBody limits an API response body string to 200 characters so that
// large or sensitive error payloads are never written verbatim into log lines.
func truncateBody(body string) string {
	const maxLen = 200
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen]
}

// Anthropic API types

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type      string          `json:"type"` // text | tool_use | tool_result
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"` // for tool_result
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type anthropicResponse struct {
	ID      string             `json:"id"`
	Content []anthropicContent `json:"content"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	StopReason string `json:"stop_reason"`
}
