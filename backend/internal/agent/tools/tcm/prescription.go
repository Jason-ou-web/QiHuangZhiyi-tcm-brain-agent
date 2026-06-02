package tcm

import (
	"context"
	"fmt"

	"agri-qa-system/internal/agent/tools"
	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
)

type PrescriptionTool struct{}

func NewPrescriptionTool() tools.Tool { return &PrescriptionTool{} }

func (t *PrescriptionTool) Name() string { return "prescription_advice" }

func (t *PrescriptionTool) Description() string {
	return "根据证型或症状推荐经典方剂，包括方剂组成、功效、主治、加减化裁等。参考伤寒论、金匮要略等经典。"
}

func (t *PrescriptionTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"prescription_name": map[string]any{"type": "string", "description": "方剂名称，如四君子汤、六味地黄丸等（可选）"},
			"pattern":           map[string]any{"type": "string", "description": "证型，如脾胃气虚、肾阴虚等（可选）"},
			"symptoms":          map[string]any{"type": "string", "description": "具体症状描述（可选）"},
		},
		"required": []string{},
	}
}

func (t *PrescriptionTool) Execute(ctx context.Context, args map[string]any, ragPipe *rag.Pipeline) (*model.ToolResult, error) {
	prescriptionName, _ := args["prescription_name"].(string)
	pattern, _ := args["pattern"].(string)
	symptoms, _ := args["symptoms"].(string)

	query := "中医方剂"
	if prescriptionName != "" {
		query = fmt.Sprintf("方剂 %s 组成 功效 主治 加减", prescriptionName)
	}
	if pattern != "" {
		query += fmt.Sprintf(" %s 经典方剂 推荐", pattern)
	}
	if symptoms != "" {
		query += fmt.Sprintf(" 针对%s", symptoms)
	}

	return tools.RetrieveKnowledge(ctx, ragPipe, query, map[string]any{
		"prescription_name": prescriptionName,
		"pattern":           pattern,
		"symptoms":          symptoms,
	})
}
