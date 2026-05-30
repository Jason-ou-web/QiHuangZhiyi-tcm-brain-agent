package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"agri-qa-system/config"
	"agri-qa-system/internal/model"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

type chatRequest struct {
	Model    string          `json:"model"`
	Messages []model.Message `json:"messages"`
	Stream   bool            `json:"stream"`
}

type chatChoice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type chatStreamChunk struct {
	Choices []chatChoice `json:"choices"`
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) Chat(messages []model.Message) (string, error) {
	body := chatRequest{
		Model:    c.cfg.DeepSeekModel,
		Messages: messages,
		Stream:   false,
	}
	resp, err := c.doRequest("/v1/chat/completions", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return result.Choices[0].Message.Content, nil
}

func (c *Client) ChatStream(ctx context.Context, messages []model.Message, eventCh chan<- model.SSEEvent) {
	defer close(eventCh)

	body := chatRequest{
		Model:    c.cfg.DeepSeekModel,
		Messages: messages,
		Stream:   true,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		sendOrDrop(eventCh, model.SSEEvent{Type: "error", Content: err.Error()})
		return
	}

	url := c.cfg.DeepSeekBaseURL + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		sendOrDrop(eventCh, model.SSEEvent{Type: "error", Content: err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.DeepSeekAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		sendOrDrop(eventCh, model.SSEEvent{Type: "error", Content: err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		sendOrDrop(eventCh, model.SSEEvent{Type: "error", Content: fmt.Sprintf("API error %d: %s", resp.StatusCode, string(bodyBytes))})
		return
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			sendOrDrop(eventCh, model.SSEEvent{Type: "error", Content: err.Error()})
			return
		}
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk chatStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				fullContent.WriteString(choice.Delta.Content)
				sendOrDrop(eventCh, model.SSEEvent{
					Type:    "token",
					Content: choice.Delta.Content,
				})
			}
		}
	}
	sendOrDrop(eventCh, model.SSEEvent{
		Type:    "done",
		Content: fullContent.String(),
	})
}

func sendOrDrop(ch chan<- model.SSEEvent, event model.SSEEvent) {
	select {
	case ch <- event:
	default:
	}
}

func (c *Client) doRequest(path string, body any) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.cfg.DeepSeekBaseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.DeepSeekAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return resp, nil
}
