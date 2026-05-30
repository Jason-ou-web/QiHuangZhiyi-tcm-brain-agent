import { useAgentStore } from '../../stores/agentStore'

export function AgentModeToggle() {
  const { isAgentMode, toggleAgentMode } = useAgentStore()

  return (
    <button
      onClick={toggleAgentMode}
      className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium transition-all ${
        isAgentMode
          ? 'bg-apricot-200 text-apricot-700 border border-apricot-300'
          : 'bg-white text-apricot-500 border border-apricot-200 hover:bg-apricot-50'
      }`}
      title={isAgentMode ? 'Agent模式：使用多步推理分析' : 'RAG模式：直接检索回答'}
    >
      <span>{isAgentMode ? '🧠' : '📖'}</span>
      <span>{isAgentMode ? 'Agent 模式' : 'RAG 模式'}</span>
    </button>
  )
}
