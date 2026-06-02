package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"agri-qa-system/config"
	"agri-qa-system/internal/model"

	"github.com/redis/go-redis/v9"
)

const (
	sessionPrefix   = "session:"
	sessionListKey  = "session:list"
	defaultTTL      = 24 * time.Hour
	maxMessages     = 50
	poolSize        = 20
	minIdleConns    = 5
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(cfg *config.Config) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           0,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxRetries:   3,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return &RedisStore{client: client}, nil
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

func (s *RedisStore) GetSession(ctx context.Context, id string) (*model.Session, error) {
	key := sessionPrefix + id
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var session model.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *RedisStore) DeleteSession(ctx context.Context, id string) error {
	pipe := s.client.Pipeline()
	pipe.Del(ctx, sessionPrefix+id)
	pipe.ZRem(ctx, sessionListKey, id)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) ListSessions(ctx context.Context) ([]model.Session, error) {
	// use ZREVRANGE on sorted set instead of KEYS command
	ids, err := s.client.ZRevRange(ctx, sessionListKey, 0, 99).Result()
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = sessionPrefix + id
	}
	data, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	var sessions []model.Session
	for _, d := range data {
		if d == nil {
			continue
		}
		var session model.Session
		if err := json.Unmarshal([]byte(d.(string)), &session); err != nil {
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

var addMessageScript = redis.NewScript(`
	local key = KEYS[1]
	local listKey = KEYS[2]
	local msgJson = ARGV[1]
	local sessionID = ARGV[2]
	local title = ARGV[3]
	local maxMsgs = tonumber(ARGV[4])
	local ttl = tonumber(ARGV[5])
	local now_ts = ARGV[6]  -- epoch seconds from Go
	local now_iso = ARGV[7] -- ISO timestamp from Go

	local data = redis.call('GET', key)
	local session
	if data then
		session = cjson.decode(data)
	else
		session = {
			id = sessionID,
			title = title,
			messages = {},
			created_at = now_iso
		}
	end

	local msg = cjson.decode(msgJson)
	msg.created_at = nil
	table.insert(session.messages, msg)

	if #session.messages > maxMsgs then
		local start = #session.messages - maxMsgs + 1
		local trimmed = {}
		for i = start, #session.messages do
			table.insert(trimmed, session.messages[i])
		end
		session.messages = trimmed
	end

	session.updated_at = now_iso
	redis.call('SET', key, cjson.encode(session), 'EX', ttl)
	redis.call('ZADD', listKey, now_ts, sessionID)
	return 'ok'
`)

func (s *RedisStore) AddMessage(ctx context.Context, sessionID string, msg model.Message) error {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	title := truncateText(msg.Content, 30)
	keys := []string{sessionPrefix + sessionID, sessionListKey}
	now := time.Now()

	return addMessageScript.Run(ctx, s.client, keys,
		string(msgJSON), sessionID, title,
		maxMessages, int64(defaultTTL.Seconds()),
		now.Unix(), now.Format("2006-01-02T15:04:05"),
	).Err()
}

func truncateText(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}
