package service

import (
	"context"
	"fmt"
	"log"

	"agri-qa-system/internal/llm"
	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
	"agri-qa-system/internal/store"
)

const (
	defaultTopK     = 5
	systemPrompt    = `你是一个专业的中医知识助手，基于提供的中医古籍和现代文献参考内容回答用户问题。

要求：
1. 仅根据参考内容回答，不要编造信息
2. 如果参考内容不足以回答，请明确说明
3. 回答中引用参考内容时，使用 [参考N] 标注
4. 使用通俗易懂的中文，中医专业术语需做简要解释
5. 回答结构清晰，适当分点，包含辨证分析、调理建议和注意事项
6. 涉及诊断、用药建议时，务必提醒用户仅供参考，建议咨询专业中医师`
)

type ChatService struct {
	rag     *rag.Pipeline
	llm     *llm.Client
	redis   *store.RedisStore
}

func NewChatService(rag *rag.Pipeline, llm *llm.Client, redis *store.RedisStore) *ChatService {
	return &ChatService{rag: rag, llm: llm, redis: redis}
}

func (s *ChatService) Chat(ctx context.Context, sessionID, query string) (string, []model.Citation, error) {
	// get or create session
	history, err := s.getHistory(ctx, sessionID)
	if err != nil {
		return "", nil, err
	}

	// retrieve knowledge
	results, err := s.rag.Retrieve(ctx, query, defaultTopK)
	if err != nil {
		return "", nil, fmt.Errorf("retrieve: %w", err)
	}

	contextStr, citations := s.rag.BuildContext(results)

	// build messages
	messages := buildMessages(history, contextStr, query)

	// save user message
	s.redis.AddMessage(ctx, sessionID, model.Message{Role: "user", Content: query})

	// generate
	answer, err := s.llm.Chat(messages)
	if err != nil {
		return "", nil, fmt.Errorf("llm: %w", err)
	}

	// save assistant message
	s.redis.AddMessage(ctx, sessionID, model.Message{Role: "assistant", Content: answer})

	return answer, citations, nil
}

func (s *ChatService) ChatStream(ctx context.Context, sessionID, query string) (<-chan model.SSEEvent, []model.Citation, error) {
	history, err := s.getHistory(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	results, err := s.rag.Retrieve(ctx, query, defaultTopK)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieve: %w", err)
	}

	contextStr, citations := s.rag.BuildContext(results)

	// send citations as first events via the same channel? No, return separately
	messages := buildMessages(history, contextStr, query)

	s.redis.AddMessage(ctx, sessionID, model.Message{Role: "user", Content: query})

	eventCh := make(chan model.SSEEvent, 100)
	go func() {
		s.llm.ChatStream(ctx, messages, eventCh)
	}()

	return eventCh, citations, nil
}

func (s *ChatService) getHistory(ctx context.Context, sessionID string) ([]model.Message, error) {
	if sessionID == "" {
		return nil, nil
	}
	session, err := s.redis.GetSession(ctx, sessionID)
	if err != nil {
		log.Printf("[ChatService] get history error: %v", err)
		return nil, nil
	}
	if session == nil {
		return nil, nil
	}
	// keep last 10 messages as context
	messages := session.Messages
	if len(messages) > 10 {
		messages = messages[len(messages)-10:]
	}
	return messages, nil
}

func buildMessages(history []model.Message, contextStr, query string) []model.Message {
	messages := []model.Message{
		{Role: "system", Content: systemPrompt},
	}
	if contextStr != "" {
		messages = append(messages, model.Message{
			Role:    "system",
			Content: fmt.Sprintf("参考内容：\n%s", contextStr),
		})
	}
	messages = append(messages, history...)
	messages = append(messages, model.Message{Role: "user", Content: query})

	// estimate token count (rough: 1 token ≈ 2 Chinese chars)
	totalChars := 0
	for _, m := range messages {
		totalChars += len([]rune(m.Content))
	}
	// if over ~6000 tokens, trim history
	if totalChars > 12000 {
		trimStart := len(history) / 3
		if trimStart > 0 {
			keep := history[trimStart:]
			newMessages := []model.Message{
				{Role: "system", Content: systemPrompt},
				{Role: "system", Content: fmt.Sprintf("参考内容：\n%s", contextStr)},
			}
			newMessages = append(newMessages, keep...)
			newMessages = append(newMessages, model.Message{Role: "user", Content: query})
			messages = newMessages
		}
	}
	return messages
}
