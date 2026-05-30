import { create } from 'zustand'
import type { ReasoningStep, AgentTask, ToolCall } from '../types/agent'

interface AgentStore {
  isAgentMode: boolean
  isStreaming: boolean
  steps: ReasoningStep[]
  tasks: AgentTask[]
  toolCalls: ToolCall[]
  finalAnswer: string
  error: string | null

  setAgentMode: (v: boolean) => void
  toggleAgentMode: () => void
  setIsStreaming: (v: boolean) => void
  addStep: (step: ReasoningStep) => void
  updateStep: (stepNum: number, updates: Partial<ReasoningStep>) => void
  setTasks: (tasks: AgentTask[]) => void
  updateTask: (taskId: string, updates: Partial<AgentTask>) => void
  addToolCall: (call: ToolCall) => void
  updateToolResult: (callId: string, result: ToolCall['result']) => void
  setFinalAnswer: (answer: string) => void
  setError: (err: string | null) => void
  resetAgent: () => void
}

export const useAgentStore = create<AgentStore>((set) => ({
  isAgentMode: false,
  isStreaming: false,
  steps: [],
  tasks: [],
  toolCalls: [],
  finalAnswer: '',
  error: null,

  setAgentMode: (v) => set({ isAgentMode: v }),
  toggleAgentMode: () => set((s) => ({ isAgentMode: !s.isAgentMode })),
  setIsStreaming: (v) => set({ isStreaming: v }),

  addStep: (step) => set((s) => ({ steps: [...s.steps, step] })),
  updateStep: (stepNum, updates) =>
    set((s) => ({
      steps: s.steps.map((st) =>
        st.step_number === stepNum ? { ...st, ...updates } : st
      ),
    })),

  setTasks: (tasks) => set({ tasks }),
  updateTask: (taskId, updates) =>
    set((s) => ({
      tasks: s.tasks.map((t) => (t.id === taskId ? { ...t, ...updates } : t)),
    })),

  addToolCall: (call) => set((s) => ({ toolCalls: [...s.toolCalls, call] })),
  updateToolResult: (callId, result) =>
    set((s) => ({
      toolCalls: s.toolCalls.map((tc) =>
        tc.call_id === callId ? { ...tc, result } : tc
      ),
    })),

  setFinalAnswer: (answer) => set({ finalAnswer: answer }),
  setError: (err) => set({ error: err }),
  resetAgent: () =>
    set({
      isStreaming: false,
      steps: [],
      tasks: [],
      toolCalls: [],
      finalAnswer: '',
      error: null,
    }),
}))
