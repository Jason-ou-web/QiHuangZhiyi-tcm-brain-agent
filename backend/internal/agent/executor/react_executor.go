package executor

import (
	"context"
	"fmt"
	"strings"

	"agri-qa-system/internal/agent/tools"
	"agri-qa-system/internal/llm"
	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
)

const agentSystemPrompt = `你是一个专业的中医智能助手，基于已收集的中医知识为用户生成最终答案。

要求：
1. 综合所有步骤收集到的信息，给出完整的调理方案
2. 方案结构清晰，包含：辨证分析、药材/方剂推荐、养生建议、注意事项
3. 使用 [参考N] 标注知识来源
4. 涉及诊断、用药建议时，务必提醒用户仅供参考，建议咨询专业中医师
5. 使用通俗易懂的中文，专业术语需做简要解释`

// ReactExecutor runs the ReAct loop: for each task, think -> act -> observe
type ReactExecutor struct {
	registry *tools.Registry
	ragPipe  *rag.Pipeline
	llm      *llm.Client
}

func NewReactExecutor(registry *tools.Registry, ragPipe *rag.Pipeline, llmClient *llm.Client) *ReactExecutor {
	return &ReactExecutor{registry: registry, ragPipe: ragPipe, llm: llmClient}
}

// Run executes all planned tasks and streams reasoning events, then synthesizes the final answer.
func (e *ReactExecutor) Run(ctx context.Context, tasks []model.AgentTask, query string, history []model.Message, eventCh chan<- model.AgentEvent) error {
	var toolResults []model.ToolCall

	for i, task := range tasks {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip the final synthesis task - handled after the loop
		if task.Title == model.SynthesisTaskTitle {
			continue
		}

		stepNum := i + 1

		// 1. THOUGHT
		thought := e.generateThought(task, query, toolResults)
		emitEvent(eventCh, model.AgentEvent{
			Type: "thought",
			Data: map[string]any{
				"step_number": stepNum,
				"task_id":     task.ID,
				"task_title":  task.Title,
				"content":     thought,
			},
		})

		// 2. ACTION - execute the tool
		task.Status = "running"
		emitEvent(eventCh, model.AgentEvent{
			Type: "task_progress",
			Data: map[string]any{
				"task_id":    task.ID,
				"task_title": task.Title,
				"status":     "running",
			},
		})

		args := e.buildArgs(task, query, toolResults)
		emitEvent(eventCh, model.AgentEvent{
			Type: "tool_call",
			Data: map[string]any{
				"step_number": stepNum,
				"tool_name":   task.ToolName,
				"args":        args,
			},
		})

		result, err := e.registry.Execute(ctx, task.ToolName, args, e.ragPipe)
		if err != nil {
			task.Status = "failed"
			emitEvent(eventCh, model.AgentEvent{
				Type:    "tool_result",
				Content: err.Error(),
				Data:    map[string]any{"step_number": stepNum, "success": false},
			})
			continue
		}

		toolCall := model.ToolCall{Name: task.ToolName, Args: args, Result: result}
		toolResults = append(toolResults, toolCall)

		// 3. OBSERVATION
		observation := formatObservation(result)
		emitEvent(eventCh, model.AgentEvent{
			Type: "tool_result",
			Data: map[string]any{
				"step_number": stepNum,
				"tool_name":   task.ToolName,
				"success":     result.Success,
				"observation": observation,
				"references":  getRefCount(result),
			},
		})

		task.Status = "done"
		emitEvent(eventCh, model.AgentEvent{
			Type: "task_progress",
			Data: map[string]any{
				"task_id":    task.ID,
				"task_title": task.Title,
				"status":     "done",
			},
		})
	}

	// 4. FINAL ANSWER - synthesize from all tool results
	answer, err := e.synthesize(ctx, query, toolResults, history)
	if err != nil {
		return fmt.Errorf("synthesize: %w", err)
	}

	emitEvent(eventCh, model.AgentEvent{
		Type:    "final_answer",
		Content: answer,
	})
	return nil
}

func (e *ReactExecutor) generateThought(task model.AgentTask, query string, prevResults []model.ToolCall) string {
	if len(prevResults) == 0 {
		return fmt.Sprintf("首先需要分析用户问题「%s」，调用%s工具获取相关信息。", task.Description, task.ToolName)
	}
	return fmt.Sprintf("基于前面的分析结果，接下来需要%s，进一步收集信息。", task.Description)
}

func (e *ReactExecutor) buildArgs(task model.AgentTask, query string, prevResults []model.ToolCall) map[string]any {
	args := map[string]any{}
	switch task.ToolName {
	case "diagnose_symptom":
		args["symptoms"] = query
	case "herb_query":
		args["condition"] = query
	case "prescription_advice":
		args["symptoms"] = query
	case "wellness_suggestion":
		args["condition"] = query
	case "acupoint_info":
		args["condition"] = query
	}
	return args
}

func (e *ReactExecutor) synthesize(ctx context.Context, query string, toolResults []model.ToolCall, history []model.Message) (string, error) {
	var contextParts []string
	for i, tc := range toolResults {
		if tc.Result != nil && tc.Result.Success {
			data, _ := tc.Result.Data.(map[string]any)
			if knowledge, ok := data["knowledge"].(string); ok {
				contextParts = append(contextParts, fmt.Sprintf("[工具%d: %s]\n%s", i+1, tc.Name, knowledge))
			}
		}
	}

	messages := []model.Message{
		{Role: "system", Content: agentSystemPrompt},
	}
	if len(contextParts) > 0 {
		messages = append(messages, model.Message{
			Role:    "system",
			Content: fmt.Sprintf("已收集的中医知识：\n%s", strings.Join(contextParts, "\n\n")),
		})
	}
	messages = append(messages, history...)
	messages = append(messages, model.Message{Role: "user", Content: "请综合以上信息，为用户问题「" + query + "」生成完整的调理方案。"})

	return e.llm.Chat(messages)
}

func emitEvent(ch chan<- model.AgentEvent, event model.AgentEvent) {
	select {
	case ch <- event:
	default:
	}
}

func formatObservation(result *model.ToolResult) string {
	if !result.Success {
		return result.Error
	}
	data, ok := result.Data.(map[string]any)
	if !ok {
		return "获取到相关信息"
	}
	refs, _ := data["references"].(int)
	return fmt.Sprintf("检索到 %d 条相关中医知识", refs)
}

func getRefCount(result *model.ToolResult) int {
	if data, ok := result.Data.(map[string]any); ok {
		if refs, ok := data["references"].(int); ok {
			return refs
		}
	}
	return 0
}
