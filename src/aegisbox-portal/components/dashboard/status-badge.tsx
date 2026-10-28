import { cn } from '@/lib/utils'
import { CheckCircle2, Clock, AlertCircle } from 'lucide-react'

type Status = 'delivered' | 'pending' | 'failed' | 'unknown'

interface StatusBadgeProps {
  status: Status
  showIcon?: boolean
  className?: string
}

const statusConfig: Record<Status, { bg: string; text: string; icon: typeof CheckCircle2; label: string }> = {
  delivered: {
    bg: 'bg-emerald-500/20 text-emerald-400',
    text: 'Delivered',
    icon: CheckCircle2,
    label: 'delivered',
  },
  pending: {
    bg: 'bg-amber-500/20 text-amber-400',
    text: 'Pending',
    icon: Clock,
    label: 'pending',
  },
  failed: {
    bg: 'bg-red-500/20 text-red-400',
    text: 'Failed',
    icon: AlertCircle,
    label: 'failed',
  },
  unknown: {
    bg: 'bg-zinc-500/20 text-zinc-400',
    text: 'Unknown',
    icon: AlertCircle,
    label: 'unknown',
  },
}

export function StatusBadge({ status, showIcon = true, className }: StatusBadgeProps) {
  const config = statusConfig[status] || statusConfig.unknown
  const Icon = config.icon

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium',
        config.bg,
        className
      )}
    >
      {showIcon && <Icon className="h-3 w-3" />}
      {config.text}
    </span>
  )
}
