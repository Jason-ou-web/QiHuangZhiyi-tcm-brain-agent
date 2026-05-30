package tcm

import (
	"context"
	"fmt"

	"agri-qa-system/internal/agent/tools"
	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
)

type WellnessTool struct{}

func NewWellnessTool() tools.Tool { return &WellnessTool{} }

func (t *WellnessTool) Name() string { return "wellness_suggestion" }

func (t *WellnessTool) Description() string {
	return "根据中医体质辨识和季节时令，提供养生保健建议，包括饮食调理、起居作息、情志调养、运动导引等。"
}

func (t *WellnessTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"condition": map[string]any{"type": "string", "description": "体质或证型，如阴虚体质、痰湿体质等（可选）"},
			"season":    map[string]any{"type": "string", "description": "季节，如春、夏、秋、冬（可选）"},
			"focus":     map[string]any{"type": "string", "description": "养生侧重点：饮食、起居、情志、运动（可选）"},
		},
		"required": []string{},
	}
}

func (t *WellnessTool) Execute(ctx context.Context, args map[string]any, ragPipe *rag.Pipeline) (*model.ToolResult, error) {
	condition, _ := args["condition"].(string)
	season, _ := args["season"].(string)
	focus, _ := args["focus"].(string)

	query := "中医养生"
	if condition != "" {
		query = fmt.Sprintf("%s 中医调理 养生建议", condition)
	}
	if season != "" {
		query += " " + season + "季养生"
	}
	if focus != "" {
		query += " " + focus
	}

	return tools.RetrieveKnowledge(ctx, ragPipe, query, map[string]any{
		"condition": condition,
		"season":    season,
		"focus":     focus,
	})
}
