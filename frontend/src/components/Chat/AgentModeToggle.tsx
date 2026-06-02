import { useAgentStore } from '../../stores/agentStore'

export function AgentModeToggle() {
  const isAgentMode = useAgentStore((s) => s.isAgentMode)
  const toggleAgentMode = useAgentStore((s) => s.toggleAgentMode)

  return (
    <button
      onClick={toggleAgentMode}
      aria-pressed={isAgentMode}
      aria-label={isAgentMode ? '当前Agent模式，点击切换到RAG模式' : '当前RAG模式，点击切换到Agent模式'}
      className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium transition-all ${
        isAgentMode
          ? 'bg-apricot-200 text-apricot-700 border border-apricot-300'
          : 'bg-white text-apricot-500 border border-apricot-200 hover:bg-apricot-50'
      }`}
    >
      <span aria-hidden="true">{isAgentMode ? '🧠' : '📖'}</span>
      <span>{isAgentMode ? 'Agent 模式' : 'RAG 模式'}</span>
    </button>
  )
}
