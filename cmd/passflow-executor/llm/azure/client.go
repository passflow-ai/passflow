package azure

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
)

// Client implements llm.Client using Azure OpenAI.
type Client struct {
	client     *openai.Client
	deployment string
}

// New creates a new Azure OpenAI client.
func New(endpoint, apiKey, deployment string) (*Client, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("azure: endpoint is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("azure: API key is required")
	}
	if deployment == "" {
		return nil, fmt.Errorf("azure: deployment name is required")
	}

	cfg := openai.DefaultAzureConfig(apiKey, endpoint)
	// Set the deployment name as the model mapper
	cfg.AzureModelMapperFunc = func(model string) string {
		return deployment
	}

	client := openai.NewClientWithConfig(cfg)
	return &Client{
		client:     client,
		deployment: deployment,
	}, nil
}

// Complete sends messages to Azure OpenAI and returns the response.
func (c *Client) Complete(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition) (*llm.Response, error) {
	oaiMsgs := convertMessages(messages)

	req := openai.ChatCompletionRequest{
		Model:    c.deployment,
		Messages: oaiMsgs,
	}

	if len(tools) > 0 {
		req.Tools = convertTools(tools)
		req.ToolChoice = "auto"
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("azure: completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("azure: no choices returned")
	}

	choice := azureChatChoice{
		Message: azureChatMessage{
			Content:   &resp.Choices[0].Message.Content,
			ToolCalls: convertToolCallsFromOpenAI(resp.Choices[0].Message.ToolCalls),
		},
	}
	usage := azureUsage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}

	return parseResponse(choice, usage), nil
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

		name := t.Name
		description := t.Description
		oaiTools = append(oaiTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        name,
				Description: description,
				Parameters:  params,
			},
		})
	}
	return oaiTools
}

// convertToolCallsFromOpenAI converts OpenAI tool calls to our internal format.
func convertToolCallsFromOpenAI(oaiCalls []openai.ToolCall) []azureToolCall {
	calls := make([]azureToolCall, 0, len(oaiCalls))
	for _, tc := range oaiCalls {
		calls = append(calls, azureToolCall{
			ID:   tc.ID,
			Type: string(tc.Type),
			Function: azureFunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}
	return calls
}

// serializeParameters converts parameters map to JSON bytes.
func serializeParameters(params map[string]interface{}) []byte {
	if params == nil {
		return []byte(`{"type":"object","properties":{}}`)
	}
	data, err := json.Marshal(params)
	if err != nil {
		return []byte(`{"type":"object","properties":{}}`)
	}
	return data
}

// parseResponse converts Azure response to llm.Response.
func parseResponse(choice azureChatChoice, usage azureUsage) *llm.Response {
	result := &llm.Response{
		Usage: llm.Usage{
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			TotalTokens:      usage.TotalTokens,
		},
	}

	if choice.Message.Content != nil {
		result.Content = *choice.Message.Content
	}

	for _, tc := range choice.Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return result
}

// Internal types for response parsing

type azureChatChoice struct {
	Message azureChatMessage
}

type azureChatMessage struct {
	Role       string
	Content    *string
	ToolCalls  []azureToolCall
	ToolCallID string
}

type azureToolCall struct {
	ID       string
	Type     string
	Function azureFunctionCall
}

type azureFunctionCall struct {
	Name      string
	Arguments string
}

type azureUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
