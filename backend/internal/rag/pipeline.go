package rag

import (
	"context"
	"fmt"
	"log"
	"strings"

	"agri-qa-system/internal/model"
	"agri-qa-system/internal/store"
)

type Pipeline struct {
	qdrant   *store.QdrantStore
	embed    *store.EmbeddingClient
	reranker *store.RerankerClient
}

func NewPipeline(qdrant *store.QdrantStore, embed *store.EmbeddingClient, reranker *store.RerankerClient) *Pipeline {
	return &Pipeline{
		qdrant:   qdrant,
		embed:    embed,
		reranker: reranker,
	}
}

func (p *Pipeline) Retrieve(ctx context.Context, query string, topK int) ([]model.SearchResult, error) {
	vector, err := p.embed.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	log.Printf("[RAG] Embedding OK, dim=%d", len(vector))

	recallK := topK * 2
	results, err := p.qdrant.Search(ctx, vector, recallK)
	if err != nil {
		return nil, fmt.Errorf("qdrant search: %w", err)
	}
	log.Printf("[RAG] Qdrant search returned %d results", len(results))

	if len(results) == 0 {
		return nil, nil
	}

	// rerank
	docs := make([]string, len(results))
	for i, r := range results {
		docs[i] = r.Content
	}
	indices, _, err := p.reranker.Rerank(query, docs)
	if err != nil {
		log.Printf("[RAG] Reranker failed: %v, fallback to top-%d", err, topK)
		// rerank failed, fallback to top-K by vector score
		if len(results) > topK {
			results = results[:topK]
		}
		log.Printf("[RAG] Returning %d results (fallback)", len(results))
		return results, nil
	}

	log.Printf("[RAG] Reranker OK, indices=%v", indices)
	var reranked []model.SearchResult
	for i := 0; i < len(indices) && len(reranked) < topK; i++ {
		reranked = append(reranked, results[indices[i]])
	}
	log.Printf("[RAG] Returning %d results (reranked)", len(reranked))
	return reranked, nil
}

func (p *Pipeline) BuildContext(results []model.SearchResult) (string, []model.Citation) {
	var contextParts []string
	var citations []model.Citation

	for i, r := range results {
		contextParts = append(contextParts, fmt.Sprintf("[参考%d] %s", i+1, r.Content))
		citations = append(citations, model.Citation{
			BookTitle: getMetaStr(r.Metadata, "book_title"),
			Chapter:   getMetaStr(r.Metadata, "chapter"),
			Content:   r.Content,
			Score:     r.Score,
		})
	}
	return strings.Join(contextParts, "\n\n"), citations
}

func getMetaStr(meta map[string]any, key string) string {
	if v, ok := meta[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
