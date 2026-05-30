interface Props {
  content: string
  isActive?: boolean
}

export function ThoughtBubble({ content, isActive }: Props) {
  if (!content) return null

  return (
    <div
      className={`flex items-start gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
        isActive
          ? 'bg-amber-50 border border-amber-200 animate-pulse'
          : 'bg-gray-50 border border-gray-100'
      }`}
    >
      <span className="shrink-0 mt-0.5">💭</span>
      <p className="text-gray-600 italic">{content}</p>
    </div>
  )
}
