import { create } from 'zustand'
import type { Message, Citation } from '../types'

interface ChatStore {
  sessionId: string | null
  messages: Message[]
  citations: Citation[]
  isStreaming: boolean
  streamingContent: string

  setSessionId: (id: string) => void
  addMessage: (msg: Message) => void
  setMessages: (msgs: Message[]) => void
  setCitations: (cits: Citation[]) => void
  setIsStreaming: (v: boolean) => void
  setStreamingContent: (content: string) => void
  appendStreamToken: (token: string) => void
  resetStream: () => void
  clearChat: () => void
}

export const useChatStore = create<ChatStore>((set) => ({
  sessionId: null,
  messages: [],
  citations: [],
  isStreaming: false,
  streamingContent: '',

  setSessionId: (id) => set({ sessionId: id }),
  addMessage: (msg) => set((s) => ({ messages: [...s.messages, msg] })),
  setMessages: (msgs) => set({ messages: msgs }),
  setCitations: (cits) => set({ citations: cits }),
  setIsStreaming: (v) => set({ isStreaming: v }),
  setStreamingContent: (content) => set({ streamingContent: content }),
  appendStreamToken: (token) =>
    set((s) => ({ streamingContent: s.streamingContent + token })),
  resetStream: () => set({ isStreaming: false, streamingContent: '' }),
  clearChat: () =>
    set({
      sessionId: null,
      messages: [],
      citations: [],
      isStreaming: false,
      streamingContent: '',
    }),
}))
