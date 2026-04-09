package mcp

import (
	"context"
	"fmt"
	"sync"
)

// Executor bridges MCP tools with the agent execution system.
type Executor struct {
	client     *Client
	tools      map[string]Tool
	toolsMu    sync.RWMutex
	toolsReady bool
}

// NewExecutor creates a new MCP executor.
func NewExecutor(client *Client) *Executor {
	return &Executor{
		client: client,
		tools:  make(map[string]Tool),
	}
}

// ListTools returns all available tools from the MCP gateway.
func (e *Executor) ListTools(ctx context.Context) ([]Tool, error) {
	tools, err := e.client.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}

	e.toolsMu.Lock()
	e.tools = make(map[string]Tool)
	for _, t := range tools {
		e.tools[t.Name] = t
	}
	e.toolsReady = true
	e.toolsMu.Unlock()

	return tools, nil
}

// CanHandle checks if a tool is available via MCP.
func (e *Executor) CanHandle(ctx context.Context, toolName string) (bool, error) {
	e.toolsMu.RLock()
	ready := e.toolsReady
	e.toolsMu.RUnlock()

	if !ready {
		if _, err := e.ListTools(ctx); err != nil {
			return false, fmt.Errorf("load tools: %w", err)
		}
	}

	e.toolsMu.RLock()
	_, ok := e.tools[toolName]
	e.toolsMu.RUnlock()

	return ok, nil
}

// Execute invokes a tool via MCP and returns the result.
func (e *Executor) Execute(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	resp, err := e.client.CallTool(ctx, toolName, arguments)
	if err != nil {
		return "", fmt.Errorf("call tool %s: %w", toolName, err)
	}

	if resp.IsError {
		if len(resp.Content) > 0 {
			return "", fmt.Errorf("tool error: %s", resp.Content[0].Text)
		}
		return "", fmt.Errorf("tool error: unknown")
	}

	if len(resp.Content) == 0 {
		return "", nil
	}

	return resp.Content[0].Text, nil
}
