import { cn } from '@/lib/utils'
import { ReactNode } from 'react'
import Link from 'next/link'

interface DataCardProps {
  title: string
  subtitle?: string
  children: ReactNode
  action?: {
    label: string
    href: string
  }
  className?: string
}

export function DataCard({
  title,
  subtitle,
  children,
  action,
  className,
}: DataCardProps) {
  return (
    <div
      className={cn(
        'rounded-2xl border border-white/10 bg-zinc-900/50 p-6',
        className
      )}
    >
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-semibold text-white">{title}</h3>
          {subtitle && <p className="text-sm text-zinc-400 mt-0.5">{subtitle}</p>}
        </div>
        {action && (
          <Link
            href={action.href}
            className="text-sm font-medium text-zinc-400 hover:text-white transition-colors"
          >
            {action.label}
          </Link>
        )}
      </div>
      {children}
    </div>
  )
}
