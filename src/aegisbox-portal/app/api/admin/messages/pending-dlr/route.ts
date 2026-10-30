import { NextResponse } from 'next/server'
import { auth } from '@clerk/nextjs/server'
import { MANAGER_API_URL, managerApiHeaders } from '@/lib/manager-api'
import { isReportDate } from '@/lib/time'
import { pendingDlrMessageIds, pendingDlrSummary, type PendingDlrFilter } from '@/lib/pending-dlr'

const STATUSES = ['DELIVRD', 'UNDELIV', 'EXPIRED', 'REJECTD', 'DELETED']
const BATCH = 200 // the gateway's bulk DLR limit
const MAX_MESSAGES = 20000

function parseFilter(src: Record<string, unknown>): PendingDlrFilter | string {
  const sp = Number(src.sp_id)
  const from = String(src.from ?? '')
  const to = String(src.to ?? '')
  const older = src.older_than_minutes === undefined || src.older_than_minutes === '' ? 60 : Number(src.older_than_minutes)
  if (!Number.isInteger(sp) || sp <= 0) return 'Choose a customer'
  if (!isReportDate(from) || !isReportDate(to)) return 'Dates must be YYYY-MM-DD'
  if (to < from) return 'The end date is before the start date'
  if (!Number.isInteger(older) || older < 0 || older > 10080) return 'Minutes must be a whole number from 0 to 10080'
  return { serviceProviderId: sp, from, to, olderThanMinutes: older }
}

export async function GET(request: Request) {
  const filter = parseFilter(Object.fromEntries(new URL(request.url).searchParams))
  if (typeof filter === 'string') return NextResponse.json({ error: filter }, { status: 400 })
  try {
    return NextResponse.json(await pendingDlrSummary(filter))
  } catch (error) {
    console.error('Error counting pending DLRs:', error)
    return NextResponse.json({ error: 'Failed to count messages waiting for a DLR' }, { status: 500 })
  }
}

// Sends the chosen DLR status to every matching message, in batches through the gateway's bulk
// manual DLR (which checks each message again and queues the DLR to the customer).
export async function POST(request: Request) {
  const { userId } = await auth()
  if (!userId) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })

  const body = await request.json().catch(() => null)
  if (!body || typeof body !== 'object') return NextResponse.json({ error: 'Invalid request body' }, { status: 400 })
  const filter = parseFilter(body)
  if (typeof filter === 'string') return NextResponse.json({ error: filter }, { status: 400 })
  const status = String(body.status ?? '').toUpperCase()
  if (!STATUSES.includes(status)) return NextResponse.json({ error: 'Choose a DLR status' }, { status: 400 })
  const includeNoAnswer = body.include_no_answer === true

  try {
    const ids = await pendingDlrMessageIds(filter, includeNoAnswer)
    if (ids.length === 0) return NextResponse.json({ error: 'No messages match' }, { status: 400 })
    if (ids.length > MAX_MESSAGES) {
      return NextResponse.json({ error: `${ids.length} messages match; narrow the dates to at most ${MAX_MESSAGES}` }, { status: 400 })
    }

    const totals = { matched: ids.length, forwarded: 0, skipped: 0, failed: 0, warnings: 0 }
    for (let i = 0; i < ids.length; i += BATCH) {
      const res = await fetch(`${MANAGER_API_URL}/api/v1/messages/dlr/bulk`, {
        method: 'POST',
        headers: managerApiHeaders(),
        body: JSON.stringify({ status, include_unconfirmed: includeNoAnswer, message_ids: ids.slice(i, i + BATCH) }),
        cache: 'no-store',
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        // Batches already sent stay sent; report how far it got.
        return NextResponse.json(
          { ...totals, error: `Stopped after ${i} of ${ids.length} messages: ${data.error || `gateway returned ${res.status}`}` },
          { status: 502 },
        )
      }
      totals.forwarded += Number(data.forwarded ?? 0)
      totals.skipped += Number(data.skipped ?? 0)
      totals.failed += Number(data.failed ?? 0)
      totals.warnings += Number(data.warnings ?? 0)
    }
    return NextResponse.json(totals)
  } catch (error) {
    console.error('Error sending pending DLRs:', error)
    return NextResponse.json({ error: 'Failed to send DLRs; check the DLR Queue for what was queued' }, { status: 500 })
  }
}
