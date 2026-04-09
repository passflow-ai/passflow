// Package knowledge provides the KnowledgeBaseProvider abstraction for the
// Agent Compatibility Framework (ACF). It defines the Provider interface,
// the Chunk and KnowledgeConfig types, and the Manager which fans out
// retrieval across multiple registered providers and merges results by score.
package knowledge

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Chunk represents a single piece of retrieved knowledge.
type Chunk struct {
	Content  string            `json:"content"`
	Source   string            `json:"source"`
	Score    float64           `json:"score"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// KnowledgeConfig defines a knowledge base configuration attached to an agent.
type KnowledgeConfig struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"` // code_rag, document_rag, issue_rag, sql_rag, social_rag, graph_rag
	Source string                 `json:"source"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// Provider is the interface every RAG knowledge source must implement.
type Provider interface {
	// Type returns the provider type string (e.g. "document_rag").
	Type() string
	// Retrieve returns the topK most relevant chunks for the given query.
	Retrieve(ctx context.Context, query string, topK int) ([]Chunk, error)
}

// Manager manages multiple named knowledge providers for a single agent
// execution. Providers are registered by name so they can be queried
// individually (RetrieveFrom) or collectively (Retrieve).
type Manager struct {
	providers map[string]Provider
}

// NewManager returns an empty Manager ready for provider registration.
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

// Register adds provider under name. If name is already registered the
// previous provider is replaced.
func (m *Manager) Register(name string, provider Provider) {
	m.providers[name] = provider
}

// Retrieve queries every registered provider with query, collects all chunks,
// deduplicates by content, sorts descending by score, and returns the global
// top topK results. If topK is <= 0 all chunks are returned.
func (m *Manager) Retrieve(ctx context.Context, query string, topK int) ([]Chunk, error) {
	var merged []Chunk

	for name, p := range m.providers {
		chunks, err := p.Retrieve(ctx, query, topK)
		if err != nil {
			// Partial failure: log source name in error but continue so that
			// other providers can still contribute results.
			return nil, fmt.Errorf("knowledge provider %q: %w", name, err)
		}
		merged = append(merged, chunks...)
	}

	// Sort by descending score so the most relevant chunks come first.
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})

	if topK > 0 && len(merged) > topK {
		merged = merged[:topK]
	}

	return merged, nil
}

// RetrieveFrom queries a single provider identified by name.
// It returns an error if no provider with that name is registered.
func (m *Manager) RetrieveFrom(ctx context.Context, name, query string, topK int) ([]Chunk, error) {
	p, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("knowledge provider %q not registered", name)
	}
	return p.Retrieve(ctx, query, topK)
}

// FormatContext converts chunks into a plain-text block suitable for
// injection into an LLM system prompt or user message. Each chunk is
// preceded by its source and separated by a horizontal rule.
//
// Example output:
//
//	[Source: docs/README.md (score: 0.92)]
//	This is the chunk content...
//	---
//	[Source: docs/CONTRIBUTING.md (score: 0.87)]
//	Another relevant section...
//	---
func (m *Manager) FormatContext(chunks []Chunk) string {
	if len(chunks) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, c := range chunks {
		fmt.Fprintf(&sb, "[Source: %s (score: %.2f)]\n%s", c.Source, c.Score, c.Content)
		if i < len(chunks)-1 {
			sb.WriteString("\n---\n")
		}
	}
	return sb.String()
}
