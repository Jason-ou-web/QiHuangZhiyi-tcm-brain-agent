package tools

import (
	"context"
	"fmt"
	"strings"

	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
)

// RetrieveKnowledge searches the RAG pipeline with the given query and returns
// a ToolResult with formatted knowledge and reference count.
func RetrieveKnowledge(ctx context.Context, ragPipe *rag.Pipeline, query string, extraData map[string]any) (*model.ToolResult, error) {
	results, err := ragPipe.Retrieve(ctx, query, 5)
	if err != nil {
		return &model.ToolResult{Success: false, Error: fmt.Sprintf("知识检索失败: %v", err)}, nil
	}

	var parts []string
	for i, r := range results {
		parts = append(parts, fmt.Sprintf("[参考%d] %s", i+1, r.Content))
	}

	data := map[string]any{
		"references": len(results),
		"knowledge":  strings.Join(parts, "\n\n"),
	}
	for k, v := range extraData {
		data[k] = v
	}
	return &model.ToolResult{Success: true, Data: data}, nil
}
