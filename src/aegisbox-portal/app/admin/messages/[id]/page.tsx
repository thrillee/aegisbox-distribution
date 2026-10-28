'use client'

import { useState, useEffect, useCallback } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { ArrowLeft, MessageSquare, Clock, Send, Check, X, AlertTriangle, FileCheck } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'
import { ManualDlrForm, isAwaitingFinalDlr } from '@/components/messages/manual-dlr-form'
import { formatMoney } from '@/lib/money'

interface Message {
  id: number
  clientMessageId: string | null
  clientRef: string | null
  originalSourceAddr: string
  sentSourceAddr: string | null
  originalDestinationAddr: string
  shortMessage: string
  totalSegments: number
  submittedAt: string | null
  receivedAt: string
  completedAt: string | null
  processingStatus: string
  finalStatus: string
  errorCode: string | null
  errorDescription: string | null
  cost: string
  currencyCode: string
  retryCount: number
  serviceProviderId: number
  spName: string | null
  spSystemId: string | null
  routedMnoId: number | null
  mnoName: string | null
  mnoCountryCode: string | null
}

interface Segment {
  id: number
  segmentSeqn: number
  mnoMessageId: string | null
  mnoConnectionId: number | null
  sentToMnoAt: string | null
  dlrStatus: string | null
  dlrReceivedAt: string | null
  errorCode: string | null
  errorDescription: string | null
  dlrSource: string | null
  createdAt: string
}

const STATUS_COLORS: Record<string, string> = {
  received: 'bg-blue-500/20 text-blue-400',
  queued: 'bg-amber-500/20 text-amber-400',
  processing: 'bg-purple-500/20 text-purple-400',
  submitted: 'bg-indigo-500/20 text-indigo-400',
  delivered: 'bg-emerald-500/20 text-emerald-400',
  accepted: 'bg-emerald-500/20 text-emerald-400',
  forwarded: 'bg-emerald-500/20 text-emerald-400',
  failed: 'bg-red-500/20 text-red-400',
  rejected: 'bg-red-500/20 text-red-400',
  undeliverable: 'bg-red-500/20 text-red-400',
  pending: 'bg-zinc-500/20 text-zinc-400',
}

function getStatusBadge(status: string) {
  const style = STATUS_COLORS[status] || STATUS_COLORS.pending
  return (
    <span className={`px-3 py-1 rounded-full text-sm font-medium ${style}`}>
      {status}
    </span>
  )
}

function getDlrStatusIcon(status: string | null) {
  if (!status) return <Clock className="h-4 w-4 text-zinc-500" />
  const delivered = ['DELIVRD', 'ACCEPTED', 'FORWARDED']
  const pending = ['SENT', 'SUBMITTED', 'QUEUED']
  const failed = ['FAILED', 'REJECTED', 'REJECTD', 'UNDELIV', 'UNDELIVERABLE', 'EXPIRED', 'DELETED']
  
  if (delivered.includes(status.toUpperCase())) return <Check className="h-4 w-4 text-emerald-500" />
  if (pending.includes(status.toUpperCase())) return <Clock className="h-4 w-4 text-amber-500" />
  if (failed.includes(status.toUpperCase())) return <X className="h-4 w-4 text-red-500" />
  return <AlertTriangle className="h-4 w-4 text-orange-500" />
}

function formatCurrency(amount: string, currency: string) {
  return formatMoney(amount, currency)
}

export default function MessageDetailPage() {
  const params = useParams()
  const messageId = params.id as string
  
  const [message, setMessage] = useState<Message | null>(null)
  const [segments, setSegments] = useState<Segment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showDlrForm, setShowDlrForm] = useState(false)
  const [dlrNotice, setDlrNotice] = useState<string | null>(null)

  const fetchMessage = useCallback(async () => {
    try {
      const res = await fetch(`/api/admin/messages/${messageId}`)
      if (res.ok) {
        const data = await res.json()
        setMessage(data.message)
        setSegments(data.segments || [])
      } else {
        setError('Message not found')
      }
    } catch (err) {
      setError('Failed to load message')
    } finally {
      setLoading(false)
    }
  }, [messageId])

  useEffect(() => {
    if (messageId) fetchMessage()
  }, [messageId, fetchMessage])

  const isClosed = message ? ['failed', 'rejected', 'expired'].includes(message.finalStatus) : false
  const hasPendingDlr = !isClosed && segments.some(isAwaitingFinalDlr)

  if (loading) {
    return <PageLoadingSkeleton />
  }

  if (error || !message) {
    return (
      <div className="p-8 text-center">
        <h1 className="text-2xl font-bold text-white mb-4">{error || 'Message Not Found'}</h1>
        <Link href="/admin/messages" className="text-white hover:underline">
          Back to Messages
        </Link>
      </div>
    )
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Message Details"
        description={`Message ID: ${message.id}`}
        actions={
          <div className="flex items-center gap-2">
            {hasPendingDlr && !showDlrForm && (
              <Button
                onClick={() => {
                  setDlrNotice(null)
                  setShowDlrForm(true)
                }}
                className="bg-white text-black hover:bg-zinc-200"
              >
                <FileCheck className="h-4 w-4" />
                Send DLR
              </Button>
            )}
            <Link href="/admin/messages">
              <Button variant="outline" className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5">
                <ArrowLeft className="h-4 w-4" />
                Back to Messages
              </Button>
            </Link>
          </div>
        }
      />

      {dlrNotice && (
        <div className="rounded-2xl border border-emerald-500/20 bg-emerald-500/10 p-4 text-sm text-emerald-400">
          {dlrNotice}
        </div>
      )}

      {showDlrForm && (
        <DataCard title="Send Manual DLR" subtitle="Use when the MNO did not return a DLR for this message">
          <ManualDlrForm
            messageIds={[message.id]}
            onCancel={() => setShowDlrForm(false)}
            onDone={(summary) => {
              setShowDlrForm(false)
              setDlrNotice(summary)
              fetchMessage()
            }}
          />
        </DataCard>
      )}

      <div className="grid gap-6 md:grid-cols-2">
        <DataCard title="Message Information" subtitle="Message details and status">
          <div className="space-y-4">
            <div className="flex justify-between">
              <span className="text-zinc-400">Status</span>
              {getStatusBadge(message.processingStatus)}
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-400">Client Ref</span>
              <span className="text-white">{message.clientRef || '-'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-400">Segments</span>
              <span className="text-white">{message.totalSegments}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-400">Cost</span>
              <span className="font-medium text-white">{formatCurrency(message.cost, message.currencyCode)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-400">Retry Count</span>
              <span className="text-white">{message.retryCount}</span>
            </div>
          </div>
        </DataCard>

        <DataCard title="Routing" subtitle="Message routing information">
          <div className="space-y-4">
            <div className="flex justify-between">
              <span className="text-zinc-400">Service Provider</span>
              <span className="text-white">{message.spName || '-'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-400">SP System ID</span>
              <span className="text-white font-mono">{message.spSystemId || '-'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-400">MNO</span>
              <span className="text-white">{message.mnoName || '-'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-400">MNO Country</span>
              <span className="text-white">+{message.mnoCountryCode || '-'}</span>
            </div>
          </div>
        </DataCard>
      </div>

      <DataCard title="Message Content" subtitle="Message text and recipients">
        <div className="grid gap-4 md:grid-cols-2 mb-4">
          <div>
            <span className="text-sm text-zinc-400">From</span>
            <p className="text-lg font-medium text-white">{message.originalSourceAddr}</p>
            {message.sentSourceAddr && message.sentSourceAddr !== message.originalSourceAddr && (
              <p className="text-sm text-zinc-400">Sent as {message.sentSourceAddr}</p>
            )}
          </div>
          <div>
            <span className="text-sm text-zinc-400">To</span>
            <p className="text-lg font-medium text-white">{message.originalDestinationAddr}</p>
          </div>
        </div>
        <div className="bg-zinc-800/50 rounded-lg p-4">
          <p className="text-white whitespace-pre-wrap">{message.shortMessage}</p>
        </div>
      </DataCard>

      {message.errorCode && (
        <div className="rounded-2xl border border-red-500/20 bg-red-500/10 p-6">
          <h3 className="text-lg font-semibold text-red-400 mb-2">Error</h3>
          <p className="text-red-400">{message.errorCode}: {message.errorDescription || 'No description'}</p>
        </div>
      )}

      <DataCard title="Message Segments" subtitle="Individual message segments and delivery status">
        {segments.length === 0 ? (
          <p className="text-zinc-500">No segments found</p>
        ) : (
          <div className="space-y-4">
            {segments.map((segment) => (
              <div key={segment.id} className="border border-white/10 rounded-lg p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium text-white">Segment {segment.segmentSeqn}</span>
                  <div className="flex items-center gap-2">
                    {getDlrStatusIcon(segment.dlrStatus)}
                    <span className="text-sm text-zinc-400">{segment.dlrStatus || 'Pending'}</span>
                    {segment.dlrSource && segment.dlrSource !== 'mno' && (
                      <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-amber-500/20 text-amber-400">
                        {segment.dlrSource === 'auto' ? 'Auto DLR' : 'Manual DLR'}
                      </span>
                    )}
                  </div>
                </div>
                <div className="grid gap-2 md:grid-cols-4 text-sm">
                  <div>
                    <span className="text-zinc-500">MNO Message ID</span>
                    <p className="font-mono text-white">{segment.mnoMessageId || '-'}</p>
                  </div>
                  <div>
                    <span className="text-zinc-500">Sent to MNO</span>
                    <p className="text-white">{segment.sentToMnoAt ? new Date(segment.sentToMnoAt).toLocaleString() : '-'}</p>
                  </div>
                  <div>
                    <span className="text-zinc-500">DLR Received</span>
                    <p className="text-white">{segment.dlrReceivedAt ? new Date(segment.dlrReceivedAt).toLocaleString() : '-'}</p>
                  </div>
                  <div>
                    <span className="text-zinc-500">Created</span>
                    <p className="text-white">{new Date(segment.createdAt).toLocaleString()}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </DataCard>

      <div className="grid gap-4 md:grid-cols-3">
        <DataCard title="">
          <div className="flex items-center gap-3 mb-2">
            <MessageSquare className="h-5 w-5 text-blue-400" />
            <span className="text-sm text-zinc-400">Received</span>
          </div>
          <p className="text-lg font-medium text-white">{new Date(message.receivedAt).toLocaleString()}</p>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3 mb-2">
            <Send className="h-5 w-5 text-indigo-400" />
            <span className="text-sm text-zinc-400">Submitted</span>
          </div>
          <p className="text-lg font-medium text-white">{message.submittedAt ? new Date(message.submittedAt).toLocaleString() : '-'}</p>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3 mb-2">
            <Check className="h-5 w-5 text-emerald-400" />
            <span className="text-sm text-zinc-400">Completed</span>
          </div>
          <p className="text-lg font-medium text-white">{message.completedAt ? new Date(message.completedAt).toLocaleString() : '-'}</p>
        </DataCard>
      </div>
    </div>
  )
}
