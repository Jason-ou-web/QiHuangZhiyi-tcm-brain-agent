export interface AgentTask {
  id: string
  title: string
  description: string
  tool_name?: string
  status: 'pending' | 'running' | 'done' | 'failed'
}

export interface ToolCall {
  call_id: string
  name: string
  args: Record<string, unknown>
  result?: ToolResult
}

export interface ToolResult {
  success: boolean
  data: unknown
  error?: string
}

export interface ReasoningStep {
  step_number: number
  thought: string
  action?: ToolCall
  observation?: string
  status: 'thinking' | 'acting' | 'observing' | 'done'
}

export interface AgentEvent {
  type: 'meta' | 'thought' | 'task_progress' | 'tool_call' | 'tool_result' | 'final_answer' | 'error'
  content: string
  data?: unknown
}

export interface AgentState {
  isAgentMode: boolean
  isStreaming: boolean
  steps: ReasoningStep[]
  tasks: AgentTask[]
  toolCalls: ToolCall[]
  finalAnswer: string
  error: string | null
}

export interface AgentAnalyzeResult {
  needs_agent: boolean
  reason: string
}
