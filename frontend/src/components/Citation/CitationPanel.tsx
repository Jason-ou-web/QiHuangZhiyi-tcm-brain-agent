import { useChatStore } from '../../stores/chatStore'
import { CitationCard } from './CitationCard'

export function CitationPanel() {
  const citations = useChatStore((s) => s.citations)

  if (citations.length === 0) return null

  return (
    <div className="hidden lg:block w-72 border-l bg-gray-50 p-4 overflow-y-auto">
      <h3 className="text-sm font-semibold text-gray-700 mb-3">
        引用来源 ({citations.length})
      </h3>
      <div className="space-y-3">
        {citations.map((cit, i) => (
          <CitationCard key={`${i}-${cit.book_title || 'unknown'}-${cit.score}`} citation={cit} index={i + 1} />
        ))}
      </div>
    </div>
  )
}
