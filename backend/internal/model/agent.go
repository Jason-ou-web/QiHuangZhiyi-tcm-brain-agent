package model

// SynthesisTaskTitle is the reserved title for the final answer synthesis task
const SynthesisTaskTitle = "整合答案"

// AgentTask represents a planned subtask
type AgentTask struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ToolName    string `json:"tool_name,omitempty"`
	Status      string `json:"status"` // pending/running/done/failed
}

// ToolCall represents a tool invocation
type ToolCall struct {
	Name   string         `json:"name"`
	Args   map[string]any `json:"args"`
	Result *ToolResult    `json:"result,omitempty"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Error   string `json:"error,omitempty"`
}

// AgentEvent extends SSEEvent for agent reasoning steps
// New event types: thought, task_progress, tool_call, tool_result, final_answer
type AgentEvent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Data    any    `json:"data,omitempty"`
}

