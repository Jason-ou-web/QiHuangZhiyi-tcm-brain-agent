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
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
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
	body := embeddingRequest{Input: []string{text}}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("embedding service returned %d", resp.StatusCode)
	}

	// Infinity returns OpenAI-compatible format: {"data": [{"embedding": [...], "index": 0}]}
	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embedding: %w", err)
	}
	if len(result.Data) == 0 || len(result.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding")
	}
	return result.Data[0].Embedding, nil
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
	Data []struct {
		Index int     `json:"index"`
		Score float64 `json:"score"`
	} `json:"data"`
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

	// Infinity /rerank returns: {"data": [{"index":0,"score":0.9}, ...]}
	var result rerankResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("decode rerank: %w", err)
	}
	indices := make([]int, len(result.Data))
	scores := make([]float64, len(result.Data))
	for i, r := range result.Data {
		indices[i] = r.Index
		scores[i] = r.Score
	}
	return indices, scores, nil
}
