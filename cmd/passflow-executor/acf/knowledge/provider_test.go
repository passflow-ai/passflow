package knowledge_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jaak-ai/passflow-agent-executor/acf/knowledge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- DocumentRAGProvider tests ---

func TestDocumentRAGProvider_Type(t *testing.T) {
	p := knowledge.NewDocumentRAGProvider(nil)
	assert.Equal(t, "document_rag", p.Type())
}

func TestDocumentRAGProvider_Retrieve_KeywordMatch(t *testing.T) {
	docs := []knowledge.Document{
		{Content: "Go is a statically typed language.", Source: "go.md"},
		{Content: "Python is dynamically typed.", Source: "python.md"},
		{Content: "Go excels at concurrency with goroutines.", Source: "concurrency.md"},
	}
	p := knowledge.NewDocumentRAGProvider(docs)
	ctx := context.Background()

	chunks, err := p.Retrieve(ctx, "Go language", 10)
	require.NoError(t, err)

	// Both "go.md" and "concurrency.md" contain "go"; only "go.md" also has
	// "language" so it should rank first (higher score).
	require.NotEmpty(t, chunks)
	assert.Equal(t, "go.md", chunks[0].Source, "go.md should rank first — matches both keywords")
}

func TestDocumentRAGProvider_Retrieve_TopKLimit(t *testing.T) {
	docs := make([]knowledge.Document, 5)
	for i := range docs {
		docs[i] = knowledge.Document{
			Content: "Go programming language",
			Source:  "doc.md",
		}
	}
	p := knowledge.NewDocumentRAGProvider(docs)

	chunks, err := p.Retrieve(context.Background(), "Go", 2)
	require.NoError(t, err)
	assert.Len(t, chunks, 2, "topK=2 must cap the result set")
}

func TestDocumentRAGProvider_Retrieve_NoMatch(t *testing.T) {
	docs := []knowledge.Document{
		{Content: "Python is great for data science.", Source: "python.md"},
	}
	p := knowledge.NewDocumentRAGProvider(docs)

	chunks, err := p.Retrieve(context.Background(), "rust ownership", 5)
	require.NoError(t, err)
	assert.Empty(t, chunks, "no documents should match an unrelated query")
}

func TestDocumentRAGProvider_Retrieve_EmptyDocs(t *testing.T) {
	p := knowledge.NewDocumentRAGProvider(nil)

	chunks, err := p.Retrieve(context.Background(), "anything", 5)
	require.NoError(t, err)
	assert.Empty(t, chunks)
}

func TestDocumentRAGProvider_Retrieve_EmptyQuery(t *testing.T) {
	docs := []knowledge.Document{
		{Content: "Some content here.", Source: "doc.md"},
	}
	p := knowledge.NewDocumentRAGProvider(docs)

	chunks, err := p.Retrieve(context.Background(), "   ", 5)
	require.NoError(t, err)
	assert.Empty(t, chunks, "blank query should return nothing")
}

func TestDocumentRAGProvider_Retrieve_ScoreRange(t *testing.T) {
	docs := []knowledge.Document{
		{Content: "alpha beta gamma", Source: "full.md"},
		{Content: "alpha only", Source: "partial.md"},
	}
	p := knowledge.NewDocumentRAGProvider(docs)

	chunks, err := p.Retrieve(context.Background(), "alpha beta gamma", 10)
	require.NoError(t, err)
	require.Len(t, chunks, 2)

	for _, c := range chunks {
		assert.GreaterOrEqual(t, c.Score, 0.0)
		assert.LessOrEqual(t, c.Score, 1.0)
	}
	// full.md matches all 3 keywords → score 1.0
	assert.InDelta(t, 1.0, chunks[0].Score, 0.001)
	assert.Equal(t, "full.md", chunks[0].Source)
}

func TestDocumentRAGProvider_Retrieve_MetadataPropagated(t *testing.T) {
	docs := []knowledge.Document{
		{
			Content:  "Kubernetes pod scheduling",
			Source:   "k8s.md",
			Metadata: map[string]string{"author": "alice", "version": "1.2"},
		},
	}
	p := knowledge.NewDocumentRAGProvider(docs)

	chunks, err := p.Retrieve(context.Background(), "Kubernetes", 5)
	require.NoError(t, err)
	require.Len(t, chunks, 1)
	assert.Equal(t, "alice", chunks[0].Metadata["author"])
	assert.Equal(t, "1.2", chunks[0].Metadata["version"])
}

func TestDocumentRAGProvider_Retrieve_TopKZeroReturnsAll(t *testing.T) {
	docs := []knowledge.Document{
		{Content: "Go concurrency", Source: "a.md"},
		{Content: "Go channels", Source: "b.md"},
		{Content: "Go goroutines", Source: "c.md"},
	}
	p := knowledge.NewDocumentRAGProvider(docs)

	chunks, err := p.Retrieve(context.Background(), "Go", 0)
	require.NoError(t, err)
	assert.Len(t, chunks, 3, "topK=0 should return all matching results")
}

// --- Manager tests ---

func TestManager_Register_And_Retrieve(t *testing.T) {
	m := knowledge.NewManager()

	docs := []knowledge.Document{
		{Content: "Redis is an in-memory data store.", Source: "redis.md"},
	}
	m.Register("docs", knowledge.NewDocumentRAGProvider(docs))

	chunks, err := m.Retrieve(context.Background(), "Redis", 5)
	require.NoError(t, err)
	require.Len(t, chunks, 1)
	assert.Equal(t, "redis.md", chunks[0].Source)
}

func TestManager_Retrieve_MergesByScore(t *testing.T) {
	m := knowledge.NewManager()

	// Provider A: one highly relevant doc.
	docsA := []knowledge.Document{
		{Content: "Go channels and goroutines are fast.", Source: "concurrency.md"},
	}
	// Provider B: a less relevant doc that still matches.
	docsB := []knowledge.Document{
		{Content: "Go is also used in systems programming.", Source: "systems.md"},
	}

	m.Register("providerA", knowledge.NewDocumentRAGProvider(docsA))
	m.Register("providerB", knowledge.NewDocumentRAGProvider(docsB))

	chunks, err := m.Retrieve(context.Background(), "Go channels goroutines", 10)
	require.NoError(t, err)
	require.Len(t, chunks, 2, "should receive one chunk from each provider")

	// Merged result must be sorted by descending score.
	assert.GreaterOrEqual(t, chunks[0].Score, chunks[1].Score)
}

func TestManager_Retrieve_TopKAcrossProviders(t *testing.T) {
	m := knowledge.NewManager()

	makeDocs := func(content, source string) []knowledge.Document {
		return []knowledge.Document{{Content: content, Source: source}}
	}
	m.Register("p1", knowledge.NewDocumentRAGProvider(makeDocs("Go language runtime", "p1.md")))
	m.Register("p2", knowledge.NewDocumentRAGProvider(makeDocs("Go standard library", "p2.md")))
	m.Register("p3", knowledge.NewDocumentRAGProvider(makeDocs("Go toolchain build", "p3.md")))

	chunks, err := m.Retrieve(context.Background(), "Go", 2)
	require.NoError(t, err)
	assert.Len(t, chunks, 2, "Manager must cap merged results at topK")
}

func TestManager_RetrieveFrom_TargetsSpecificProvider(t *testing.T) {
	m := knowledge.NewManager()

	docsA := []knowledge.Document{{Content: "MongoDB aggregation pipelines", Source: "mongo.md"}}
	docsB := []knowledge.Document{{Content: "PostgreSQL window functions", Source: "pg.md"}}
	m.Register("mongo", knowledge.NewDocumentRAGProvider(docsA))
	m.Register("postgres", knowledge.NewDocumentRAGProvider(docsB))

	chunks, err := m.RetrieveFrom(context.Background(), "postgres", "PostgreSQL", 5)
	require.NoError(t, err)
	require.Len(t, chunks, 1)
	assert.Equal(t, "pg.md", chunks[0].Source, "must only query the postgres provider")
}

func TestManager_RetrieveFrom_UnknownProvider(t *testing.T) {
	m := knowledge.NewManager()

	_, err := m.RetrieveFrom(context.Background(), "nonexistent", "query", 5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestManager_Retrieve_EmptyManagerReturnsEmpty(t *testing.T) {
	m := knowledge.NewManager()

	chunks, err := m.Retrieve(context.Background(), "anything", 5)
	require.NoError(t, err)
	assert.Empty(t, chunks)
}

func TestManager_FormatContext_ProducesReadableOutput(t *testing.T) {
	m := knowledge.NewManager()

	chunks := []knowledge.Chunk{
		{Content: "First relevant paragraph.", Source: "doc1.md", Score: 0.95},
		{Content: "Second relevant paragraph.", Source: "doc2.md", Score: 0.80},
	}

	out := m.FormatContext(chunks)
	assert.Contains(t, out, "doc1.md")
	assert.Contains(t, out, "First relevant paragraph.")
	assert.Contains(t, out, "doc2.md")
	assert.Contains(t, out, "Second relevant paragraph.")
	assert.Contains(t, out, "0.95")
	assert.Contains(t, out, "0.80")
	// Chunks should be separated by a rule.
	assert.Contains(t, out, "---")
}

func TestManager_FormatContext_EmptyChunks(t *testing.T) {
	m := knowledge.NewManager()
	out := m.FormatContext(nil)
	assert.Equal(t, "", out)
}

func TestManager_FormatContext_SingleChunkNoSeparator(t *testing.T) {
	m := knowledge.NewManager()
	chunks := []knowledge.Chunk{
		{Content: "Only chunk.", Source: "only.md", Score: 1.0},
	}
	out := m.FormatContext(chunks)
	assert.NotContains(t, out, "---", "single chunk should not have a trailing separator")
	assert.True(t, strings.HasPrefix(out, "[Source:"), "output should start with source header")
}

// --- Stub provider tests ---

func TestCodeRAGProvider_Type(t *testing.T) {
	p := knowledge.NewCodeRAGProvider()
	assert.Equal(t, "code_rag", p.Type())
}

func TestCodeRAGProvider_Retrieve_ReturnsEmpty(t *testing.T) {
	p := knowledge.NewCodeRAGProvider()
	chunks, err := p.Retrieve(context.Background(), "any query", 5)
	require.NoError(t, err)
	assert.Empty(t, chunks)
}

func TestSQLRAGProvider_Type(t *testing.T) {
	p := knowledge.NewSQLRAGProvider()
	assert.Equal(t, "sql_rag", p.Type())
}

func TestSQLRAGProvider_Retrieve_ReturnsEmpty(t *testing.T) {
	p := knowledge.NewSQLRAGProvider()
	chunks, err := p.Retrieve(context.Background(), "SELECT * FROM users", 5)
	require.NoError(t, err)
	assert.Empty(t, chunks)
}

// --- Interface compliance ---

// Compile-time assertions that all providers satisfy the Provider interface.
var (
	_ knowledge.Provider = (*knowledge.DocumentRAGProvider)(nil)
	_ knowledge.Provider = (*knowledge.CodeRAGProvider)(nil)
	_ knowledge.Provider = (*knowledge.SQLRAGProvider)(nil)
)
