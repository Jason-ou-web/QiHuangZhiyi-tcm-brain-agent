import { useChat } from '../../hooks/useChat'
import { useAgent } from '../../hooks/useAgent'
import { useAgentStore } from '../../stores/agentStore'
import { ChatInput } from './ChatInput'
import { MessageList } from './MessageList'
import { AgentModeToggle } from './AgentModeToggle'
import { CitationPanel } from '../Citation/CitationPanel'
import { ReasoningSteps } from '../Agent/ReasoningSteps'
import { TaskProgressBar } from '../Agent/TaskProgressBar'
import { ToolInvocationPanel } from '../Agent/ToolInvocationPanel'

export function ChatContainer() {
  const { isStreaming: chatStreaming, send } = useChat()
  const { sendAgent, isStreaming: agentStreaming } = useAgent()
  const isAgentMode = useAgentStore((s) => s.isAgentMode)
  const stepsLen = useAgentStore((s) => s.steps.length)
  const steps = useAgentStore((s) => s.steps)
  const tasks = useAgentStore((s) => s.tasks)
  const toolCalls = useAgentStore((s) => s.toolCalls)

  const isStreaming = chatStreaming || agentStreaming

  const handleSend = (query: string) => {
    if (isStreaming) return
    if (isAgentMode) {
      sendAgent(query)
    } else {
      send(query, true)
    }
  }

  return (
    <div className="flex h-full">
      {/* Main chat area */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Mode toggle bar */}
        <div className="flex items-center justify-between px-4 py-1.5 border-b border-apricot-100 bg-apricot-50/80">
          <AgentModeToggle />
          {isAgentMode && stepsLen > 0 && (
            <span className="text-[10px] text-apricot-400 font-mono">
              {stepsLen} 步推理
            </span>
          )}
        </div>
        <MessageList />
        <ChatInput onSend={handleSend} disabled={isStreaming} />
      </div>

      {/* Right side panel */}
      <div className="hidden lg:flex lg:w-80 border-l border-apricot-200 bg-apricot-50/60 flex-col overflow-hidden">
        {isAgentMode ? (
          <div className="flex-1 overflow-y-auto p-4 space-y-4">
            <TaskProgressBar tasks={tasks} />
            <ReasoningSteps
              steps={steps}
              isActive={agentStreaming}
            />
            <ToolInvocationPanel toolCalls={toolCalls} />
          </div>
        ) : (
          <CitationPanel />
        )}
      </div>
    </div>
  )
}
