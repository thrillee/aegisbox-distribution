import { Loader2 } from 'lucide-react'

export function LoadingSpinner({ className = '' }: { className?: string }) {
  return (
    <div className={`flex items-center justify-center ${className}`}>
      <Loader2 className="h-8 w-8 animate-spin text-zinc-400" />
    </div>
  )
}

export function PageLoadingSkeleton() {
  return (
    <div className="p-8 space-y-6">
      {/* Header skeleton */}
      <div className="space-y-2">
        <div className="h-8 w-64 bg-white/10 rounded-lg animate-pulse" />
        <div className="h-4 w-96 bg-white/5 rounded-lg animate-pulse" />
      </div>

      {/* Cards skeleton */}
      <div className="grid gap-4 md:grid-cols-3">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-32 bg-white/5 rounded-lg animate-pulse" />
        ))}
      </div>

      {/* Table skeleton */}
      <div className="bg-white/5 rounded-lg p-6 space-y-4">
        <div className="h-6 w-48 bg-white/10 rounded animate-pulse" />
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="h-16 bg-white/5 rounded animate-pulse" />
        ))}
      </div>
    </div>
  )
}

export function TableLoadingSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="h-6 w-48 bg-white/10 rounded animate-pulse" />
        <div className="h-10 w-32 bg-white/10 rounded animate-pulse" />
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3">
        <div className="h-10 w-64 bg-white/5 rounded animate-pulse" />
        <div className="h-10 w-48 bg-white/5 rounded animate-pulse" />
      </div>

      {/* Table rows */}
      <div className="space-y-2">
        {Array.from({ length: rows }).map((_, i) => (
          <div key={i} className="h-16 bg-white/5 rounded animate-pulse" />
        ))}
      </div>
    </div>
  )
}
