import type { AgentEvent } from '../types/agent'

export function createAgentStream(
  query: string,
  sessionId: string | null,
  onEvent: (event: AgentEvent) => void,
  onError: (error: string) => void,
  onDone: () => void
): AbortController {
  const controller = new AbortController()
  let finished = false

  const safeOnDone = () => {
    if (!finished) {
      finished = true
      onDone()
    }
  }

  fetch('/api/v1/agent/chat/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query, session_id: sessionId }),
    signal: controller.signal,
  })
    .then(async (response) => {
      if (!response.ok) {
        const err = await response.json().catch(() => ({ error: 'stream failed' }))
        onError(err.error || 'stream failed')
        safeOnDone()
        return
      }

      const reader = response.body?.getReader()
      if (!reader) {
        onError('no response body')
        safeOnDone()
        return
      }

      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          const trimmed = line.trim()
          if (!trimmed.startsWith('data: ')) continue
          const data = trimmed.slice(6)
          try {
            const event: AgentEvent = JSON.parse(data)
            onEvent(event)
            if (event.type === 'done' || event.type === 'error') {
              reader.cancel().catch(() => {})
              safeOnDone()
              return
            }
          } catch {
            // skip invalid JSON
          }
        }
      }
      safeOnDone()
    })
    .catch((err) => {
      if (err.name !== 'AbortError') {
        onError(err.message || 'connection failed')
      }
      safeOnDone()
    })

  return controller
}

export async function analyzeQuery(query: string): Promise<{ needs_agent: boolean; reason: string }> {
  const res = await fetch('/api/v1/agent/analyze', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query }),
  })
  if (!res.ok) {
    return { needs_agent: false, reason: '分析失败，默认使用RAG模式' }
  }
  return res.json()
}
