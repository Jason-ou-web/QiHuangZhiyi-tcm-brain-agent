import { useCallback, useRef } from 'react'
import { useChatStore } from '../stores/chatStore'
import { useSessionStore } from '../stores/sessionStore'
import { sendChat } from '../services/api'
import { createSSEStream } from '../services/sse'
import type { Message } from '../types'

export function useChat() {
  const {
    sessionId, setSessionId,
    messages, addMessage, setMessages,
    citations, setCitations,
    isStreaming, setIsStreaming,
    streamingContent, appendStreamToken, resetStream,
    clearChat,
  } = useChatStore()
  const { addSession } = useSessionStore()
  const abortRef = useRef<AbortController | null>(null)

  const stopStream = useCallback(() => {
    if (abortRef.current) {
      abortRef.current.abort()
      abortRef.current = null
    }
  }, [])

  const send = useCallback(async (query: string, useStream = true) => {
    if (!query.trim() || isStreaming) return

    // abort any existing stream
    stopStream()

    const userMsg: Message = { role: 'user', content: query }
    addMessage(userMsg)

    if (useStream) {
      setIsStreaming(true)
      let newSessionId = sessionId
      let fullAnswer = ''
      const seenCitations: typeof citations = []

      abortRef.current = createSSEStream(
        query,
        sessionId,
        (event) => {
          switch (event.type) {
            case 'meta': {
              const data = event.data as { session_id?: string } | undefined
              if (data?.session_id) {
                newSessionId = data.session_id
                setSessionId(data.session_id)
              }
              break
            }
            case 'citations': {
              const cits = event.data as typeof citations | undefined
              if (cits && Array.isArray(cits)) {
                seenCitations.push(...cits)
                setCitations([...seenCitations])
              }
              break
            }
            case 'token':
              fullAnswer += event.content
              appendStreamToken(event.content)
              break
            case 'error':
              addMessage({ role: 'assistant', content: `错误: ${event.content}` })
              break
          }
        },
        (error) => {
          addMessage({ role: 'assistant', content: `连接失败: ${error}` })
        },
        () => {
          abortRef.current = null
          if (fullAnswer) {
            addMessage({ role: 'assistant', content: fullAnswer })
          }
          if (newSessionId) {
            addSession({
              id: newSessionId,
              title: query.slice(0, 30),
              messages: [...messages, userMsg, { role: 'assistant', content: fullAnswer }],
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            })
          }
          resetStream()
        }
      )
    } else {
      try {
        const resp = await sendChat(query, sessionId || undefined)
        setSessionId(resp.session_id)
        setCitations(resp.citations)
        addMessage({ role: 'assistant', content: resp.answer })
        addSession({
          id: resp.session_id,
          title: query.slice(0, 30),
          messages: [...messages, userMsg, { role: 'assistant', content: resp.answer }],
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        })
      } catch (err) {
        addMessage({ role: 'assistant', content: `请求失败: ${err instanceof Error ? err.message : 'unknown'}` })
      }
    }
  }, [sessionId, messages, isStreaming, citations])

  return {
    sessionId, messages, citations, isStreaming, streamingContent,
    send, stopStream, clearChat, setMessages,
  }
}
