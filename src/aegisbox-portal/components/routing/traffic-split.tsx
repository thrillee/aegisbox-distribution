'use client'

import { useState, useEffect, useCallback, useId } from 'react'
import { Loader2, Save, Shuffle, AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface RouteRow {
  assignment_id: number
  network: string
  mno_id: number
  mno_name: string
  weight: number
  assignment_active: boolean
  mno_active: boolean
  has_connection: boolean
  manually_unavailable: boolean
  health_excluded: boolean
  consecutive_failures: number
}

// Why a route gets no traffic right now, mirroring the gateway's route lookup.
function unusableReason(r: RouteRow): string | null {
  if (!r.assignment_active) return 'route turned off'
  if (!r.mno_active) return 'MNO inactive'
  if (r.manually_unavailable) return 'marked unavailable on the route health page'
  if (r.health_excluded) return `taken out after ${r.consecutive_failures} failed sends in a row; retried every minute`
  if (!r.has_connection) return 'no active connection'
  return null
}

// Share of a network's traffic each route gets: weight over the total of usable routes. All
// usable routes at weight 0 are split evenly, as the router does.
function networkShares(routes: RouteRow[], weightOf: (r: RouteRow) => number) {
  const usable = routes.filter((r) => unusableReason(r) === null)
  const total = usable.reduce((sum, r) => sum + weightOf(r), 0)
  return new Map(
    routes.map((r) => {
      if (unusableReason(r) !== null) return [r.assignment_id, 0]
      return [r.assignment_id, total > 0 ? weightOf(r) / total : 1 / usable.length]
    })
  )
}

const pct = (share: number) => `${(share * 100).toFixed(share > 0 && share < 0.1 ? 1 : 0)}%`

export function TrafficSplit({ groupId }: Readonly<{ groupId: number }>) {
  const ids = useId()
  const [loading, setLoading] = useState(true)
  const [routes, setRoutes] = useState<RouteRow[]>([])
  const [draft, setDraft] = useState<Record<number, string>>({})
  const [saving, setSaving] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [savedMessage, setSavedMessage] = useState<string | null>(null)

  const apply = (rows: RouteRow[]) => {
    setRoutes(rows)
    const next: Record<number, string> = {}
    for (const r of rows) {
      // One weight per MNO; if networks differ, start from the first and flag it.
      if (next[r.mno_id] === undefined) next[r.mno_id] = String(r.weight)
    }
    setDraft(next)
  }

  const load = useCallback(async () => {
    try {
      const res = await fetch(`/api/admin/routing-groups/${groupId}/split`)
      const data = await res.json().catch(() => ({}))
      if (res.ok) apply(data.routes)
      else setErrorMessage(data.error || 'Failed to load the traffic split')
    } catch {
      setErrorMessage('Failed to load the traffic split')
    } finally {
      setLoading(false)
    }
  }, [groupId])

  useEffect(() => {
    load()
  }, [load])

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-zinc-500" />
      </div>
    )
  }
  if (routes.length === 0) return null

  const mnos = [...new Map(routes.map((r) => [r.mno_id, r.mno_name])).entries()]
  const networks = [...new Set(routes.map((r) => r.network))]
  const mixed = mnos.filter(([id]) => new Set(routes.filter((r) => r.mno_id === id).map((r) => r.weight)).size > 1)

  const parsed = (mnoId: number) => {
    const n = Number(draft[mnoId])
    return Number.isInteger(n) ? n : NaN
  }
  const invalid = mnos.some(([id]) => {
    const n = parsed(id)
    return Number.isNaN(n) || n < 0 || n > 100
  })
  const changed = mnos.some(([id]) => routes.some((r) => r.mno_id === id && r.weight !== parsed(id)))
  const draftTotal = mnos.reduce((sum, [id]) => sum + (Number.isNaN(parsed(id)) ? 0 : parsed(id)), 0)

  const save = async () => {
    setErrorMessage(null)
    setSavedMessage(null)
    setSaving(true)
    try {
      const weights = Object.fromEntries(mnos.map(([id]) => [id, parsed(id)]))
      const res = await fetch(`/api/admin/routing-groups/${groupId}/split`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ weights }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setErrorMessage(data.error || 'Failed to save the traffic split')
        return
      }
      apply(data.routes)
      setSavedMessage('Traffic split saved. New messages use it immediately.')
    } catch {
      setErrorMessage('Failed to save the traffic split')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
      <h2 className="font-semibold text-white flex items-center gap-2">
        <Shuffle className="h-4 w-4" />
        Traffic Split
      </h2>
      <p className="text-sm text-zinc-400 mt-1 mb-5">
        Each message goes to one route at random in proportion to its weight (0 to 100). A route at 0 only takes
        traffic when the others are unavailable. The same weights apply to every network in this group.
      </p>

      {mixed.length > 0 && (
        <p className="text-sm text-amber-400 mb-4 flex items-start gap-2">
          <AlertTriangle className="h-4 w-4 mt-0.5 shrink-0" />
          {mixed.map(([, name]) => name).join(', ')} currently {mixed.length === 1 ? 'has' : 'have'} different weights on
          different networks. Saving sets the same weight on all of them.
        </p>
      )}

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 mb-5">
        {mnos.map(([id, name]) => {
          const n = parsed(id)
          const share = !Number.isNaN(n) && draftTotal > 0 ? n / draftTotal : 0
          return (
            <div key={id} className="rounded-lg border border-white/10 px-4 py-3">
              <label htmlFor={`${ids}-${id}`} className="block text-sm font-medium text-white mb-2">{name}</label>
              <div className="flex items-center gap-3">
                <input
                  id={`${ids}-${id}`}
                  type="number"
                  min={0}
                  max={100}
                  step={1}
                  value={draft[id] ?? ''}
                  onChange={(e) => setDraft((d) => ({ ...d, [id]: e.target.value }))}
                  className="w-24 px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
                />
                <span className="text-sm text-zinc-400">
                  {n === 0 ? 'backup only' : `${pct(share)} when all routes are up`}
                </span>
              </div>
            </div>
          )
        })}
      </div>

      <h3 className="text-sm font-medium text-zinc-300 mb-2">Share each route gets right now</h3>
      <div className="overflow-x-auto mb-5">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-zinc-400 border-b border-white/10">
              <th className="py-2 pr-4 font-medium">Network</th>
              {mnos.map(([id, name]) => (
                <th key={id} className="py-2 pr-4 font-medium">{name}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {networks.map((network) => {
              const rows = routes.filter((r) => r.network === network)
              const shares = networkShares(rows, (r) => r.weight)
              const noRoute = rows.every((r) => unusableReason(r) !== null)
              return (
                <tr key={network} className="border-b border-white/5 align-top">
                  <td className="py-2 pr-4 text-white">
                    {network}
                    {noRoute && <span className="block text-xs text-red-400">No usable route: messages fail</span>}
                  </td>
                  {mnos.map(([id]) => {
                    const r = rows.find((row) => row.mno_id === id)
                    if (!r) return <td key={id} className="py-2 pr-4 text-zinc-600">not routed</td>
                    const reason = unusableReason(r)
                    return (
                      <td key={id} className="py-2 pr-4">
                        <span className={reason ? 'text-zinc-500' : 'text-white'}>{pct(shares.get(r.assignment_id) ?? 0)}</span>
                        <span className="text-zinc-500"> (weight {r.weight})</span>
                        {reason && <span className="block text-xs text-amber-400">{reason}</span>}
                      </td>
                    )
                  })}
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {invalid && <p className="text-sm text-red-400 mb-3">Each weight must be a whole number from 0 to 100.</p>}
      {errorMessage && <p className="text-sm text-red-400 mb-3">{errorMessage}</p>}
      {savedMessage && <p className="text-sm text-emerald-400 mb-3">{savedMessage}</p>}

      <Button
        type="button"
        onClick={save}
        disabled={saving || invalid || !changed}
        className="bg-white text-black hover:bg-zinc-200"
      >
        {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
        {saving ? 'Saving...' : 'Save split'}
      </Button>
    </div>
  )
}
