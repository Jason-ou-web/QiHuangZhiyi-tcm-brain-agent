import { useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getSessions, deleteSession as apiDeleteSession } from '../services/api'
import { useSessionStore } from '../stores/sessionStore'
import { useChatStore } from '../stores/chatStore'

export function useSessions() {
  const { sessions, activeId, setSessions, setActiveId, removeSession } = useSessionStore()
  const { clearChat } = useChatStore()

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['sessions'],
    queryFn: getSessions,
    refetchOnWindowFocus: false,
  })

  useEffect(() => {
    if (data?.sessions) {
      setSessions(data.sessions)
    }
  }, [data, setSessions])

  const selectSession = (id: string) => {
    setActiveId(id)
  }

  const newChat = () => {
    setActiveId(null)
    clearChat()
  }

  const handleDelete = async (id: string) => {
    try {
      await apiDeleteSession(id)
      removeSession(id)
      if (activeId === id) {
        newChat()
      }
    } catch {
      // silently fail
    }
  }

  return {
    sessions, activeId, isLoading,
    selectSession, newChat, deleteSession: handleDelete, refetch,
  }
}
