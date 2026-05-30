export interface Message {
  role: 'user' | 'assistant' | 'system'
  content: string
}

export interface Citation {
  book_title: string
  chapter?: string
  content: string
  page?: number
  score: number
}

export interface Session {
  id: string
  title: string
  messages: Message[]
  created_at: string
  updated_at: string
}

export interface ChatRequest {
  session_id?: string
  query: string
}

export interface ChatResponse {
  session_id: string
  answer: string
  citations: Citation[]
}

export interface SSEEvent {
  type: 'token' | 'citations' | 'meta' | 'done' | 'error'
  content: string
  data?: unknown
}

export interface ChatState {
  sessionId: string | null
  messages: Message[]
  citations: Citation[]
  isStreaming: boolean
}
