package openai

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
)

// Client implements llm.Client using the OpenAI (or OpenAI-compatible) API.
type Client struct {
	client  *openai.Client
	modelID string
}

// New creates a new OpenAI-compatible client.
// baseURL can be overridden for Ollama or other compatible backends.
func New(apiKey, baseURL, modelID string) *Client {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" && baseURL != "https://api.openai.com/v1" {
		cfg.BaseURL = baseURL
	}
	return &Client{
		client:  openai.NewClientWithConfig(cfg),
		modelID: modelID,
	}
}

// Complete sends messages to OpenAI and returns the response.
func (c *Client) Complete(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition) (*llm.Response, error) {
	oaiMsgs := convertMessages(messages)

	req := openai.ChatCompletionRequest{
		Model:    c.modelID,
		Messages: oaiMsgs,
	}

	if len(tools) > 0 {
		req.Tools = convertTools(tools)
		req.ToolChoice = "auto"
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai: completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices returned")
	}

	choice := resp.Choices[0].Message
	result := &llm.Response{
		Content: choice.Content,
		Usage: llm.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	for _, tc := range choice.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return result, nil
}

// Close releases any resources held by the client.
// For OpenAI, this is a no-op as the HTTP client doesn't require cleanup.
func (c *Client) Close() error {
	return nil
}

// convertMessages converts llm.Message to OpenAI message format.
func convertMessages(messages []llm.Message) []openai.ChatCompletionMessage {
	oaiMsgs := make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, m := range messages {
		msg := openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
		if m.ToolCallID != "" {
			msg.ToolCallID = m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				msg.ToolCalls = append(msg.ToolCalls, openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				})
			}
		}
		oaiMsgs = append(oaiMsgs, msg)
	}
	return oaiMsgs
}

// convertTools converts llm.ToolDefinition to OpenAI tool format.
func convertTools(tools []llm.ToolDefinition) []openai.Tool {
	oaiTools := make([]openai.Tool, 0, len(tools))
	for _, t := range tools {
		params := t.Parameters
		if params == nil {
			params = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
		}
		oaiTools = append(oaiTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		})
	}
	return oaiTools
}
