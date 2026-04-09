package knowledge

import "context"

// CodeRAGProvider is a stub for AST-based code knowledge retrieval.
// Full implementation will use AST-based chunking (function/class indexing)
// against GitHub, GitLab, or locally cloned repositories. The vector store
// integration and embedding pipeline are deferred to a future iteration.
type CodeRAGProvider struct{}

// NewCodeRAGProvider returns an uninitialized CodeRAGProvider stub.
func NewCodeRAGProvider() *CodeRAGProvider {
	return &CodeRAGProvider{}
}

// Type implements Provider.
func (p *CodeRAGProvider) Type() string { return "code_rag" }

// Retrieve implements Provider. Returns empty results until the full
// AST-based pipeline is wired in.
func (p *CodeRAGProvider) Retrieve(_ context.Context, _ string, _ int) ([]Chunk, error) {
	return nil, nil
}
