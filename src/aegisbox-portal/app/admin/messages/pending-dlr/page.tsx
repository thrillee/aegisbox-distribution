'use client'

import { useCallback, useEffect, useState } from 'react'
import Link from 'next/link'
import { AlertTriangle, Loader2, Send } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'
import { DLR_STATUSES } from '@/components/messages/manual-dlr-form'
import { reportDate } from '@/lib/time'

interface Summary {
  accepted: number
  noAnswer: number
  failed: number
  tooRecent: number
}

interface SendResult {
  matched: number
  forwarded: number
  skipped: number
  failed: number
  warnings: number
  error?: string
}

const input = 'w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white'

export default function PendingDlrPage() {
  const [customers, setCustomers] = useState<{ id: number; name: string }[]>([])
  const [filter, setFilter] = useState(() => ({ sp_id: '', from: reportDate(), to: reportDate(), older_than_minutes: '60' }))
  const [summary, setSummary] = useState<Summary | null>(null)
  const [loading, setLoading] = useState(false)
  const [status, setStatus] = useState('DELIVRD')
  const [includeNoAnswer, setIncludeNoAnswer] = useState(false)
  const [confirming, setConfirming] = useState(false)
  const [sending, setSending] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [result, setResult] = useState<SendResult | null>(null)

  useEffect(() => {
    fetch('/api/admin/service-providers')
      .then((r) => r.json())
      .then((d) => setCustomers((d.providers ?? []).map((p: { id: number; name: string }) => ({ id: p.id, name: p.name }))))
      .catch(() => setErrorMessage('Failed to load customers'))
  }, [])

  const load = useCallback(async () => {
    setSummary(null)
    setConfirming(false)
    setErrorMessage(null)
    if (!filter.sp_id) return
    setLoading(true)
    try {
      const res = await fetch(`/api/admin/messages/pending-dlr?${new URLSearchParams(filter)}`)
      const data = await res.json().catch(() => ({}))
      if (res.ok) setSummary(data)
      else setErrorMessage(data.error || 'Failed to count messages')
    } catch {
      setErrorMessage('Failed to count messages')
    } finally {
      setLoading(false)
    }
  }, [filter])

  useEffect(() => {
    load()
  }, [load])

  const toSend = summary ? summary.accepted + (includeNoAnswer ? summary.noAnswer : 0) : 0
  const customerName = customers.find((c) => String(c.id) === filter.sp_id)?.name ?? ''

  const send = async () => {
    setSending(true)
    setErrorMessage(null)
    setResult(null)
    try {
      const res = await fetch('/api/admin/messages/pending-dlr', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...filter, status, include_no_answer: includeNoAnswer }),
      })
      const data = await res.json().catch(() => ({}))
      if (data.forwarded !== undefined) setResult(data)
      if (!res.ok) setErrorMessage(data.error || 'Failed to send DLRs')
    } catch {
      setErrorMessage('Failed to send DLRs; check the DLR Queue for what was queued')
    } finally {
      setSending(false)
      setConfirming(false)
      load()
    }
  }

  return (
    <div className="p-8 space-y-6 max-w-4xl">
      <PageHeader
        title="Pending DLRs"
        description="Send a DLR to a customer for messages that never got one from the upstream"
      />

      <DataCard title="Messages">
        <div className="grid gap-4 md:grid-cols-4">
          <label className="text-sm text-zinc-400 md:col-span-2">
            <span className="block mb-1">Customer</span>
            <select value={filter.sp_id} onChange={(e) => setFilter((f) => ({ ...f, sp_id: e.target.value }))} className={input}>
              <option value="">Choose a customer</option>
              {customers.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </label>
          <label className="text-sm text-zinc-400">
            <span className="block mb-1">From (Lagos)</span>
            <input type="date" value={filter.from} onChange={(e) => setFilter((f) => ({ ...f, from: e.target.value }))} className={input} />
          </label>
          <label className="text-sm text-zinc-400">
            <span className="block mb-1">To (Lagos)</span>
            <input type="date" value={filter.to} onChange={(e) => setFilter((f) => ({ ...f, to: e.target.value }))} className={input} />
          </label>
          <label className="text-sm text-zinc-400 md:col-span-2">
            <span className="block mb-1">Only messages sent at least this many minutes ago</span>
            <input
              type="number"
              min={0}
              value={filter.older_than_minutes}
              onChange={(e) => setFilter((f) => ({ ...f, older_than_minutes: e.target.value }))}
              className={input}
            />
          </label>
        </div>

        {loading && <Loader2 className="h-5 w-5 animate-spin text-zinc-500 mt-6" />}
        {summary && (
          <div className="mt-6 overflow-x-auto">
            <table className="w-full text-sm">
              <tbody>
                <tr className="border-b border-white/5">
                  <td className="py-2 pr-4 text-white">Accepted by the upstream, no DLR yet</td>
                  <td className="py-2 text-right text-white font-medium">{summary.accepted.toLocaleString()}</td>
                </tr>
                <tr className="border-b border-white/5">
                  <td className="py-2 pr-4 text-white">
                    No answer from the upstream
                    <span className="block text-xs text-zinc-500">Sent, but the upstream never confirmed or rejected it, so it may not have received it</span>
                  </td>
                  <td className="py-2 text-right text-white font-medium">{summary.noAnswer.toLocaleString()}</td>
                </tr>
                <tr className="border-b border-white/5">
                  <td className="py-2 pr-4 text-zinc-400">
                    Failed or not sent: never changed here
                    <span className="block text-xs text-zinc-500">Rejected or failed by the upstream, or not sent yet</span>
                  </td>
                  <td className="py-2 text-right text-zinc-400">{summary.failed.toLocaleString()}</td>
                </tr>
                <tr>
                  <td className="py-2 pr-4 text-zinc-400">Sent too recently (still waiting for a real DLR)</td>
                  <td className="py-2 text-right text-zinc-400">{summary.tooRecent.toLocaleString()}</td>
                </tr>
              </tbody>
            </table>
          </div>
        )}
      </DataCard>

      {summary && (
        <DataCard title="Send DLR">
          <div className="space-y-4">
            <label className="text-sm text-zinc-400 block max-w-xs">
              <span className="block mb-1">Status to send</span>
              <select value={status} onChange={(e) => { setStatus(e.target.value); setConfirming(false) }} className={input}>
                {DLR_STATUSES.map((s) => (
                  <option key={s.value} value={s.value}>{s.label}</option>
                ))}
              </select>
            </label>
            <label className="flex items-start gap-2 text-sm text-zinc-300">
              <input
                type="checkbox"
                checked={includeNoAnswer}
                onChange={(e) => { setIncludeNoAnswer(e.target.checked); setConfirming(false) }}
                className="mt-1"
              />
              <span>
                Also send it for the {summary.noAnswer.toLocaleString()} message(s) with no answer from the upstream
                <span className="block text-xs text-zinc-500">The upstream may never have received these.</span>
              </span>
            </label>

            <p className="text-sm text-zinc-400">
              Each message gets {status} and the DLR is sent to the customer straight away. A real DLR arriving later for
              these messages is ignored. Wallets are not changed.
            </p>

            {!confirming ? (
              <Button
                type="button"
                onClick={() => setConfirming(true)}
                disabled={toSend === 0 || sending}
                className="bg-white text-black hover:bg-zinc-200"
              >
                <Send className="h-4 w-4" />
                Send {status} to {toSend.toLocaleString()} message(s)
              </Button>
            ) : (
              <div className="p-4 rounded-xl border border-amber-500/30 bg-amber-500/10 space-y-3">
                <p className="text-sm text-amber-300 flex items-start gap-2">
                  <AlertTriangle className="h-4 w-4 mt-0.5 shrink-0" />
                  Send {status} to {customerName} for {toSend.toLocaleString()} message(s) from {filter.from} to {filter.to}? This cannot be undone.
                </p>
                <div className="flex gap-3">
                  <Button type="button" onClick={send} disabled={sending} className="bg-white text-black hover:bg-zinc-200">
                    {sending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
                    {sending ? 'Sending...' : 'Yes, send'}
                  </Button>
                  <Button type="button" variant="ghost" onClick={() => setConfirming(false)} disabled={sending} className="text-zinc-300">
                    Cancel
                  </Button>
                </div>
              </div>
            )}
          </div>
        </DataCard>
      )}

      {errorMessage && <p className="text-sm text-red-400">{errorMessage}</p>}
      {result && (
        <DataCard title="Result">
          <p className="text-sm text-zinc-300">
            {result.forwarded.toLocaleString()} DLR(s) queued to the customer, {result.skipped.toLocaleString()} skipped
            (already completed in the meantime), {result.failed.toLocaleString()} failed.
            {result.warnings > 0 && ` ${result.warnings} cannot reach the customer until their connection or callback is set up.`}
          </p>
          <p className="text-sm text-zinc-400 mt-2">
            Delivery to the customer is shown on the <Link href="/admin/messages/queues/dlr" className="text-white hover:underline">DLR Queue</Link> page.
          </p>
        </DataCard>
      )}
    </div>
  )
}
