import { cn } from '@/lib/utils'
import { ReactNode } from 'react'

interface PageHeaderProps {
  title: string
  description?: string
  actions?: ReactNode
  className?: string
}

export function PageHeader({
  title,
  description,
  actions,
  className,
}: PageHeaderProps) {
  return (
    <div className={cn('flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between', className)}>
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-white">{title}</h1>
        {description && (
          <p className="text-zinc-400 mt-1">{description}</p>
        )}
      </div>
      {actions && (
        <div className="flex items-center gap-3 mt-4 sm:mt-0">
          {actions}
        </div>
      )}
    </div>
  )
}
