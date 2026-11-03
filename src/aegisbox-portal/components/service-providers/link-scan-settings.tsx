'use client'

import { useEffect, useState } from 'react'
import { Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface Config {
  mode: 'off' | 'monitor' | 'hold' | 'block'
  hold_seconds: number
  fail_open: boolean
  block_suspicious: boolean
  charge_blocked: boolean
  updated_by: string | null
  updated_at: string | null
}

const MODES: { value: Config['mode']; label: string; help: string }[] = [
  { value: 'off', label: 'Off', help: 'Real messages are not scanned.' },
  { value: 'monitor', label: 'Monitor', help: 'Links are scanned after the message is sent. Nothing is delayed or blocked; results appear on Link Scans.' },
  { value: 'hold', label: 'Hold for review', help: 'Messages with links wait for the verdict. A bad verdict holds the message on Link Scans until an admin sends or rejects it (rejected after 24 hours).' },
  { value: 'block', label: 'Block', help: 'Messages with links wait for the verdict. A bad verdict rejects the message: it is not sent and the customer gets a REJECTD DLR.' },
]

export function LinkScanSettings({ providerId }: Readonly<{ providerId: number }>) {
  const [cfg, setCfg] = useState<Config | null>(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [saved, setSaved] = useState<string | null>(null)

  useEffect(() => {
    fetch(`/api/admin/service-providers/${providerId}/link-scan-settings`)
      .then(async (r) => {
        const d = await r.json().catch(() => ({}))
        if (r.ok) setCfg(d)
        else setError(d.error || 'Failed to load')
      })
      .catch(() => setError('Failed to load'))
  }, [providerId])

  if (!cfg) {
    return error ? <p className="text-sm text-red-400">{error}</p> : (
      <div className="flex items-center justify-center py-12"><Loader2 className="h-6 w-6 animate-spin text-zinc-500" /></div>
    )
  }

  async function save(e: React.FormEvent) {
    e.preventDefault()
    if (!cfg) return
    if ((cfg.mode === 'hold' || cfg.mode === 'block') &&
      !window.confirm('This delays this customer\'s real messages that contain links, and can stop some of them. Have the contract and NCC checks been done?')) {
      return
    }
    setSaving(true)
    setError(null)
    setSaved(null)
    try {
      const res = await fetch(`/api/admin/service-providers/${providerId}/link-scan-settings`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(cfg),
      })
      const d = await res.json().catch(() => ({}))
      if (!res.ok) {
        setError(d.error || 'Failed to save')
        return
      }
      setCfg(d)
      setSaved('Saved. The gateway picks it up within 30 seconds.')
    } finally {
      setSaving(false)
    }
  }

  const waits = cfg.mode === 'hold' || cfg.mode === 'block'
  const input = 'rounded-lg bg-zinc-900 border border-white/10 px-3 py-2 text-sm text-white'
  const check = (key: 'fail_open' | 'block_suspicious' | 'charge_blocked', label: string, help: string) => (
    <label className="flex items-start gap-3 rounded-lg border border-white/10 px-4 py-3">
      <input type="checkbox" className="mt-1" checked={cfg[key]} onChange={(e) => setCfg({ ...cfg, [key]: e.target.checked })} />
      <span>
        <span className="block font-medium text-white text-sm">{label}</span>
        <span className="block text-sm text-zinc-500">{help}</span>
      </span>
    </label>
  )

  return (
    <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
      <h3 className="text-lg font-semibold text-white">Link check of real messages</h3>
      <p className="text-sm text-zinc-400 mt-0.5 mb-6">
        Opens the links in this customer&apos;s messages in the link scanner. Needs link scanning turned on for the gateway.
        Messages that look like OTPs and messages without web links are never delayed.
      </p>
      <form onSubmit={save} className="space-y-5">
        <div className="grid gap-2 md:grid-cols-2">
          {MODES.map((m) => (
            <label key={m.value} className={`rounded-lg border px-4 py-3 cursor-pointer ${cfg.mode === m.value ? 'border-white/40 bg-white/5' : 'border-white/10'}`}>
              <span className="flex items-center gap-2">
                <input type="radio" name="mode" checked={cfg.mode === m.value} onChange={() => setCfg({ ...cfg, mode: m.value })} />
                <span className="font-medium text-white text-sm">{m.label}</span>
              </span>
              <span className="block text-sm text-zinc-500 mt-1">{m.help}</span>
            </label>
          ))}
        </div>
        {waits && (
          <div className="space-y-3">
            <label className="flex flex-wrap items-center gap-3 text-sm text-zinc-300">
              Wait at most
              <input type="number" min={5} max={600} className={`${input} w-24`} value={cfg.hold_seconds}
                onChange={(e) => setCfg({ ...cfg, hold_seconds: Number(e.target.value) })} />
              seconds for verdicts
            </label>
            {check('fail_open', 'Send when there is no verdict in time', 'If the scanner is down or slow, send the message anyway. Off: treat it as bad (review or reject).')}
            {check('block_suspicious', 'Also stop suspicious messages', 'Off: only malicious verdicts stop a message.')}
            {check('charge_blocked', 'Charge for rejected messages', 'Off: the charge is refunded when a message is rejected.')}
          </div>
        )}
        <div className="flex flex-wrap items-center gap-3">
          <Button type="submit" disabled={saving} className="bg-white text-black hover:bg-zinc-200">{saving ? 'Saving…' : 'Save'}</Button>
          {saved && <span className="text-sm text-emerald-400">{saved}</span>}
          {error && <span className="text-sm text-red-400">{error}</span>}
          {cfg.updated_by && cfg.updated_at && <span className="text-xs text-zinc-500">Last changed by {cfg.updated_by} on {new Date(cfg.updated_at).toLocaleString()}</span>}
        </div>
      </form>
    </div>
  )
}
