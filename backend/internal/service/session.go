package service

import (
	"context"

	"agri-qa-system/internal/model"
)

func (s *ChatService) SaveAnswer(ctx context.Context, sessionID, answer string) {
	s.redis.AddMessage(ctx, sessionID, model.Message{Role: "assistant", Content: answer})
}

func (s *ChatService) ListSessions(ctx context.Context) ([]model.Session, error) {
	return s.redis.ListSessions(ctx)
}

func (s *ChatService) GetSession(ctx context.Context, id string) (*model.Session, error) {
	return s.redis.GetSession(ctx, id)
}

func (s *ChatService) DeleteSession(ctx context.Context, id string) error {
	return s.redis.DeleteSession(ctx, id)
}
