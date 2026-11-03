'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'

type Customer = { id: number; name: string }

async function post(url: string, body: unknown, method = 'POST') {
  const res = await fetch(url, { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
  const data = await res.json().catch(() => ({}))
  return { ok: res.ok, error: (data as { error?: string }).error }
}

const small = 'h-8 px-3 text-xs'

// Release (send) or reject a message held by the link check.
export function HoldButtons({ messageId }: { messageId: number }) {
  const router = useRouter()
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  async function act(action: 'release' | 'reject') {
    const question = action === 'release'
      ? `Send message ${messageId} now, despite the link check?`
      : `Reject message ${messageId}? It will not be sent and the customer gets a REJECTD DLR.`
    if (!window.confirm(question)) return
    setBusy(true)
    setError(null)
    try {
      const r = await post(`/api/admin/link-holds/${messageId}/${action}`, {})
      if (!r.ok) setError(r.error || 'Failed')
      else router.refresh()
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="flex flex-wrap items-center gap-2">
      <Button disabled={busy} onClick={() => act('release')} className={`${small} bg-emerald-600 hover:bg-emerald-500 text-white`}>Send</Button>
      <Button disabled={busy} onClick={() => act('reject')} className={`${small} bg-red-600 hover:bg-red-500 text-white`}>Reject</Button>
      {error && <span className="text-xs text-red-400">{error}</span>}
    </div>
  )
}

// An admin's correction of one scan result.
export function OverrideButtons({ scanId, current }: { scanId: number; current: string | null }) {
  const router = useRouter()
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  async function set(verdict: string | null) {
    let note: string | null = null
    if (verdict !== null) {
      note = window.prompt(verdict === 'safe' ? 'Why is this safe? (optional)' : 'Why is this malicious? (optional)', '')
      if (note === null) return
    }
    setBusy(true)
    setError(null)
    try {
      const r = await post(`/api/admin/link-scans/${scanId}/override`, { verdict, note })
      if (!r.ok) setError(r.error || 'Failed')
      else router.refresh()
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="flex flex-wrap items-center gap-2">
      {current !== 'safe' && (
        <Button disabled={busy} variant="outline" onClick={() => set('safe')} className={small}>False positive: mark safe</Button>
      )}
      {current !== 'malicious' && (
        <Button disabled={busy} variant="outline" onClick={() => set('malicious')} className={small}>Mark malicious</Button>
      )}
      {current !== null && (
        <Button disabled={busy} variant="outline" onClick={() => set(null)} className={small}>Clear correction</Button>
      )}
      {error && <span className="text-xs text-red-400">{error}</span>}
    </div>
  )
}

// Allow or block a domain from a scan result, for every customer or the message's customer.
export function QuickListButtons({ domain, customer }: { domain: string; customer: Customer }) {
  const router = useRouter()
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<string | null>(null)
  async function add(listType: 'allow' | 'block', spId: number | null) {
    const scope = spId === null ? 'every customer' : customer.name
    if (!window.confirm(`${listType === 'allow' ? 'Allow' : 'Block'} ${domain} and its subdomains for ${scope}?`)) return
    setBusy(true)
    setMsg(null)
    try {
      const r = await post('/api/admin/link-lists', { domain, list_type: listType, service_provider_id: spId })
      setMsg(r.ok ? `Saved: applies to new scans within 30 seconds` : r.error || 'Failed')
      if (r.ok) router.refresh()
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-xs text-zinc-500">{domain}:</span>
      <Button disabled={busy} variant="outline" onClick={() => add('allow', customer.id)} className={small}>Allow for {customer.name}</Button>
      <Button disabled={busy} variant="outline" onClick={() => add('allow', null)} className={small}>Allow for all</Button>
      <Button disabled={busy} variant="outline" onClick={() => add('block', null)} className={small}>Block for all</Button>
      {msg && <span className="text-xs text-zinc-400">{msg}</span>}
    </div>
  )
}

export function ListEntryForm({ customers }: { customers: Customer[] }) {
  const router = useRouter()
  const [domain, setDomain] = useState('')
  const [listType, setListType] = useState<'allow' | 'block'>('block')
  const [spId, setSpId] = useState('')
  const [note, setNote] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const r = await post('/api/admin/link-lists', { domain, list_type: listType, service_provider_id: spId === '' ? null : Number(spId), note })
      if (!r.ok) {
        setError(r.error || 'Failed to save')
        return
      }
      setDomain('')
      setNote('')
      router.refresh()
    } finally {
      setSaving(false)
    }
  }
  const input = 'w-full rounded-lg bg-zinc-900 border border-white/10 px-3 py-2 text-sm text-white placeholder:text-zinc-500 focus:outline-none focus:border-white/30'
  return (
    <form onSubmit={submit} className="grid gap-3 md:grid-cols-[2fr_1fr_1.5fr_1.5fr_auto] items-end">
      <label className="space-y-1 text-sm text-zinc-400">
        <span>Domain</span>
        <input className={input} value={domain} onChange={(e) => setDomain(e.target.value)} placeholder="example.com" required />
      </label>
      <label className="space-y-1 text-sm text-zinc-400">
        <span>List</span>
        <select className={input} value={listType} onChange={(e) => setListType(e.target.value as 'allow' | 'block')}>
          <option value="block">Block</option>
          <option value="allow">Allow</option>
        </select>
      </label>
      <label className="space-y-1 text-sm text-zinc-400">
        <span>Applies to</span>
        <select className={input} value={spId} onChange={(e) => setSpId(e.target.value)}>
          <option value="">Every customer</option>
          {customers.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
        </select>
      </label>
      <label className="space-y-1 text-sm text-zinc-400">
        <span>Note (optional)</span>
        <input className={input} value={note} onChange={(e) => setNote(e.target.value)} maxLength={300} />
      </label>
      <Button type="submit" disabled={saving} className="bg-white text-black hover:bg-zinc-200">{saving ? 'Saving…' : 'Save'}</Button>
      {error && <p className="md:col-span-5 text-sm text-red-400">{error}</p>}
    </form>
  )
}

export function RemoveListEntry({ id, domain }: { id: number; domain: string }) {
  const router = useRouter()
  const [busy, setBusy] = useState(false)
  async function remove() {
    if (!window.confirm(`Remove ${domain} from the list?`)) return
    setBusy(true)
    try {
      const res = await fetch(`/api/admin/link-lists/${id}`, { method: 'DELETE' })
      if (res.ok) router.refresh()
    } finally {
      setBusy(false)
    }
  }
  return (
    <Button disabled={busy} variant="outline" onClick={remove} className={small}>Remove</Button>
  )
}
