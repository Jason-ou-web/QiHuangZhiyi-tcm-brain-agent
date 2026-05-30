import { useRef, useEffect } from 'react'
import { useChatStore } from '../../stores/chatStore'
import { MessageBubble } from './MessageBubble'

export function MessageList() {
  const messages = useChatStore((s) => s.messages)
  const isStreaming = useChatStore((s) => s.isStreaming)
  const streamingContent = useChatStore((s) => s.streamingContent)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, streamingContent])

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-4">
      {messages.length === 0 && !isStreaming && (
        <div className="flex items-center justify-center h-full text-apricot-400">
          <div className="text-center">
            <div className="text-6xl mb-4">🌿</div>
            <p className="text-lg font-medium text-apricot-500">欢迎使用中医智能助手</p>
            <p className="text-sm mt-2 text-apricot-400">输入您想了解的中医问题，获取专业解答</p>
          </div>
        </div>
      )}
      {messages.map((msg, i) => (
        <MessageBubble key={i} message={msg} />
      ))}
      {isStreaming && streamingContent && (
        <MessageBubble
          message={{ role: 'assistant', content: streamingContent }}
          isStreaming
        />
      )}
      {isStreaming && !streamingContent && (
        <div className="flex items-center gap-1.5 text-apricot-300 px-2">
          <div className="w-2 h-2 bg-apricot-300 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
          <div className="w-2 h-2 bg-apricot-300 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
          <div className="w-2 h-2 bg-apricot-300 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
        </div>
      )}
      <div ref={bottomRef} />
    </div>
  )
}
