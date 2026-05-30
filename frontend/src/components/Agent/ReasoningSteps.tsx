import type { ReasoningStep } from '../../types/agent'
import { ThoughtBubble } from './ThoughtBubble'

interface Props {
  steps: ReasoningStep[]
  isActive?: boolean
}

export function ReasoningSteps({ steps, isActive }: Props) {
  if (steps.length === 0) return null

  return (
    <div className="space-y-3">
      <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wide">
        推理过程
      </h4>
      <div className="relative pl-6 border-l-2 border-amber-200 space-y-4">
        {steps.map((step) => (
          <div key={step.step_number} className="relative">
            {/* Timeline dot */}
            <div
              className={`absolute -left-[25px] w-3 h-3 rounded-full border-2 border-white ${
                step.status === 'thinking'
                  ? 'bg-amber-400'
                  : step.status === 'acting'
                  ? 'bg-blue-400 animate-pulse'
                  : step.status === 'observing'
                  ? 'bg-green-400'
                  : 'bg-gray-300'
              }`}
            />
            {/* Step content */}
            <div className="space-y-1.5">
              <span className="text-[10px] text-gray-400 font-mono">
                Step {step.step_number}
              </span>
              <ThoughtBubble
                content={step.thought}
                isActive={step.status === 'thinking'}
              />
              {step.action && (
                <div className="text-xs text-blue-600 bg-blue-50 rounded px-2 py-1">
                  🔧 调用工具：{step.action.name}
                </div>
              )}
              {step.observation && (
                <div className="text-xs text-green-700 bg-green-50 rounded px-2 py-1">
                  📋 {step.observation}
                </div>
              )}
            </div>
          </div>
        ))}
        {/* Pending indicator */}
        {isActive && (
          <div className="relative">
            <div className="absolute -left-[25px] w-3 h-3 rounded-full border-2 border-white bg-gray-300 animate-pulse" />
            <div className="flex items-center gap-1 text-xs text-gray-400 py-1">
              <span>思考中</span>
              <span className="animate-bounce">...</span>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
