'use client'

import { useState, useEffect, useCallback, useId } from 'react'
import { Loader2, Plus, Trash2, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface RouteSender {
  id: number
  sender_id: string
  status: 'active' | 'inactive'
  position: number
  usage_count: number
  last_used_at: string | null
}

interface RouteSenderPool {
  strategy: string
  senders: RouteSender[]
}

export const SENDER_POOL_STRATEGIES = [
  { value: 'round_robin', label: 'Round robin', description: 'Senders take turns: each send uses the one used least recently.' },
  { value: 'random', label: 'Random', description: 'Each send uses any active sender at random.' },
  { value: 'ordered', label: 'Ordered', description: 'Always the first active sender by position; the others are used only when it is turned off.' },
]

export function RouteSenderPool({ connectionId }: Readonly<{ connectionId: number }>) {
  const ids = useId()
  const [loading, setLoading] = useState(true)
  const [pool, setPool] = useState<RouteSenderPool | null>(null)
  const [strategy, setStrategy] = useState('round_robin')
  const [newSender, setNewSender] = useState('')
  const [newPosition, setNewPosition] = useState('0')
  const [busy, setBusy] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [savedMessage, setSavedMessage] = useState<string | null>(null)

  const base = `/api/admin/mno-connections/${connectionId}/route-senders`

  const fetchPool = useCallback(async () => {
    try {
      const res = await fetch(base)
      const data = await res.json().catch(() => ({}))
      if (res.ok) {
        setPool(data)
        setStrategy(data.strategy)
      } else {
        setErrorMessage(data.error || 'Failed to load the route sender pool')
      }
    } catch {
      setErrorMessage('Failed to load the route sender pool')
    } finally {
      setLoading(false)
    }
  }, [base])

  useEffect(() => {
    fetchPool()
  }, [fetchPool])

  const run = async (action: () => Promise<Response>, success: string, onSuccess?: () => void) => {
    setErrorMessage(null)
    setSavedMessage(null)
    setBusy(true)
    try {
      const res = await action()
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        setErrorMessage(data.error || 'Request failed')
        return
      }
      onSuccess?.()
      await fetchPool()
      setSavedMessage(success)
    } catch {
      setErrorMessage('Request failed')
    } finally {
      setBusy(false)
    }
  }

  const json = (method: string, body: unknown) => ({
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })

  const saveStrategy = () =>
    run(() => fetch(`${base}/strategy`, json('PUT', { strategy })), 'Strategy saved')

  const addSender = async (e: React.FormEvent) => {
    e.preventDefault()
    const position = Number(newPosition)
    if (!Number.isInteger(position)) {
      setErrorMessage('Position must be a whole number')
      return
    }
    await run(
      () => fetch(base, json('POST', { sender_id: newSender.trim(), position })),
      `${newSender.trim()} added`,
      () => {
        setNewSender('')
        setNewPosition('0')
      }
    )
  }

  const updateSender = (s: RouteSender, changes: Partial<Pick<RouteSender, 'status' | 'position'>>) =>
    run(
      () => fetch(`/api/admin/route-senders/${s.id}`, json('PUT', { status: s.status, position: s.position, ...changes })),
      `${s.sender_id} updated`
    )

  const deleteSender = (s: RouteSender) => {
    if (!confirm(`Remove ${s.sender_id} from this connection's pool?`)) return
    run(() => fetch(`/api/admin/route-senders/${s.id}`, { method: 'DELETE' }), `${s.sender_id} removed`)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-zinc-500" />
      </div>
    )
  }

  const activeCount = pool?.senders.filter((s) => s.status === 'active').length ?? 0
  const selectedStrategy = SENDER_POOL_STRATEGIES.find((s) => s.value === strategy)

  return (
    <div className="bg-zinc-900 rounded-lg border border-white/10 p-6">
      <h2 className="text-lg font-semibold text-white mb-1 flex items-center gap-2">
        <Users className="h-5 w-5" />
        Route Sender Pool
      </h2>
      <p className="text-sm text-zinc-400 mb-6">
        Sender IDs this connection accepts. For service providers with the sender pool turned on, messages sent through
        this connection go out with one of these senders instead of the provider&apos;s own.
      </p>

      <div className="flex flex-wrap items-end gap-3 mb-6">
        <div>
          <label htmlFor={`${ids}-strategy`} className="block text-sm font-medium text-zinc-300 mb-1">
            Rotation
          </label>
          <select
            id={`${ids}-strategy`}
            value={strategy}
            onChange={(e) => setStrategy(e.target.value)}
            className="px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
          >
            {SENDER_POOL_STRATEGIES.map((s) => (
              <option key={s.value} value={s.value}>{s.label}</option>
            ))}
          </select>
        </div>
        <Button
          type="button"
          onClick={saveStrategy}
          disabled={busy || strategy === pool?.strategy}
          className="bg-white text-black hover:bg-zinc-200"
        >
          Save rotation
        </Button>
      </div>
      {selectedStrategy && <p className="text-xs text-zinc-500 -mt-4 mb-6">{selectedStrategy.description}</p>}

      {activeCount === 0 && (
        <p className="text-sm text-amber-400 mb-4">
          No active sender. Messages from sender pool providers fail on this connection and move to the next route.
        </p>
      )}

      {pool && pool.senders.length > 0 && (
        <div className="overflow-x-auto mb-6">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-zinc-400 border-b border-white/10">
                <th className="py-2 pr-4 font-medium">Sender ID</th>
                <th className="py-2 pr-4 font-medium">Position</th>
                <th className="py-2 pr-4 font-medium">Status</th>
                <th className="py-2 pr-4 font-medium">Used</th>
                <th className="py-2 pr-4 font-medium">Last used</th>
                <th className="py-2 font-medium sr-only">Actions</th>
              </tr>
            </thead>
            <tbody>
              {pool.senders.map((s) => (
                <tr key={s.id} className="border-b border-white/5">
                  <td className="py-2 pr-4 font-mono text-white">{s.sender_id}</td>
                  <td className="py-2 pr-4">
                    <input
                      type="number"
                      aria-label={`Position of ${s.sender_id}`}
                      defaultValue={s.position}
                      disabled={busy}
                      onBlur={(e) => {
                        const position = Number(e.target.value)
                        if (Number.isInteger(position) && position !== s.position) updateSender(s, { position })
                      }}
                      className="w-20 px-2 py-1 border border-white/10 rounded bg-zinc-800 text-white"
                    />
                  </td>
                  <td className="py-2 pr-4">
                    <button
                      type="button"
                      disabled={busy}
                      onClick={() => updateSender(s, { status: s.status === 'active' ? 'inactive' : 'active' })}
                      className={`px-2 py-0.5 rounded-full text-xs ${
                        s.status === 'active' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-zinc-500/20 text-zinc-400'
                      }`}
                      aria-label={`${s.status === 'active' ? 'Turn off' : 'Turn on'} ${s.sender_id}`}
                    >
                      {s.status}
                    </button>
                  </td>
                  <td className="py-2 pr-4 text-zinc-300">{s.usage_count.toLocaleString()}</td>
                  <td className="py-2 pr-4 text-zinc-400">{s.last_used_at ? new Date(s.last_used_at).toLocaleString() : 'Never'}</td>
                  <td className="py-2 text-right">
                    <button
                      type="button"
                      disabled={busy}
                      onClick={() => deleteSender(s)}
                      className="text-zinc-500 hover:text-red-400"
                      aria-label={`Remove ${s.sender_id}`}
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <form onSubmit={addSender} className="flex flex-wrap items-end gap-3">
        <div>
          <label htmlFor={`${ids}-sender`} className="block text-sm font-medium text-zinc-300 mb-1">
            Sender ID
          </label>
          <input
            id={`${ids}-sender`}
            type="text"
            maxLength={16}
            value={newSender}
            onChange={(e) => setNewSender(e.target.value)}
            placeholder="e.g. SONDER"
            className="px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white placeholder:text-zinc-500"
          />
        </div>
        <div>
          <label htmlFor={`${ids}-position`} className="block text-sm font-medium text-zinc-300 mb-1">
            Position
          </label>
          <input
            id={`${ids}-position`}
            type="number"
            value={newPosition}
            onChange={(e) => setNewPosition(e.target.value)}
            className="w-24 px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
          />
        </div>
        <Button type="submit" disabled={busy || !newSender.trim()} className="bg-white text-black hover:bg-zinc-200">
          {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
          Add sender
        </Button>
      </form>

      {errorMessage && <p className="text-sm text-red-400 mt-4">{errorMessage}</p>}
      {savedMessage && <p className="text-sm text-emerald-400 mt-4">{savedMessage}</p>}
    </div>
  )
}
