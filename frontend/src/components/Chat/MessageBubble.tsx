import type { Message } from '../../types'
import { MarkdownRenderer } from '../Markdown/MarkdownRenderer'

function AssistantAvatar() {
  return (
    <img
      src="/avatars/qihuang.png"
      alt="岐黄"
      className="w-8 h-8 rounded-full shrink-0 object-cover shadow-sm"
    />
  )
}

function UserAvatar() {
  return (
    <img
      src="/avatars/user.png"
      alt="用户"
      className="w-8 h-8 rounded-full shrink-0 object-cover shadow-sm"
    />
  )
}

interface Props {
  message: Message
  isStreaming?: boolean
}

export function MessageBubble({ message, isStreaming }: Props) {
  const isUser = message.role === 'user'

  return (
    <div className={`flex items-start gap-2.5 ${isUser ? 'flex-row-reverse' : 'flex-row'}`}>
      {isUser ? <UserAvatar /> : <AssistantAvatar />}
      <div
        className={`max-w-[80%] rounded-2xl px-4 py-3 ${
          isUser
            ? 'bg-primary-500 text-white rounded-br-md'
            : 'bg-white text-gray-800 rounded-bl-md border border-apricot-100 shadow-sm'
        }`}
      >
        {isUser ? (
          <p className="whitespace-pre-wrap text-sm">{message.content}</p>
        ) : (
          <div className="text-sm prose prose-sm max-w-none">
            <MarkdownRenderer content={message.content} />
          </div>
        )}
        {isStreaming && (
          <span className="inline-block w-2 h-4 bg-apricot-400 animate-pulse ml-0.5 align-text-bottom rounded-sm" />
        )}
      </div>
    </div>
  )
}
