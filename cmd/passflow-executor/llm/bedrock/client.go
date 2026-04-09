package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"

	"github.com/jaak-ai/passflow-agent-executor/llm"
)

// Client implements llm.Client using AWS Bedrock.
type Client struct {
	client  *bedrockruntime.Client
	modelID string
}

// New creates a new Bedrock client with explicit credentials.
func New(ctx context.Context, region, accessKey, secretKey, modelID string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("bedrock: failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)
	return &Client{
		client:  client,
		modelID: modelID,
	}, nil
}

// NewWithConfig creates a Bedrock client from an existing AWS config.
func NewWithConfig(cfg aws.Config, modelID string) *Client {
	return &Client{
		client:  bedrockruntime.NewFromConfig(cfg),
		modelID: modelID,
	}
}

// Complete sends messages to Bedrock and returns the response.
func (c *Client) Complete(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition) (*llm.Response, error) {
	if isClaudeModel(c.modelID) {
		return c.completeClaude(ctx, messages, tools)
	}
	return nil, fmt.Errorf("bedrock: unsupported model %q, only Claude models are currently supported", c.modelID)
}

// completeClaude handles completion for Anthropic Claude models on Bedrock.
func (c *Client) completeClaude(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition) (*llm.Response, error) {
	request := claudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		Messages:         convertMessagesForClaude(messages),
	}

	if systemPrompt := extractSystemPrompt(messages); systemPrompt != "" {
		request.System = systemPrompt
	}

	if len(tools) > 0 {
		request.Tools = convertToolsForClaude(tools)
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("bedrock: failed to marshal request: %w", err)
	}

	output, err := c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.modelID),
		ContentType: aws.String("application/json"),
		Body:        payload,
	})
	if err != nil {
		return nil, fmt.Errorf("bedrock: InvokeModel failed: %w", err)
	}

	var response claudeResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("bedrock: failed to unmarshal response: %w", err)
	}

	return parseClaudeResponse(&response), nil
}

// isClaudeModel checks if the model ID is an Anthropic Claude model.
func isClaudeModel(modelID string) bool {
	return strings.HasPrefix(modelID, "anthropic.claude")
}

// extractSystemPrompt finds and returns the system message content.
func extractSystemPrompt(messages []llm.Message) string {
	for _, m := range messages {
		if m.Role == "system" {
			return m.Content
		}
	}
	return ""
}

// convertMessagesForClaude converts llm.Message to Claude message format.
func convertMessagesForClaude(messages []llm.Message) []claudeMessage {
	var result []claudeMessage
	for _, m := range messages {
		if m.Role == "system" {
			continue // System is handled separately
		}

		if m.Role == "tool" {
			toolResult := convertToolResultForClaude(m)
			result = append(result, toolResult)
			continue
		}

		msg := claudeMessage{
			Role: m.Role,
		}

		// Add text content
		if m.Content != "" {
			msg.Content = append(msg.Content, claudeContent{
				Type: "text",
				Text: m.Content,
			})
		}

		// Add tool use for assistant messages with tool calls
		for _, tc := range m.ToolCalls {
			msg.Content = append(msg.Content, claudeContent{
				Type: "tool_use",
				ToolUse: &claudeToolUse{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: json.RawMessage(tc.Arguments),
				},
			})
		}

		result = append(result, msg)
	}
	return result
}

// convertToolResultForClaude converts a tool result message to Claude format.
func convertToolResultForClaude(m llm.Message) claudeMessage {
	return claudeMessage{
		Role: "user",
		Content: []claudeContent{
			{
				Type:      "tool_result",
				ToolUseID: m.ToolCallID,
				Content:   m.Content,
			},
		},
	}
}

// convertToolsForClaude converts llm.ToolDefinition to Claude tool format.
func convertToolsForClaude(tools []llm.ToolDefinition) []claudeTool {
	result := make([]claudeTool, 0, len(tools))
	for _, t := range tools {
		result = append(result, claudeTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}
	return result
}

// convertToolCall converts a claudeToolUse to llm.ToolCall.
func convertToolCall(tu claudeToolUse) llm.ToolCall {
	return llm.ToolCall{
		ID:        tu.ID,
		Name:      tu.Name,
		Arguments: string(tu.Input),
	}
}

// parseClaudeResponse converts a Claude response to llm.Response.
func parseClaudeResponse(resp *claudeResponse) *llm.Response {
	result := &llm.Response{
		Usage: llm.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}

	for _, content := range resp.Content {
		switch content.Type {
		case "text":
			result.Content += content.Text
		case "tool_use":
			if content.ToolUse != nil {
				result.ToolCalls = append(result.ToolCalls, convertToolCall(*content.ToolUse))
			}
		}
	}

	return result
}
