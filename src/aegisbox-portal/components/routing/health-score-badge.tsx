import * as React from 'react'
import { cn } from '@/lib/utils'

interface HealthScoreBadgeProps {
  score: number | string | null | undefined
  className?: string
}

export function HealthScoreBadge({ score, className }: HealthScoreBadgeProps) {
  // Convert string to number if needed
  const numericScore = typeof score === 'string' ? parseFloat(score) : score

  if (numericScore === null || numericScore === undefined || isNaN(numericScore)) {
    return (
      <span className={cn('inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-zinc-500/20 text-zinc-400', className)}>
        N/A
      </span>
    )
  }

  // Scores are stored as percentages (0-100)
  const percentage = Math.round(numericScore)

  // Determine color based on health score
  let colorClasses = ''
  if (numericScore > 70) {
    colorClasses = 'bg-emerald-500/20 text-emerald-400'
  } else if (numericScore > 50) {
    colorClasses = 'bg-amber-500/20 text-amber-400'
  } else {
    colorClasses = 'bg-red-500/20 text-red-400'
  }

  return (
    <span className={cn('inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', colorClasses, className)}>
      {percentage}%
    </span>
  )
}
