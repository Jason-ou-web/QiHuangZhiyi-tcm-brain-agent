package tools

import (
	"context"

	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
)

// Tool is the interface all TCM tools must implement.
// Each tool reuses the RAG Pipeline for knowledge retrieval.
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any // JSON Schema for tool parameters
	Execute(ctx context.Context, args map[string]any, ragPipe *rag.Pipeline) (*model.ToolResult, error)
}
