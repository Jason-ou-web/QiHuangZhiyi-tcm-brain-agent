package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type EmbeddingClient struct {
	baseURL    string
	httpClient *http.Client
}

type embeddingRequest struct {
	Texts []string `json:"texts"`
}

type embeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func NewEmbeddingClient(baseURL string) *EmbeddingClient {
	return &EmbeddingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *EmbeddingClient) Embed(text string) ([]float32, error) {
	body := embeddingRequest{Texts: []string{text}}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()

	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embedding: %w", err)
	}
	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("empty embedding")
	}
	return result.Embeddings[0], nil
}

type RerankerClient struct {
	baseURL    string
	httpClient *http.Client
}

type rerankRequest struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
}

type rerankResponse struct {
	Results []struct {
		Index int     `json:"index"`
		Score float64 `json:"score"`
	} `json:"results"`
}

func NewRerankerClient(baseURL string) *RerankerClient {
	return &RerankerClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *RerankerClient) Rerank(query string, documents []string) ([]int, []float64, error) {
	body := rerankRequest{Query: query, Documents: documents}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal rerank request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("rerank request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("rerank service returned %d", resp.StatusCode)
	}

	var result rerankResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("decode rerank: %w", err)
	}
	indices := make([]int, len(result.Results))
	scores := make([]float64, len(result.Results))
	for i, r := range result.Results {
		indices[i] = r.Index
		scores[i] = r.Score
	}
	return indices, scores, nil
}
