import * as React from 'react'
import { cn } from '@/lib/utils'
import { CheckCircle, AlertTriangle, XCircle } from 'lucide-react'

interface RouteStatusBadgeProps {
  isAvailable: boolean
  consecutiveFailures?: number
  className?: string
}

export function RouteStatusBadge({ isAvailable, consecutiveFailures = 0, className }: RouteStatusBadgeProps) {
  if (!isAvailable) {
    return (
      <span className={cn('inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-500/20 text-red-400', className)}>
        <XCircle className="h-3 w-3" />
        Unavailable
      </span>
    )
  }

  if (consecutiveFailures >= 3) {
    return (
      <span className={cn('inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/20 text-amber-400', className)}>
        <AlertTriangle className="h-3 w-3" />
        Degraded
      </span>
    )
  }

  return (
    <span className={cn('inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/20 text-emerald-400', className)}>
      <CheckCircle className="h-3 w-3" />
      Available
    </span>
  )
}
