import { useCallback, useRef } from 'react'
import { useAgentStore } from '../stores/agentStore'
import { useChatStore } from '../stores/chatStore'
import { useSessionStore } from '../stores/sessionStore'
import { createAgentStream, analyzeQuery } from '../services/agent'
import type { Message } from '../types'
import type { AgentEvent, AgentTask, ToolCall } from '../types/agent'

export function useAgent() {
  const {
    isAgentMode, setAgentMode, toggleAgentMode,
    isStreaming, setIsStreaming,
    steps, addStep, updateStep,
    tasks, setTasks, updateTask,
    toolCalls, addToolCall, updateToolResult,
    finalAnswer, setFinalAnswer,
    error, setError, resetAgent,
  } = useAgentStore()

  const { sessionId, setSessionId, addMessage, messages } = useChatStore()
  const { addSession } = useSessionStore()
  const abortRef = useRef<AbortController | null>(null)

  const stopStream = useCallback(() => {
    if (abortRef.current) {
      abortRef.current.abort()
      abortRef.current = null
    }
  }, [])

  const sendAgent = useCallback(async (query: string) => {
    if (!query.trim() || isStreaming) return

    stopStream()
    resetAgent()

    const userMsg: Message = { role: 'user', content: query }
    addMessage(userMsg)

    setIsStreaming(true)
    let newSessionId = sessionId
    let currentStep = 0
    let lastCallId = ''

    abortRef.current = createAgentStream(
      query,
      sessionId,
      (event: AgentEvent) => {
        switch (event.type) {
          case 'meta': {
            const data = event.data as { session_id?: string } | undefined
            if (data?.session_id) {
              newSessionId = data.session_id
              setSessionId(data.session_id)
            }
            break
          }
          case 'task_progress': {
            const data = event.data as { tasks?: AgentTask[]; task_id?: string; status?: string } | undefined
            if (data?.tasks && Array.isArray(data.tasks)) {
              setTasks(data.tasks)
            } else if (data?.task_id) {
              updateTask(data.task_id, { status: data.status as AgentTask['status'] })
            }
            break
          }
          case 'thought': {
            const data = event.data as { step_number?: number; task_title?: string; content?: string } | undefined
            if (data) {
              currentStep = data.step_number ?? currentStep
              addStep({
                step_number: currentStep,
                thought: data.content || '',
                status: 'thinking',
              })
            }
            break
          }
          case 'tool_call': {
            const data = event.data as { tool_name?: string; args?: Record<string, unknown> } | undefined
            if (data) {
              lastCallId = `${data.tool_name || 'tool'}_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`
              const call: ToolCall = { call_id: lastCallId, name: data.tool_name || '', args: data.args || {} }
              addToolCall(call)
              updateStep(currentStep, { status: 'acting', action: call })
            }
            break
          }
          case 'tool_result': {
            const data = event.data as { success?: boolean; observation?: string; tool_name?: string } | undefined
            if (data) {
              updateToolResult(lastCallId, {
                success: data.success || false,
                data: { observation: data.observation, tool_name: data.tool_name },
              })
              updateStep(currentStep, {
                status: 'observing',
                observation: data.observation,
              })
            }
            break
          }
          case 'final_answer':
            setFinalAnswer(event.content)
            break
          case 'error':
            setError(event.content || '未知错误')
            break
        }
      },
      (err) => {
        setError(err)
        addMessage({ role: 'assistant', content: `Agent执行失败: ${err}` })
      },
      () => {
        abortRef.current = null
        setIsStreaming(false)
        // Read latest state directly from store (not closure)
        const state = useAgentStore.getState()
        const answer = state.finalAnswer || (state.steps.length > 0 ? '分析完成' : '')
        if (answer) {
          addMessage({ role: 'assistant', content: answer })
          if (newSessionId) {
            addSession({
              id: newSessionId,
              title: query.slice(0, 30),
              messages: [...useChatStore.getState().messages, userMsg, { role: 'assistant', content: answer }],
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            })
          }
        }
      }
    )
  }, [sessionId, messages, isStreaming])

  // Auto-analyze and suggest Agent mode
  const checkQuery = useCallback(async (query: string): Promise<boolean> => {
    try {
      const result = await analyzeQuery(query)
      if (result.needs_agent) {
        setAgentMode(true)
        return true
      }
      return false
    } catch {
      return false
    }
  }, [setAgentMode])

  return {
    isAgentMode, setAgentMode, toggleAgentMode,
    isStreaming, steps, tasks, toolCalls, finalAnswer, error,
    sendAgent, stopStream, resetAgent, checkQuery,
  }
}
