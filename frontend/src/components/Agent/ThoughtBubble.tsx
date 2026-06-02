interface Props {
  content: string
  isActive?: boolean
}

export function ThoughtBubble({ content, isActive }: Props) {
  if (!content?.trim()) return null

  return (
    <div
      className={`flex items-start gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
        isActive
          ? 'bg-amber-50 border border-amber-200'
          : 'bg-gray-50 border border-gray-100'
      }`}
    >
      <span className="shrink-0 mt-0.5">{isActive ? <span className="inline-block w-2 h-2 rounded-full bg-amber-400 animate-pulse mr-1" /> : null}💭</span>
      <p className="text-gray-600 italic break-words">{content}</p>
    </div>
  )
}
