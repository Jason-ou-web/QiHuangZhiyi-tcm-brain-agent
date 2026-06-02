import { useState } from 'react'
import type { ToolCall } from '../../types/agent'

interface Props {
  toolCalls: ToolCall[]
}

export function ToolInvocationPanel({ toolCalls }: Props) {
  if (toolCalls.length === 0) return null

  return (
    <div className="space-y-2">
      <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wide">
        工具调用 ({toolCalls.length})
      </h4>
      {toolCalls.map((call, i) => (
        <ToolCallItem key={call.call_id || i} call={call} index={i} />
      ))}
    </div>
  )
}

function ToolCallItem({ call, index }: { call: ToolCall; index: number }) {
  const [expanded, setExpanded] = useState(false)

  return (
    <div className="bg-white border rounded-lg text-xs">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between px-3 py-2 hover:bg-gray-50 transition-colors"
      >
        <div className="flex items-center gap-2">
          <span className="inline-block w-5 h-5 rounded bg-amber-100 text-amber-700 text-center leading-5 font-mono text-[10px]">
            {index + 1}
          </span>
          <span className="font-medium text-gray-700">{call.name}</span>
          {call.result && (
            <span
              className={`inline-block w-1.5 h-1.5 rounded-full ${
                call.result.success ? 'bg-green-500' : 'bg-red-500'
              }`}
            />
          )}
        </div>
        <span className="text-gray-400 text-[10px]">{expanded ? '收起' : '详情'}</span>
      </button>
      {expanded && (
        <div className="border-t px-3 py-2 space-y-1.5 bg-gray-50/50">
          <div>
            <span className="text-gray-500">参数：</span>
            <code className="text-gray-700 bg-gray-100 px-1 py-0.5 rounded">
              {JSON.stringify(call.args, null, 2)}
            </code>
          </div>
          {call.result != null && (
            <div>
              <span className="text-gray-500">结果：</span>
              <span className={call.result.success ? 'text-green-700' : 'text-red-600'}>
                {call.result.success
                  ? '执行成功'
                  : call.result.error || '执行失败'}
              </span>
              {call.result.data != null && typeof call.result.data === 'object' && (
                <pre className="mt-1 text-[10px] text-gray-600 bg-gray-100 p-1.5 rounded overflow-x-auto max-h-32">
                  {JSON.stringify(call.result.data as Record<string, unknown>, null, 2)}
                </pre>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
