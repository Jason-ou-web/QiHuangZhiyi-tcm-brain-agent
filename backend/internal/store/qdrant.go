package store

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"agri-qa-system/config"
	"agri-qa-system/internal/model"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type QdrantStore struct {
	conn       *grpc.ClientConn
	client     pb.PointsClient
	collection string
}

func NewQdrantStore(cfg *config.Config) (*QdrantStore, error) {
	port, err := strconv.Atoi(cfg.QdrantPort)
	if err != nil {
		return nil, fmt.Errorf("invalid QDRANT_PORT %q: %w", cfg.QdrantPort, err)
	}
	addr := fmt.Sprintf("%s:%d", cfg.QdrantHost, port)
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect qdrant: %w", err)
	}
	return &QdrantStore{
		conn:       conn,
		client:     pb.NewPointsClient(conn),
		collection: cfg.QdrantCollection,
	}, nil
}

func (s *QdrantStore) Close() error {
	return s.conn.Close()
}

func (s *QdrantStore) Search(ctx context.Context, vector []float32, topK int) ([]model.SearchResult, error) {
	req := &pb.SearchPoints{
		CollectionName: s.collection,
		Vector:         vector,
		Limit:          uint64(topK),
		WithPayload:    &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	}

	resp, err := s.client.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("qdrant search: %w", err)
	}

	log.Printf("[Qdrant] Search collection=%s got %d results", s.collection, len(resp.GetResult()))

	var results []model.SearchResult
	for _, point := range resp.GetResult() {
		payload := point.GetPayload()
		results = append(results, model.SearchResult{
			ID:      point.GetId().GetUuid(),
			Content: payload["content"].GetStringValue(),
			Score:   float64(point.GetScore()),
			Metadata: map[string]any{
				"book_title": payload["book_title"].GetStringValue(),
				"chapter":    payload["chapter"].GetStringValue(),
				"page":       payload["page"].GetIntegerValue(),
			},
		})
	}
	return results, nil
}
