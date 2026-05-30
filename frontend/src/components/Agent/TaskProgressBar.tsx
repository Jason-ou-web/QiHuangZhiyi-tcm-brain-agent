import type { AgentTask } from '../../types/agent'

interface Props {
  tasks: AgentTask[]
}

const statusIcon: Record<string, string> = {
  pending: '⏳',
  running: '🔄',
  done: '✅',
  failed: '❌',
}

const statusColor: Record<string, string> = {
  pending: 'text-gray-400',
  running: 'text-amber-600',
  done: 'text-green-600',
  failed: 'text-red-500',
}

export function TaskProgressBar({ tasks }: Props) {
  if (tasks.length === 0) return null

  const doneCount = tasks.filter((t) => t.status === 'done').length

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-xs text-gray-500">
        <span>任务进度</span>
        <span>
          {doneCount}/{tasks.length}
        </span>
      </div>
      {/* Progress bar */}
      <div className="w-full bg-gray-200 rounded-full h-1.5">
        <div
          className="bg-amber-500 h-1.5 rounded-full transition-all duration-500"
          style={{ width: `${tasks.length > 0 ? (doneCount / tasks.length) * 100 : 0}%` }}
        />
      </div>
      {/* Task list */}
      <div className="space-y-1">
        {tasks.map((task) => (
          <div key={task.id} className="flex items-center gap-2 text-xs">
            <span>{statusIcon[task.status] || '⏳'}</span>
            <span className={`${statusColor[task.status] || 'text-gray-500'} truncate`}>
              {task.title}
            </span>
            {task.status === 'running' && (
              <span className="ml-auto w-2 h-2 bg-amber-500 rounded-full animate-pulse" />
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
