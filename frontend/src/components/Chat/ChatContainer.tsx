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
  const agentStore = useAgentStore()

  const isAgentMode = agentStore.isAgentMode
  const isStreaming = isAgentMode ? agentStreaming : chatStreaming

  const handleSend = (query: string) => {
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
          {isAgentMode && agentStore.steps.length > 0 && (
            <span className="text-[10px] text-apricot-400 font-mono">
              {agentStore.steps.length} 步推理
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
            <TaskProgressBar tasks={agentStore.tasks} />
            <ReasoningSteps
              steps={agentStore.steps}
              isActive={agentStreaming}
            />
            <ToolInvocationPanel toolCalls={agentStore.toolCalls} />
          </div>
        ) : (
          <CitationPanel />
        )}
      </div>
    </div>
  )
}
