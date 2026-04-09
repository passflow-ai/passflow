package acf

import (
	"fmt"
	"strings"
)

const defaultMaxLength = 100_000

// CADConfig is the lightweight config passed in a job spec, carrying the
// agent definition that the PromptTranslator needs at execution time.
type CADConfig struct {
	SystemPrompt string            `json:"system_prompt"`
	Skills       []SkillDef        `json:"skills,omitempty"`
	ContextFiles map[string]string `json:"context_files,omitempty"` // filename → content
	SourceType   string            `json:"source_type"`
}

// SkillDef represents a named block of skill instructions attached to an agent.
type SkillDef struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ToolDef describes a single tool available to the agent during execution.
type ToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // system, mcp, http
}

// PromptTranslator converts a CADConfig + tool list into a system prompt
// that is optimised for the target LLM provider.
type PromptTranslator struct {
	// MaxLength is the character budget for the generated prompt.
	// When <= 0, defaultMaxLength (100 000 chars) is used.
	MaxLength int
}

// NewPromptTranslator returns a PromptTranslator with the default max length.
func NewPromptTranslator() *PromptTranslator {
	return &PromptTranslator{MaxLength: defaultMaxLength}
}

// Translate builds the final system prompt for the given provider.
// It returns an empty string when cad is nil.
func (pt *PromptTranslator) Translate(provider string, cad *CADConfig, tools []ToolDef) string {
	if cad == nil {
		return ""
	}

	var b strings.Builder

	// 1. Base system prompt from the agent definition.
	b.WriteString(cad.SystemPrompt)

	// 2. Context files (e.g. CLAUDE.md content).
	if len(cad.ContextFiles) > 0 {
		b.WriteString("\n\n## Context Files\n")
		for name, content := range cad.ContextFiles {
			b.WriteString(fmt.Sprintf("\n### %s\n\n%s\n", name, content))
		}
	}

	// 3. Active skills.
	if len(cad.Skills) > 0 {
		b.WriteString("\n\n## Skills\n")
		for _, skill := range cad.Skills {
			b.WriteString(fmt.Sprintf("\n### %s\n\n%s\n", skill.Name, skill.Content))
		}
	}

	// 4. Provider-specific tool usage instructions.
	if len(tools) > 0 {
		b.WriteString(pt.buildToolInstructions(provider, tools))
	}

	prompt := b.String()

	// 5. Enforce max length — truncate with a trailing notice.
	maxLen := pt.MaxLength
	if maxLen <= 0 {
		maxLen = defaultMaxLength
	}
	if len(prompt) > maxLen {
		const truncationNotice = "\n\n[NOTICE: system prompt was truncated to fit the provider's context limit]"
		cutAt := maxLen - len(truncationNotice)
		if cutAt < 0 {
			cutAt = 0
		}
		prompt = prompt[:cutAt] + truncationNotice
	}

	return prompt
}

// buildToolInstructions returns the provider-specific tool usage section.
func (pt *PromptTranslator) buildToolInstructions(provider string, tools []ToolDef) string {
	var b strings.Builder

	switch provider {
	case "anthropic":
		// Anthropic natively understands tool_use blocks; a brief enumeration suffices.
		b.WriteString("\n\n## Available Tools\n")
		b.WriteString("Use the `tool_use` content block to invoke any of the following tools:\n\n")
		for _, t := range tools {
			b.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", t.Name, t.Type, t.Description))
		}

	case "openai", "azure":
		// OpenAI / Azure OpenAI expect strict JSON function calls.
		b.WriteString("\n\n## Tool Calling Instructions (OpenAI Function Calling)\n")
		b.WriteString("You MUST call tools using the `function_call` format with strict JSON arguments.\n")
		b.WriteString("Do NOT describe tool usage in prose — always emit a well-formed function call.\n")
		b.WriteString("Argument values must conform exactly to the function's parameter schema.\n\n")
		b.WriteString("Available functions:\n\n")
		for _, t := range tools {
			b.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", t.Name, t.Type, t.Description))
		}

	case "gemini":
		// Gemini uses functionDeclarations; remind the model to use functionCall responses.
		b.WriteString("\n\n## Tool Calling Instructions (Gemini functionDeclarations)\n")
		b.WriteString("Tools are registered as `functionDeclarations`. When you need to use a tool,\n")
		b.WriteString("respond with a `functionCall` part — do NOT call tools inline in text.\n")
		b.WriteString("The result will be provided as a `functionResponse` part in the next turn.\n\n")
		b.WriteString("Available tools:\n\n")
		for _, t := range tools {
			b.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", t.Name, t.Type, t.Description))
		}

	default:
		// Generic fallback for Ollama and other local/custom providers.
		b.WriteString("\n\n## Tool Calling Instructions\n")
		b.WriteString("You have access to the following tools. To use a tool, output a JSON object\n")
		b.WriteString("on its own line in this exact format:\n\n")
		b.WriteString("```json\n{\"tool\": \"<name>\", \"params\": {<key>: <value>, ...}}\n```\n\n")
		b.WriteString("Wait for the tool result before continuing. Do NOT invent tool outputs.\n\n")
		b.WriteString("Available tools:\n\n")
		for _, t := range tools {
			b.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", t.Name, t.Type, t.Description))
		}
	}

	return b.String()
}
