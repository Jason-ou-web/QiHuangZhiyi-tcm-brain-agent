package planner

import (
	"context"
	"fmt"
	"strings"

	"agri-qa-system/internal/agent/tools"
	"agri-qa-system/internal/model"
)

// TaskPlanner decomposes a user query into subtasks with tool assignments
type TaskPlanner struct {
	registry *tools.Registry
}

func NewTaskPlanner(registry *tools.Registry) *TaskPlanner {
	return &TaskPlanner{registry: registry}
}

// Plan decomposes the query into ordered subtasks.
// For MVP, uses keyword-based planning. Future: LLM-driven planning.
func (p *TaskPlanner) Plan(ctx context.Context, query string) []model.AgentTask {
	var tasks []model.AgentTask
	id := 0

	// Helper to add a task
	addTask := func(title, desc, tool string) {
		tasks = append(tasks, model.AgentTask{
			ID: fmtID(id), Title: title,
			Description: desc, ToolName: tool, Status: "pending",
		})
		id++
	}

	// Check which aspects the query covers and assign tools accordingly
	if containsAny(query, []string{"症状", "诊断", "辨证", "不舒服", "疼痛", "头痛", "失眠", "乏力", "食欲", "消化"}) {
		addTask("症状辨证", "根据用户描述的症状进行中医辨证分析", "diagnose_symptom")
	}

	if containsAny(query, []string{"药", "药材", "中药", "草药", "推荐药", "用什么药"}) {
		addTask("药材查询", "查询相关中药材的性味归经和功效", "herb_query")
	}

	if containsAny(query, []string{"方剂", "汤药", "汤剂", "丸剂", "散剂", "处方", "经方"}) {
		addTask("方剂推荐", "根据证型推荐经典方剂", "prescription_advice")
	}

	if containsAny(query, []string{"养生", "调理", "保健", "饮食", "作息", "运动", "情志"}) {
		addTask("养生建议", "提供中医养生保健建议", "wellness_suggestion")
	}

	if containsAny(query, []string{"穴位", "针灸", "按摩", "艾灸", "经络", "足三里", "合谷"}) {
		addTask("穴位指导", "查询相关穴位信息与按摩方法", "acupoint_info")
	}

	// If no specific domain matched, create a general diagnosis task
	if len(tasks) == 0 {
		addTask("综合分析", "综合查询相关中医知识", "diagnose_symptom")
	}

	// Always add a final synthesis task
	tasks = append(tasks, model.AgentTask{
		ID: fmtID(id), Title: model.SynthesisTaskTitle,
		Description: "整合所有步骤的结果，生成完整的调理方案",
		Status:      "pending",
	})

	return tasks
}

func containsAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func fmtID(n int) string { return fmt.Sprintf("task_%d", n) }
