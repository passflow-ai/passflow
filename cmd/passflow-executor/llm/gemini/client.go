package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
)

// Client implements llm.Client using Google Gemini.
type Client struct {
	client  *genai.Client
	modelID string
}

// New creates a new Gemini client.
func New(ctx context.Context, apiKey, modelID string) (*Client, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to create client: %w", err)
	}
	return &Client{
		client:  client,
		modelID: modelID,
	}, nil
}

// Close releases the underlying gRPC connection held by the genai.Client.
// It should be called when the Client is no longer needed, typically during
// application shutdown. Close is safe to call on a zero-value Client.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

// Complete sends messages to Gemini and returns the response.
func (c *Client) Complete(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition) (*llm.Response, error) {
	model := c.client.GenerativeModel(c.modelID)

	// Extract system instruction
	for _, m := range messages {
		if m.Role == "system" {
			model.SystemInstruction = &genai.Content{
				Parts: []genai.Part{genai.Text(m.Content)},
			}
			break
		}
	}

	// Configure tools if provided
	if len(tools) > 0 {
		model.Tools = convertTools(tools)
	}

	// Build contents from messages
	contents := convertMessages(messages)
	if len(contents) == 0 {
		return nil, fmt.Errorf("gemini: no messages to send")
	}

	// Start chat session
	chat := model.StartChat()
	if len(contents) > 1 {
		chat.History = contents[:len(contents)-1]
	}

	// Send the last message
	lastContent := contents[len(contents)-1]
	resp, err := chat.SendMessage(ctx, lastContent.Parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini: completion failed: %w", err)
	}

	return convertResponse(resp), nil
}

func convertMessages(msgs []llm.Message) []*genai.Content {
	var contents []*genai.Content
	for _, m := range msgs {
		if m.Role == "system" {
			continue // Handled separately as SystemInstruction
		}

		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}

		content := &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(m.Content)},
		}

		// Handle tool calls from assistant
		for _, tc := range m.ToolCalls {
			args := make(map[string]any)
			json.Unmarshal([]byte(tc.Arguments), &args)
			content.Parts = append(content.Parts, genai.FunctionCall{
				Name: tc.Name,
				Args: args,
			})
		}

		// Handle tool results
		if m.Role == "tool" {
			content = &genai.Content{
				Role: "user",
				Parts: []genai.Part{
					genai.FunctionResponse{
						Name:     m.ToolCallID,
						Response: map[string]any{"result": m.Content},
					},
				},
			}
		}

		contents = append(contents, content)
	}
	return contents
}

func convertTools(tools []llm.ToolDefinition) []*genai.Tool {
	var geminiTools []*genai.Tool
	var declarations []*genai.FunctionDeclaration

	for _, t := range tools {
		declarations = append(declarations, &genai.FunctionDeclaration{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  convertSchema(t.Parameters),
		})
	}

	if len(declarations) > 0 {
		geminiTools = append(geminiTools, &genai.Tool{
			FunctionDeclarations: declarations,
		})
	}
	return geminiTools
}

func convertSchema(params map[string]interface{}) *genai.Schema {
	if params == nil {
		return &genai.Schema{Type: genai.TypeObject}
	}

	schema := &genai.Schema{
		Type: genai.TypeObject,
	}

	if props, ok := params["properties"].(map[string]interface{}); ok {
		schema.Properties = make(map[string]*genai.Schema)
		for name, prop := range props {
			if propMap, ok := prop.(map[string]interface{}); ok {
				schema.Properties[name] = &genai.Schema{
					Type:        getSchemaType(propMap["type"]),
					Description: getString(propMap, "description"),
				}
			}
		}
	}

	if required, ok := params["required"].([]interface{}); ok {
		for _, r := range required {
			if s, ok := r.(string); ok {
				schema.Required = append(schema.Required, s)
			}
		}
	}

	return schema
}

func getSchemaType(t interface{}) genai.Type {
	switch t {
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	default:
		return genai.TypeObject
	}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func convertResponse(resp *genai.GenerateContentResponse) *llm.Response {
	result := &llm.Response{}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			switch p := part.(type) {
			case genai.Text:
				result.Content += string(p)
			case genai.FunctionCall:
				args, _ := json.Marshal(p.Args)
				result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
					ID:        p.Name,
					Name:      p.Name,
					Arguments: string(args),
				})
			}
		}
	}

	if resp.UsageMetadata != nil {
		result.Usage = llm.Usage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	return result
}
