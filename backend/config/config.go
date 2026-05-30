package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort    string
	DeepSeekAPIKey string
	DeepSeekBaseURL string
	DeepSeekModel  string
	QdrantHost     string
	QdrantPort     string
	QdrantCollection string
	RedisAddr      string
	RedisPassword  string
	JWTSecret      string
	EmbeddingURL   string
	RerankerURL    string
}

func Load() (*Config, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required, generate one with: openssl rand -hex 32")
	}

	return &Config{
		ServerPort:       envOrDefault("SERVER_PORT", "8080"),
		DeepSeekAPIKey:   apiKey,
		DeepSeekBaseURL:  envOrDefault("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepSeekModel:    envOrDefault("DEEPSEEK_MODEL", "deepseek-chat"),
		QdrantHost:       envOrDefault("QDRANT_HOST", "localhost"),
		QdrantPort:       envOrDefault("QDRANT_PORT", "6334"),
		QdrantCollection: envOrDefault("QDRANT_COLLECTION", "tcm_knowledge"),
		RedisAddr:        envOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
		JWTSecret:        jwtSecret,
		EmbeddingURL:     envOrDefault("EMBEDDING_URL", "http://localhost:8081/embed"),
		RerankerURL:      envOrDefault("RERANKER_URL", "http://localhost:8082/rerank"),
	}, nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
