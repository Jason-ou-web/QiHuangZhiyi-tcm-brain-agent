package memory

import (
	"context"

	"agri-qa-system/internal/model"
	"agri-qa-system/internal/store"
)

// ShortTermMemory wraps Redis for conversation context
type ShortTermMemory struct {
	redis *store.RedisStore
}

func NewShortTermMemory(redis *store.RedisStore) *ShortTermMemory {
	return &ShortTermMemory{redis: redis}
}

func (m *ShortTermMemory) GetHistory(ctx context.Context, sessionID string) ([]model.Message, error) {
	session, err := m.redis.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}
	messages := session.Messages
	if len(messages) > 10 {
		messages = messages[len(messages)-10:]
	}
	return messages, nil
}

func (m *ShortTermMemory) SaveMessage(ctx context.Context, sessionID string, msg model.Message) error {
	return m.redis.AddMessage(ctx, sessionID, msg)
}
