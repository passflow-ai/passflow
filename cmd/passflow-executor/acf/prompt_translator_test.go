package acf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers -----------------------------------------------------------------

func newCAD(systemPrompt string) *CADConfig {
	return &CADConfig{
		SystemPrompt: systemPrompt,
		SourceType:   "claude_code",
	}
}

func sampleTools() []ToolDef {
	return []ToolDef{
		{Name: "Read", Description: "Read a file from disk", Type: "system"},
		{Name: "jira_update", Description: "Update a Jira ticket", Type: "mcp"},
	}
}

// --- nil / empty guard -------------------------------------------------------

func TestTranslate_NilCAD_ReturnsEmpty(t *testing.T) {
	pt := NewPromptTranslator()
	result := pt.Translate("anthropic", nil, sampleTools())
	assert.Empty(t, result)
}

func TestTranslate_EmptySystemPrompt_NoTools_ReturnsEmpty(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("")
	result := pt.Translate("anthropic", cad, nil)
	// No system prompt and no tools → empty string (the base prompt is "").
	assert.Empty(t, result)
}

// --- base system prompt ------------------------------------------------------

func TestTranslate_BaseSystemPromptIsIncluded(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("You are a helpful assistant.")
	result := pt.Translate("anthropic", cad, nil)
	assert.Contains(t, result, "You are a helpful assistant.")
}

// --- skills ------------------------------------------------------------------

func TestTranslate_SkillsAreAppended(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base prompt.")
	cad.Skills = []SkillDef{
		{Name: "code-review", Content: "Always review for security issues."},
		{Name: "summariser", Content: "Summarise findings concisely."},
	}

	result := pt.Translate("anthropic", cad, nil)

	assert.Contains(t, result, "## Skills")
	assert.Contains(t, result, "### code-review")
	assert.Contains(t, result, "Always review for security issues.")
	assert.Contains(t, result, "### summariser")
	assert.Contains(t, result, "Summarise findings concisely.")
}

func TestTranslate_NoSkills_NoSkillsSection(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base prompt.")
	result := pt.Translate("anthropic", cad, nil)
	assert.NotContains(t, result, "## Skills")
}

// --- context files -----------------------------------------------------------

func TestTranslate_ContextFilesAreIncluded(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base prompt.")
	cad.ContextFiles = map[string]string{
		"CLAUDE.md": "# Project Guide\nThis is the project guide.",
	}

	result := pt.Translate("anthropic", cad, nil)

	assert.Contains(t, result, "## Context Files")
	assert.Contains(t, result, "### CLAUDE.md")
	assert.Contains(t, result, "# Project Guide")
}

func TestTranslate_NoContextFiles_NoContextFilesSection(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base prompt.")
	result := pt.Translate("anthropic", cad, nil)
	assert.NotContains(t, result, "## Context Files")
}

// --- provider differentiation ------------------------------------------------

func TestTranslate_Anthropic_UsesToolUseFormat(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base.")
	result := pt.Translate("anthropic", cad, sampleTools())

	assert.Contains(t, result, "tool_use")
	// Should NOT contain OpenAI-specific language.
	assert.NotContains(t, result, "function_call")
	assert.NotContains(t, result, "functionDeclarations")
}

func TestTranslate_OpenAI_EmphasisesFunctionCallFormat(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base.")
	result := pt.Translate("openai", cad, sampleTools())

	assert.Contains(t, result, "function_call")
	assert.Contains(t, result, "strict JSON")
	assert.NotContains(t, result, "tool_use")
}

func TestTranslate_Azure_SameAsOpenAI(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base.")
	openaiResult := pt.Translate("openai", cad, sampleTools())
	azureResult := pt.Translate("azure", cad, sampleTools())
	assert.Equal(t, openaiResult, azureResult)
}

func TestTranslate_Gemini_UsesFunctionDeclarations(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base.")
	result := pt.Translate("gemini", cad, sampleTools())

	assert.Contains(t, result, "functionDeclarations")
	assert.Contains(t, result, "functionCall")
	assert.NotContains(t, result, "function_call")
}

func TestTranslate_Ollama_UsesGenericJSONInstructions(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base.")
	result := pt.Translate("ollama", cad, sampleTools())

	assert.Contains(t, result, `"tool"`)
	assert.Contains(t, result, `"params"`)
	// Generic path must not bleed OpenAI / Gemini / Anthropic keywords.
	assert.NotContains(t, result, "function_call")
	assert.NotContains(t, result, "tool_use")
	assert.NotContains(t, result, "functionDeclarations")
}

func TestTranslate_UnknownProvider_FallsBackToGeneric(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base.")
	ollamaResult := pt.Translate("ollama", cad, sampleTools())
	unknownResult := pt.Translate("unknown-provider", cad, sampleTools())
	assert.Equal(t, ollamaResult, unknownResult)
}

// Each provider must produce a distinct output (not identical to another provider).
func TestTranslate_ProvidersProduceDifferentOutput(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("You are an assistant.")
	tools := sampleTools()

	anthropic := pt.Translate("anthropic", cad, tools)
	openai := pt.Translate("openai", cad, tools)
	gemini := pt.Translate("gemini", cad, tools)
	ollama := pt.Translate("ollama", cad, tools)

	require.NotEqual(t, anthropic, openai, "anthropic and openai must differ")
	require.NotEqual(t, anthropic, gemini, "anthropic and gemini must differ")
	require.NotEqual(t, anthropic, ollama, "anthropic and ollama must differ")
	require.NotEqual(t, openai, gemini, "openai and gemini must differ")
	require.NotEqual(t, openai, ollama, "openai and ollama must differ")
	require.NotEqual(t, gemini, ollama, "gemini and ollama must differ")
}

// Tool names and descriptions must appear in every provider's output.
func TestTranslate_ToolListAppearsInAllProviders(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("Base.")
	tools := sampleTools()

	for _, provider := range []string{"anthropic", "openai", "azure", "gemini", "ollama"} {
		result := pt.Translate(provider, cad, tools)
		assert.Contains(t, result, "Read", "provider %s: tool Read missing", provider)
		assert.Contains(t, result, "jira_update", "provider %s: tool jira_update missing", provider)
	}
}

// --- truncation --------------------------------------------------------------

func TestTranslate_Truncation_LongPromptIsShortened(t *testing.T) {
	pt := &PromptTranslator{MaxLength: 100}
	longPrompt := strings.Repeat("A", 200)
	cad := newCAD(longPrompt)

	result := pt.Translate("anthropic", cad, nil)

	assert.LessOrEqual(t, len(result), 100, "result must not exceed MaxLength")
	assert.Contains(t, result, "[NOTICE:", "truncation notice must be present")
}

func TestTranslate_Truncation_ShortPromptIsUnchanged(t *testing.T) {
	pt := &PromptTranslator{MaxLength: 100}
	cad := newCAD("Short prompt.")

	result := pt.Translate("anthropic", cad, nil)

	assert.Equal(t, "Short prompt.", result)
	assert.NotContains(t, result, "[NOTICE:")
}

func TestTranslate_Truncation_DefaultMaxLengthApplied(t *testing.T) {
	pt := NewPromptTranslator()
	// Default max is 100 000; generate something well within that.
	cad := newCAD(strings.Repeat("X", 500))
	result := pt.Translate("openai", cad, sampleTools())
	assert.LessOrEqual(t, len(result), defaultMaxLength)
}

func TestTranslate_Truncation_ZeroMaxLengthUsesDefault(t *testing.T) {
	pt := &PromptTranslator{MaxLength: 0}
	cad := newCAD(strings.Repeat("Y", 500))
	result := pt.Translate("openai", cad, sampleTools())
	assert.LessOrEqual(t, len(result), defaultMaxLength)
	assert.NotContains(t, result, "[NOTICE:")
}

// --- ordering ----------------------------------------------------------------

// System prompt must appear before skills, which must appear before tool instructions.
func TestTranslate_SectionOrdering(t *testing.T) {
	pt := NewPromptTranslator()
	cad := newCAD("SYSTEM_PROMPT_SENTINEL")
	cad.Skills = []SkillDef{{Name: "s1", Content: "SKILL_SENTINEL"}}
	cad.ContextFiles = map[string]string{"f.md": "FILE_SENTINEL"}

	result := pt.Translate("openai", cad, sampleTools())

	systemIdx := strings.Index(result, "SYSTEM_PROMPT_SENTINEL")
	fileIdx := strings.Index(result, "FILE_SENTINEL")
	skillIdx := strings.Index(result, "SKILL_SENTINEL")
	toolIdx := strings.Index(result, "function_call")

	require.Greater(t, fileIdx, systemIdx, "context files should follow system prompt")
	require.Greater(t, skillIdx, fileIdx, "skills should follow context files")
	require.Greater(t, toolIdx, skillIdx, "tool instructions should follow skills")
}
