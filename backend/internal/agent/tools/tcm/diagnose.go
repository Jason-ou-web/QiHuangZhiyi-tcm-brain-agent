package tcm

import (
	"context"
	"fmt"

	"agri-qa-system/internal/agent/tools"
	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
)

type DiagnoseTool struct{}

func NewDiagnoseTool() tools.Tool { return &DiagnoseTool{} }

func (t *DiagnoseTool) Name() string { return "diagnose_symptom" }

func (t *DiagnoseTool) Description() string {
	return "根据用户描述的症状进行中医辨证分析，结合古籍知识给出可能的证型判断。返回症状对应的病因病机、证型分类和鉴别要点。"
}

func (t *DiagnoseTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symptoms":  map[string]any{"type": "string", "description": "用户描述的症状，如头痛、失眠、食欲不振等"},
			"body_area": map[string]any{"type": "string", "description": "症状所在部位，如头部、腹部、四肢等（可选）"},
			"duration":  map[string]any{"type": "string", "description": "症状持续时间（可选）"},
		},
		"required": []string{"symptoms"},
	}
}

func (t *DiagnoseTool) Execute(ctx context.Context, args map[string]any, ragPipe *rag.Pipeline) (*model.ToolResult, error) {
	symptoms, _ := args["symptoms"].(string)
	if symptoms == "" {
		return &model.ToolResult{Success: false, Error: "缺少症状描述"}, nil
	}

	bodyArea, _ := args["body_area"].(string)
	duration, _ := args["duration"].(string)

	query := fmt.Sprintf("中医辨证 症状分析 %s", symptoms)
	if bodyArea != "" {
		query += " " + bodyArea
	}
	if duration != "" {
		query += " " + duration
	}

	return tools.RetrieveKnowledge(ctx, ragPipe, query, map[string]any{
		"symptoms":  symptoms,
		"body_area": bodyArea,
		"duration":  duration,
	})
}
