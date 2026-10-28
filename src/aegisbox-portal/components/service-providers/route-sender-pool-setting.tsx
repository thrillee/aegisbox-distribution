'use client'

import { useState, useEffect } from 'react'
import { Loader2, ToggleLeft, ToggleRight } from 'lucide-react'

export function RouteSenderPoolSetting({ providerId }: Readonly<{ providerId: number }>) {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [enabled, setEnabled] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [savedMessage, setSavedMessage] = useState<string | null>(null)

  const url = `/api/admin/service-providers/${providerId}/route-sender-pool`

  useEffect(() => {
    const load = async () => {
      try {
        const res = await fetch(url)
        const data = await res.json().catch(() => ({}))
        if (res.ok) {
          setEnabled(data.enabled)
        } else {
          setErrorMessage(data.error || 'Failed to load the sender pool setting')
        }
      } catch {
        setErrorMessage('Failed to load the sender pool setting')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [url])

  const toggle = async () => {
    const next = !enabled
    if (next && !confirm(
      "Messages from this provider will go out with the route's pool senders instead of their own sender IDs, and their own sender IDs will no longer be checked. Continue?"
    )) {
      return
    }
    setErrorMessage(null)
    setSavedMessage(null)
    setSaving(true)
    try {
      const res = await fetch(url, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled: next }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setErrorMessage(data.error || 'Failed to save the sender pool setting')
        return
      }
      setEnabled(data.enabled)
      setSavedMessage(data.enabled ? 'Sender pool turned on' : 'Sender pool turned off')
    } catch {
      setErrorMessage('Failed to save the sender pool setting')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-zinc-500" />
      </div>
    )
  }

  return (
    <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
      <h3 className="text-lg font-semibold text-white">Sender Pool</h3>
      <p className="text-sm text-zinc-400 mt-0.5 mb-6">
        When on, each message goes out with a sender from the pool of the connection it is sent through, set on each
        MNO connection&apos;s page. If a route fails over, the sender switches to the new route&apos;s pool. The
        provider&apos;s own sender is kept on the message for reports, and billing and delivery reports are unchanged.
      </p>

      <div className="flex items-center justify-between rounded-lg border border-white/10 px-4 py-3">
        <div>
          <p className="font-medium text-white">Use route sender pool</p>
          <p className="text-sm text-zinc-500">
            {enabled ? 'On: messages use the route senders' : "Off: messages use the provider's approved sender IDs"}
          </p>
        </div>
        <button
          type="button"
          role="switch"
          aria-checked={enabled}
          aria-label="Use route sender pool"
          disabled={saving}
          onClick={toggle}
          className="text-zinc-400 hover:text-white disabled:opacity-50"
        >
          {enabled ? <ToggleRight className="h-8 w-8 text-emerald-400" /> : <ToggleLeft className="h-8 w-8" />}
        </button>
      </div>

      {errorMessage && <p className="text-sm text-red-400 mt-4">{errorMessage}</p>}
      {savedMessage && <p className="text-sm text-emerald-400 mt-4">{savedMessage}</p>}
    </div>
  )
}
