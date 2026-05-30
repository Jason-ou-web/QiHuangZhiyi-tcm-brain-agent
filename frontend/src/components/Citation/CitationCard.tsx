import { useState } from 'react'
import type { Citation } from '../../types'

interface Props {
  citation: Citation
  index: number
}

export function CitationCard({ citation, index }: Props) {
  const [expanded, setExpanded] = useState(false)

  return (
    <div className="bg-white rounded-lg border p-3 text-sm">
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <span className="inline-block bg-primary-100 text-primary-700 text-xs px-1.5 py-0.5 rounded mr-2">
            [{index}]
          </span>
          <span className="font-medium text-gray-800">{citation.book_title}</span>
          {citation.chapter && (
            <span className="text-gray-500 ml-1">· {citation.chapter}</span>
          )}
        </div>
        <button
          onClick={() => setExpanded(!expanded)}
          className="shrink-0 text-gray-400 hover:text-gray-600 ml-2"
        >
          {expanded ? '收起' : '展开'}
        </button>
      </div>
      {expanded && (
        <blockquote className="mt-2 border-l-2 border-primary-300 pl-3 text-gray-600 text-xs leading-relaxed">
          {citation.content}
        </blockquote>
      )}
      {citation.score > 0 && (
        <div className="mt-1 text-xs text-gray-400">
          相关度: {(citation.score * 100).toFixed(0)}%
        </div>
      )}
    </div>
  )
}
