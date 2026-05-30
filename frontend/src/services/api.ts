import type { ChatResponse, Session } from '../types'

const BASE_URL = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'request failed' }))
    throw new Error(err.error || 'request failed')
  }
  return res.json()
}

export async function sendChat(query: string, sessionId?: string): Promise<ChatResponse> {
  return request<ChatResponse>('/chat', {
    method: 'POST',
    body: JSON.stringify({ query, session_id: sessionId }),
  })
}

export async function getSessions(): Promise<{ sessions: Session[] }> {
  return request('/sessions')
}

export async function getSession(id: string): Promise<{ session: Session }> {
  return request(`/sessions/${id}`)
}

export async function deleteSession(id: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/sessions/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'request failed' }))
    throw new Error(err.error || 'request failed')
  }
}
