package knowledge

import "context"

// SQLRAGProvider is a stub for schema-injection and Text-to-SQL/MQL knowledge
// retrieval. Full implementation will introspect a target database schema and
// generate SQL/MQL queries from natural language via the LLM. Supported
// backends are PostgreSQL, MongoDB, and BigQuery. Deferred to a future
// iteration pending vector store and SQL generation integration.
type SQLRAGProvider struct{}

// NewSQLRAGProvider returns an uninitialized SQLRAGProvider stub.
func NewSQLRAGProvider() *SQLRAGProvider {
	return &SQLRAGProvider{}
}

// Type implements Provider.
func (p *SQLRAGProvider) Type() string { return "sql_rag" }

// Retrieve implements Provider. Returns empty results until the full
// schema-injection pipeline is wired in.
func (p *SQLRAGProvider) Retrieve(_ context.Context, _ string, _ int) ([]Chunk, error) {
	return nil, nil
}
