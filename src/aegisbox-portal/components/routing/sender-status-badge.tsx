import * as React from 'react'
import { cn } from '@/lib/utils'
import { CheckCircle, AlertTriangle, Ban } from 'lucide-react'

interface SenderStatusBadgeProps {
  isBlocked: boolean
  healthScore?: number | string | null
  deliveryRate?: number | string | null
  className?: string
}

export function SenderStatusBadge({ isBlocked, healthScore, deliveryRate, className }: SenderStatusBadgeProps) {
  if (isBlocked) {
    return (
      <span className={cn('inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-500/20 text-red-400', className)}>
        <Ban className="h-3 w-3" />
        Blocked
      </span>
    )
  }

  // Use deliveryRate if healthScore is not provided
  const scoreToUse = healthScore !== undefined && healthScore !== null ? healthScore : deliveryRate

  // Convert string to number if needed
  const numericScore = typeof scoreToUse === 'string' ? parseFloat(scoreToUse) : scoreToUse

  if (numericScore === null || numericScore === undefined || isNaN(numericScore)) {
    return (
      <span className={cn('inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-zinc-500/20 text-zinc-400', className)}>
        Unknown
      </span>
    )
  }

  // Scores are stored as percentages (0-100)
  if (numericScore < 50) {
    return (
      <span className={cn('inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/20 text-amber-400', className)}>
        <AlertTriangle className="h-3 w-3" />
        Warning
      </span>
    )
  }

  return (
    <span className={cn('inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/20 text-emerald-400', className)}>
      <CheckCircle className="h-3 w-3" />
      Healthy
    </span>
  )
}
