package tcm

import (
	"context"
	"fmt"

	"agri-qa-system/internal/agent/tools"
	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
)

type HerbTool struct{}

func NewHerbTool() tools.Tool { return &HerbTool{} }

func (t *HerbTool) Name() string { return "herb_query" }

func (t *HerbTool) Description() string {
	return "查询中药材信息，包括性味归经、功效主治、用法用量、配伍禁忌等。可根据证型或症状推荐相关药材。"
}

func (t *HerbTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"herb_name": map[string]any{"type": "string", "description": "药材名称，如柴胡、当归、黄芪等（可选）"},
			"condition": map[string]any{"type": "string", "description": "需要调理的证型或症状，如肝郁气滞、气血两虚等（可选）"},
			"property":  map[string]any{"type": "string", "description": "药性筛选，如温、寒、平（可选）"},
		},
		"required": []string{},
	}
}

func (t *HerbTool) Execute(ctx context.Context, args map[string]any, ragPipe *rag.Pipeline) (*model.ToolResult, error) {
	herbName, _ := args["herb_name"].(string)
	condition, _ := args["condition"].(string)
	property, _ := args["property"].(string)

	query := "中药材"
	if herbName != "" {
		query = fmt.Sprintf("中药材 %s 性味归经 功效主治 用法", herbName)
	} else if condition != "" {
		query = fmt.Sprintf("中医 %s 药材推荐 配伍", condition)
	}
	if property != "" {
		query += " " + property + "性"
	}

	return tools.RetrieveKnowledge(ctx, ragPipe, query, map[string]any{
		"herb_name": herbName,
		"condition": condition,
		"property":  property,
	})
}
