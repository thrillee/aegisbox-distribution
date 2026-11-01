'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2, ToggleLeft, ToggleRight } from 'lucide-react'

// A labelled on/off switch that PATCHes { [field]: value } to endpoint, after a confirmation
// because these flags change routing and reporting.
export function FlagSwitch({
  endpoint,
  field,
  initial,
  label,
  description,
  confirmOn,
  confirmOff,
  onChange,
}: Readonly<{
  endpoint: string
  field: string
  initial: boolean
  label: string
  description: string
  confirmOn: string
  confirmOff: string
  onChange?: (value: boolean) => void
}>) {
  const router = useRouter()
  const [on, setOn] = useState(initial)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const toggle = async () => {
    const next = !on
    if (!window.confirm(next ? confirmOn : confirmOff)) return
    setSaving(true)
    setError(null)
    try {
      const res = await fetch(endpoint, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ [field]: next }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setError(data.error || 'Failed to save')
        return
      }
      setOn(next)
      onChange?.(next)
      router.refresh()
    } catch {
      setError('Failed to save')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex items-center justify-between gap-4 rounded-lg border border-white/10 px-4 py-3">
      <div>
        <p className="font-medium text-white">{label}</p>
        <p className="text-sm text-zinc-500">{description}</p>
        {error && <p className="text-sm text-red-400 mt-1">{error}</p>}
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={on}
        aria-label={label}
        onClick={toggle}
        disabled={saving}
        className="text-zinc-400 hover:text-white disabled:opacity-50 shrink-0"
      >
        {saving ? (
          <Loader2 className="h-8 w-8 animate-spin" />
        ) : on ? (
          <ToggleRight className="h-8 w-8 text-amber-400" />
        ) : (
          <ToggleLeft className="h-8 w-8" />
        )}
      </button>
    </div>
  )
}

export function SimulatorBadge() {
  return <span className="px-2 py-0.5 rounded-full text-xs font-semibold bg-amber-500/20 text-amber-300">SIMULATOR</span>
}

export function TestBadge() {
  return <span className="px-2 py-0.5 rounded-full text-xs font-semibold bg-sky-500/20 text-sky-300">TEST</span>
}
