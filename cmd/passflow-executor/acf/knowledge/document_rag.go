package knowledge

import (
	"context"
	"strings"
)

// Document is a single document stored in a DocumentRAGProvider.
type Document struct {
	Content  string
	Source   string
	Metadata map[string]string
}

// DocumentRAGProvider is an in-memory document store that implements Provider
// using keyword matching. This is the MVP implementation and will be replaced
// with embedding-based vector search in a future iteration.
type DocumentRAGProvider struct {
	documents []Document
}

// NewDocumentRAGProvider returns a DocumentRAGProvider pre-loaded with docs.
func NewDocumentRAGProvider(docs []Document) *DocumentRAGProvider {
	return &DocumentRAGProvider{documents: docs}
}

// Type implements Provider.
func (p *DocumentRAGProvider) Type() string { return "document_rag" }

// Retrieve searches all documents for query keywords, scores each document by
// the number of keyword matches normalised by the total keyword count, and
// returns the topK highest-scoring chunks. Documents with zero matches are
// excluded. If topK <= 0 all matching documents are returned.
//
// Scoring formula per document:
//
//	score = matchedKeywords / totalKeywords
//
// Tie-breaking preserves insertion order.
func (p *DocumentRAGProvider) Retrieve(_ context.Context, query string, topK int) ([]Chunk, error) {
	if len(p.documents) == 0 || strings.TrimSpace(query) == "" {
		return nil, nil
	}

	keywords := tokenise(query)
	if len(keywords) == 0 {
		return nil, nil
	}

	results := make([]scoredChunk, 0, len(p.documents))

	for _, doc := range p.documents {
		lower := strings.ToLower(doc.Content)
		matched := 0
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				matched++
			}
		}
		if matched == 0 {
			continue
		}
		score := float64(matched) / float64(len(keywords))

		meta := make(map[string]string, len(doc.Metadata))
		for k, v := range doc.Metadata {
			meta[k] = v
		}
		if len(meta) == 0 {
			meta = nil
		}

		results = append(results, scoredChunk{
			chunk: Chunk{
				Content:  doc.Content,
				Source:   doc.Source,
				Score:    score,
				Metadata: meta,
			},
			score: score,
		})
	}

	// Stable sort descending by score.
	sortScoredChunks(results)

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}

	chunks := make([]Chunk, len(results))
	for i, r := range results {
		chunks[i] = r.chunk
	}
	return chunks, nil
}

// tokenise splits query into lowercased, non-empty tokens by whitespace and
// common punctuation. Duplicate tokens are deduplicated so that repeated words
// do not artificially inflate the score denominator.
func tokenise(query string) []string {
	// Replace common punctuation with spaces.
	replacer := strings.NewReplacer(
		",", " ", ".", " ", "?", " ", "!", " ",
		";", " ", ":", " ", "(", " ", ")", " ",
		"\"", " ", "'", " ",
	)
	normalised := replacer.Replace(strings.ToLower(query))

	raw := strings.Fields(normalised)
	seen := make(map[string]bool, len(raw))
	unique := make([]string, 0, len(raw))
	for _, t := range raw {
		if t != "" && !seen[t] {
			seen[t] = true
			unique = append(unique, t)
		}
	}
	return unique
}

// scoredChunk pairs a Chunk with its relevance score for sorting.
type scoredChunk struct {
	chunk Chunk
	score float64
}

// sortScoredChunks sorts results in-place by descending score while preserving
// the relative order of equal-score entries (stable insertion sort).
func sortScoredChunks(results []scoredChunk) {
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].score > results[j-1].score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}
}
