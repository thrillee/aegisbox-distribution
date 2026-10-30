'use client'

import { useState } from 'react'
import { Send, X } from 'lucide-react'
import { Button } from '@/components/ui/button'

export const DLR_STATUSES = [
  { value: 'DELIVRD', label: 'DELIVRD (Delivered)' },
  { value: 'UNDELIV', label: 'UNDELIV (Undelivered)' },
  { value: 'EXPIRED', label: 'EXPIRED (Expired)' },
  { value: 'REJECTD', label: 'REJECTD (Rejected)' },
  { value: 'DELETED', label: 'DELETED (Deleted)' },
] as const

export const FAILURE_DLR_STATUSES = ['UNDELIV', 'EXPIRED', 'REJECTD', 'DELETED']

const FINAL_DLR_STATUSES = DLR_STATUSES.map((s) => s.value as string)

interface SegmentDlrState {
  dlrStatus: string | null
  mnoMessageId: string | null
  sentToMnoAt?: string | null
  errorCode?: string | null
  errorDescription?: string | null
}

const hasFinalDlr = (s: SegmentDlrState) => !!s.dlrStatus && FINAL_DLR_STATUSES.includes(s.dlrStatus.toUpperCase())

// Sent, but the MNO never confirmed or rejected it: no MNO message ID and either no error recorded
// (TIMEOUT only marks "no DLR in time") or the submit timed out (MNO_TIMEOUT). Same rule as the
// gateway's include_unconfirmed; rejections are never included.
export function isUnconfirmedSegment(s: SegmentDlrState): boolean {
  if (s.mnoMessageId || !s.sentToMnoAt) return false
  return ((!s.errorCode || s.errorCode === 'TIMEOUT') && !s.errorDescription) || s.errorCode === 'MNO_TIMEOUT'
}

// The MNO accepted the segment but no final DLR came.
export function isAwaitingFinalDlr(segment: SegmentDlrState): boolean {
  return !!segment.mnoMessageId && !hasFinalDlr(segment)
}

// Waiting for a DLR with no answer from the MNO.
export function isAwaitingUnconfirmed(segment: SegmentDlrState): boolean {
  return isUnconfirmedSegment(segment) && !hasFinalDlr(segment)
}

interface BulkResult {
  forwarded: number
  skipped: number
  failed: number
  warnings: number
}

interface ManualDlrFormProps {
  messageIds: number[]
  onDone: (summary: string) => void
  onCancel: () => void
  // Tick "no answer from the MNO" by default (e.g. a message whose only waiting parts are unconfirmed).
  defaultIncludeUnconfirmed?: boolean
}

export function ManualDlrForm({ messageIds, onDone, onCancel, defaultIncludeUnconfirmed = false }: Readonly<ManualDlrFormProps>) {
  const [status, setStatus] = useState<string>('DELIVRD')
  const [errorCode, setErrorCode] = useState('')
  const [includeUnconfirmed, setIncludeUnconfirmed] = useState(defaultIncludeUnconfirmed)
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const isBulk = messageIds.length > 1
  const isFailure = FAILURE_DLR_STATUSES.includes(status)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setErrorMessage(null)
    try {
      const body = { status, error_code: errorCode.trim() || undefined, include_unconfirmed: includeUnconfirmed }
      const res = isBulk
        ? await fetch('/api/admin/messages/dlr/bulk', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ...body, message_ids: messageIds.map(Number) }),
          })
        : await fetch(`/api/admin/messages/${messageIds[0]}/dlr`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
          })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setErrorMessage(data.error || 'Failed to send DLR')
        return
      }
      if (isBulk) {
        const result = data as BulkResult
        const warning = result.warnings
          ? ` ${result.warnings} cannot reach the customer until their DLR callback/connection is configured; retry them from the DLR Queue page afterwards.`
          : ''
        const skippedWhy = includeUnconfirmed
          ? 'already had a final DLR, failed to send, or already refunded'
          : 'already had a final DLR, no answer from the MNO (tick the option to include them), failed to send, or already refunded'
        onDone(`${status} queued for forwarding to the customer for ${result.forwarded} message(s). ${result.skipped} skipped (${skippedWhy}), ${result.failed} failed.${warning}`)
      } else {
        const warning = data.forward_warning ? ` Warning: ${data.forward_warning}.` : ''
        onDone(`${status} DLR set on ${data.segments_updated} segment(s) and queued for forwarding to the customer (queue #${data.queue_id}).${warning}`)
      }
    } catch {
      setErrorMessage('Failed to send DLR')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <p className="text-sm text-zinc-400">
        Sets the DLR on {isBulk ? `${messageIds.length} messages` : 'this message'} for segments the MNO accepted but never sent a final DLR for, then forwards it to the service provider. Segments that already have a final DLR, or that failed to send, are left unchanged.
      </p>
      <label className="flex items-start gap-2 text-sm text-zinc-300">
        <input
          type="checkbox"
          checked={includeUnconfirmed}
          onChange={(e) => setIncludeUnconfirmed(e.target.checked)}
          className="mt-1"
        />
        <span>
          Include segments with no answer from the MNO
          <span className="block text-xs text-zinc-500">Sent, but the MNO never confirmed or rejected them, so it may not have received them.</span>
        </span>
      </label>
      <div className="grid gap-4 md:grid-cols-2">
        <div>
          <label className="block text-sm font-medium text-zinc-300 mb-1">DLR Status</label>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
          >
            {DLR_STATUSES.map((s) => (
              <option key={s.value} value={s.value}>{s.label}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-zinc-300 mb-1">Error Code (optional)</label>
          <input
            type="text"
            maxLength={10}
            placeholder="e.g. 000"
            value={errorCode}
            onChange={(e) => setErrorCode(e.target.value)}
            className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white placeholder:text-zinc-500"
          />
        </div>
      </div>
      {isFailure && (
        <p className="text-sm text-amber-400">
          A failure status marks the message as failed. The charge is not refunded automatically; refund it from the provider's wallet if needed.
        </p>
      )}
      {errorMessage && <p className="text-sm text-red-400">{errorMessage}</p>}
      <div className="flex items-center gap-2">
        <Button type="submit" disabled={submitting} className="bg-white text-black hover:bg-zinc-200">
          <Send className="h-4 w-4" />
          {submitting ? 'Sending...' : 'Send DLR'}
        </Button>
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={submitting}
          className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5"
        >
          <X className="h-4 w-4" />
          Cancel
        </Button>
      </div>
    </form>
  )
}
