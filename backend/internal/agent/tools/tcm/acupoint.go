package tcm

import (
	"context"
	"fmt"

	"agri-qa-system/internal/agent/tools"
	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
)

type AcupointTool struct{}

func NewAcupointTool() tools.Tool { return &AcupointTool{} }

func (t *AcupointTool) Name() string { return "acupoint_info" }

func (t *AcupointTool) Description() string {
	return "查询穴位信息，包括穴位定位、所属经络、功效主治、按摩方法、艾灸注意事项等。"
}

func (t *AcupointTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"acupoint_name": map[string]any{"type": "string", "description": "穴位名称，如足三里、合谷、太冲等（可选）"},
			"meridian":      map[string]any{"type": "string", "description": "经络名称，如足阳明胃经、手太阴肺经等（可选）"},
			"condition":     map[string]any{"type": "string", "description": "需要调理的症状，如头痛、失眠、胃痛等（可选）"},
		},
		"required": []string{},
	}
}

func (t *AcupointTool) Execute(ctx context.Context, args map[string]any, ragPipe *rag.Pipeline) (*model.ToolResult, error) {
	acupointName, _ := args["acupoint_name"].(string)
	meridian, _ := args["meridian"].(string)
	condition, _ := args["condition"].(string)

	query := "中医穴位"
	if acupointName != "" {
		query = fmt.Sprintf("穴位 %s 定位 功效 按摩", acupointName)
	}
	if meridian != "" {
		query += fmt.Sprintf(" %s 穴位 定位 功效", meridian)
	}
	if condition != "" {
		query += fmt.Sprintf(" %s 对应穴位 按摩调理", condition)
	}

	return tools.RetrieveKnowledge(ctx, ragPipe, query, map[string]any{
		"acupoint_name": acupointName,
		"meridian":      meridian,
		"condition":     condition,
	})
}
