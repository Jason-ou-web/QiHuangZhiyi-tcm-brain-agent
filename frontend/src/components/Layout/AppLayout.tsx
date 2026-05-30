import { useState } from 'react'
import { SessionSidebar } from '../Session/SessionSidebar'
import { ChatContainer } from '../Chat/ChatContainer'

export function AppLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(true)

  return (
    <div className="h-screen flex overflow-hidden bg-apricot-50">
      {/* Mobile overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/30 z-20 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}
      {/* Sidebar */}
      <div
        className={`fixed lg:static inset-y-0 left-0 z-30 transform transition-transform
          ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'}
          lg:translate-x-0`}
      >
        <SessionSidebar />
      </div>
      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Top bar */}
        <div className="h-12 border-b border-apricot-200 flex items-center px-4 gap-3 bg-white/80 backdrop-blur-sm">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="lg:hidden text-apricot-500 hover:text-apricot-600"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <img
            src="/avatars/qihuang.png"
            alt="岐黄"
            className="w-7 h-7 rounded-full object-cover shadow-sm"
          />
          <h1 className="font-semibold text-apricot-600 text-sm">岐黄智医</h1>
        </div>
        {/* Chat area */}
        <ChatContainer />
      </div>
    </div>
  )
}
