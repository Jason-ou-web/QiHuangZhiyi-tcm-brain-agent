import { useState, useRef, useEffect, useCallback } from 'react'

interface Props {
  onSend: (query: string) => void
  disabled?: boolean
}

export function ChatInput({ onSend, disabled }: Props) {
  const [input, setInput] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const composingRef = useRef(false)

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = Math.min(textareaRef.current.scrollHeight, 150) + 'px'
    }
  }, [input])

  const handleSubmit = useCallback(() => {
    if (!input.trim() || disabled) return
    onSend(input.trim())
    setInput('')
  }, [input, disabled, onSend])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey && !composingRef.current) {
      e.preventDefault()
      handleSubmit()
    }
  }

  return (
    <div className="border-t border-apricot-100 bg-white/80 backdrop-blur-sm p-4">
      <div className="max-w-3xl mx-auto flex gap-3 items-end">
        <textarea
          ref={textareaRef}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          onCompositionStart={() => { composingRef.current = true }}
          onCompositionEnd={() => { composingRef.current = false }}
          placeholder="输入中医相关问题，如：肝肾阴虚如何调理？"
          rows={1}
          disabled={disabled}
          aria-label="输入中医问题"
          className="flex-1 resize-none rounded-xl border border-apricot-200 px-4 py-3 text-sm
                     focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent
                     disabled:bg-apricot-50 disabled:text-apricot-300 placeholder:text-apricot-300"
        />
        <button
          onClick={handleSubmit}
          disabled={disabled || !input.trim()}
          className="shrink-0 w-10 h-10 rounded-xl bg-primary-500 text-white
                     hover:bg-primary-600 disabled:bg-apricot-200 disabled:cursor-not-allowed
                     flex items-center justify-center transition-colors"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14m-7-7l7 7-7 7" />
          </svg>
        </button>
      </div>
      <p className="text-xs text-apricot-400 text-center mt-2">
        按 Enter 发送，Shift+Enter 换行 | 回答基于中医知识库，请核实关键信息
      </p>
    </div>
  )
}
