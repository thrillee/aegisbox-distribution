'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'

type Customer = { id: number; name: string }

export function WhitelistForm({ testCustomers }: { testCustomers: Customer[] }) {
  const router = useRouter()
  const [msisdn, setMsisdn] = useState('')
  const [spId, setSpId] = useState('')
  const [note, setNote] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setSaving(true)
    try {
      const res = await fetch('/api/admin/simulator-whitelist', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ msisdn, service_provider_id: spId === '' ? null : Number(spId), note }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setError(data.error || 'Failed to add the number')
        return
      }
      setMsisdn('')
      setNote('')
      router.refresh()
    } finally {
      setSaving(false)
    }
  }

  const input = 'w-full rounded-lg bg-zinc-900 border border-white/10 px-3 py-2 text-sm text-white placeholder:text-zinc-500 focus:outline-none focus:border-white/30'
  return (
    <form onSubmit={submit} className="grid gap-3 md:grid-cols-[1fr_1fr_1fr_auto] items-end">
      <label className="space-y-1 text-sm text-zinc-400">
        <span>Phone number</span>
        <input className={input} value={msisdn} onChange={(e) => setMsisdn(e.target.value)} placeholder="2348031234567 or 08031234567" required inputMode="tel" />
      </label>
      <label className="space-y-1 text-sm text-zinc-400">
        <span>Applies to</span>
        <select className={input} value={spId} onChange={(e) => setSpId(e.target.value)}>
          <option value="">All test customers</option>
          {testCustomers.map((c) => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </select>
      </label>
      <label className="space-y-1 text-sm text-zinc-400">
        <span>Note (optional)</span>
        <input className={input} value={note} onChange={(e) => setNote(e.target.value)} placeholder="Whose phone" maxLength={200} />
      </label>
      <Button type="submit" disabled={saving} className="bg-white text-black hover:bg-zinc-200">
        {saving ? 'Adding…' : 'Add number'}
      </Button>
      {error && <p className="md:col-span-4 text-sm text-red-400">{error}</p>}
    </form>
  )
}

export function RemoveButton({ id, msisdn }: { id: number; msisdn: string }) {
  const router = useRouter()
  const [busy, setBusy] = useState(false)
  async function remove() {
    if (!window.confirm(`Remove ${msisdn}? Test messages to it will go to the simulator again.`)) return
    setBusy(true)
    try {
      const res = await fetch(`/api/admin/simulator-whitelist/${id}`, { method: 'DELETE' })
      if (res.ok) router.refresh()
    } finally {
      setBusy(false)
    }
  }
  return (
    <button onClick={remove} disabled={busy} className="text-sm text-red-400 hover:text-red-300 disabled:opacity-50">
      {busy ? 'Removing…' : 'Remove'}
    </button>
  )
}
