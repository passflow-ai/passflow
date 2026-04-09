package engine

import (
	"time"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/job"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
)

func buildInitialMessages(spec *job.Spec) []llm.Message {
	var msgs []llm.Message
	if spec.SystemPrompt != "" {
		msgs = append(msgs, llm.Message{Role: "system", Content: spec.SystemPrompt})
	}
	msgs = append(msgs, llm.Message{Role: "user", Content: spec.Input})
	return msgs
}

func addUsage(total *Usage, u llm.Usage) {
	total.PromptTokens += u.PromptTokens
	total.CompletionTokens += u.CompletionTokens
	total.TotalTokens += u.TotalTokens
}

func recordLLMUsage(total *Usage, u llm.Usage, latency time.Duration) {
	addUsage(total, u)
	total.LLMCalls++
	total.LLMLatencyMs += latency.Milliseconds()
}

func recordToolUsage(total *Usage, err string) {
	total.ToolCalls++
	if err != "" {
		total.ToolErrors++
	}
}

func findToolArgs(calls []llm.ToolCall, id string) string {
	for _, c := range calls {
		if c.ID == id {
			return c.Arguments
		}
	}
	return ""
}

func lastUserMessage(msgs []llm.Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			return msgs[i].Content
		}
	}
	return ""
}
