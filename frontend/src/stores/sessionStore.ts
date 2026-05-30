import { create } from 'zustand'
import type { Session } from '../types'

interface SessionStore {
  sessions: Session[]
  activeId: string | null
  setSessions: (sessions: Session[]) => void
  setActiveId: (id: string | null) => void
  addSession: (session: Session) => void
  removeSession: (id: string) => void
}

export const useSessionStore = create<SessionStore>((set) => ({
  sessions: [],
  activeId: null,
  setSessions: (sessions) => set({ sessions }),
  setActiveId: (id) => set({ activeId: id }),
  addSession: (session) =>
    set((s) => ({
      sessions: [session, ...s.sessions.filter((x) => x.id !== session.id)],
    })),
  removeSession: (id) =>
    set((s) => ({
      sessions: s.sessions.filter((x) => x.id !== id),
      activeId: s.activeId === id ? null : s.activeId,
    })),
}))
