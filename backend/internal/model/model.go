package model

import "time"

type Message struct {
	Role    string `json:"role"`    // user / assistant / system
	Content string `json:"content"`
}

type Citation struct {
	BookTitle string  `json:"book_title"`
	Chapter   string  `json:"chapter,omitempty"`
	Content   string  `json:"content"`
	Page      int     `json:"page,omitempty"`
	Score     float64 `json:"score"`
	Category  string  `json:"category,omitempty"` // 症状/药材/方剂/养生/穴位
}

type ChatRequest struct {
	SessionID string `json:"session_id,omitempty"`
	Query     string `json:"query"`
}

type ChatResponse struct {
	SessionID string    `json:"session_id"`
	Answer    string    `json:"answer"`
	Citations []Citation `json:"citations"`
}

type SSEEvent struct {
	Type    string `json:"type"` // token / citation / done / error
	Content string `json:"content"`
	Data    any    `json:"data,omitempty"`
}

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SearchResult struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Score    float64           `json:"score"`
	Metadata map[string]any    `json:"metadata"`
}
