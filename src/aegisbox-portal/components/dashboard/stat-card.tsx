import { cn } from '@/lib/utils'
import { LucideIcon, TrendingUp, TrendingDown, Minus } from 'lucide-react'
import { ReactElement } from 'react'

interface StatCardProps {
  title: string
  value: string | number
  icon: LucideIcon
  trend?: 'up' | 'down' | null
  trendValue?: string
  prefix?: string
  suffix?: string
  className?: string
}

export function StatCard({
  title,
  value,
  icon: Icon,
  trend,
  trendValue,
  prefix = '',
  suffix = '',
  className,
}: StatCardProps) {
  const TrendIcon = trend === 'up' ? TrendingUp : trend === 'down' ? TrendingDown : Minus

  return (
    <div
      className={cn(
        'rounded-2xl border border-white/10 bg-zinc-900/50 p-6',
        className
      )}
    >
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium text-zinc-400">{title}</p>
        <Icon className="h-5 w-5 text-zinc-400" />
      </div>
      <div className="mt-2">
        <p className="text-3xl font-bold text-white">
          {prefix}
          {typeof value === 'number' ? value.toLocaleString() : value}
          {suffix}
        </p>
        {trend && trendValue && (
          <div className="flex items-center mt-1">
            <TrendIcon
              className={cn(
                'h-4 w-4',
                trend === 'up' ? 'text-emerald-500' : trend === 'down' ? 'text-red-500' : 'text-zinc-500'
              )}
            />
            <span
              className={cn(
                'text-sm',
                trend === 'up' ? 'text-emerald-500' : trend === 'down' ? 'text-red-500' : 'text-zinc-500'
              )}
            >
              {trendValue}
            </span>
            <span className="text-sm text-zinc-500 ml-1">vs last week</span>
          </div>
        )}
      </div>
    </div>
  )
}
