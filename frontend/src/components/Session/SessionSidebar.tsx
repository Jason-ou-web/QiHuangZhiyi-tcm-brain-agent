import { useSessions } from '../../hooks/useSessions'

export function SessionSidebar() {
  const { sessions, activeId, isLoading, selectSession, newChat, deleteSession } = useSessions()

  return (
    <div className="w-64 bg-apricot-100/70 border-r border-apricot-200 flex flex-col h-full">
      <div className="p-3">
        <button
          onClick={newChat}
          className="w-full py-2 px-4 rounded-lg border border-apricot-400 text-apricot-600
                     hover:bg-apricot-200 text-sm font-medium transition-colors"
        >
          + 新对话
        </button>
      </div>
      <div className="flex-1 overflow-y-auto px-2">
        {isLoading && (
          <p className="text-xs text-apricot-400 text-center py-4">加载中...</p>
        )}
        {sessions.length === 0 && !isLoading && (
          <p className="text-xs text-apricot-400 text-center py-4">暂无历史对话</p>
        )}
        {sessions.map((s) => (
          <div
            key={s.id}
            className={`group flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer mb-1 text-sm
              ${activeId === s.id
                ? 'bg-apricot-200 text-apricot-700'
                : 'hover:bg-apricot-200/60 text-apricot-600'
              }`}
            onClick={() => selectSession(s.id)}
          >
            <span className="flex-1 truncate">{s.title || '新对话'}</span>
            <button
              onClick={(e) => {
                e.stopPropagation()
                deleteSession(s.id)
              }}
              className="hidden group-hover:block text-apricot-400 hover:text-red-500 shrink-0"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        ))}
      </div>
      <div className="p-3 border-t border-apricot-200 text-xs text-apricot-400">
        中医智能助手 v0.1.0
      </div>
    </div>
  )
}
