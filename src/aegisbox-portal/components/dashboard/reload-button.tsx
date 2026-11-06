'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface ReloadButtonProps {
  /** Re-fetch for client pages. Without it, the button re-renders the server page with router.refresh(). */
  onReload?: () => unknown
  /** The page's own loading flag, for when its fetch is not the promise onReload returns. */
  loading?: boolean
  label?: string
  className?: string
}

/** Reloads a page's data in place, without a full browser reload. */
export function ReloadButton({ onReload, loading = false, label = 'Reload', className }: ReloadButtonProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [running, setRunning] = useState(false)
  const busy = loading || isPending || running

  const handleClick = async () => {
    if (onReload) {
      setRunning(true)
      try {
        await onReload()
      } finally {
        setRunning(false)
      }
      return
    }
    startTransition(() => router.refresh())
  }

  return (
    <Button
      type="button"
      variant="outline"
      onClick={handleClick}
      disabled={busy}
      title="Reload data"
      className={cn('border-white/10 text-zinc-400 hover:text-white hover:bg-white/5', className)}
    >
      <RefreshCw className={cn('h-4 w-4', busy && 'animate-spin')} />
      {label}
    </Button>
  )
}
