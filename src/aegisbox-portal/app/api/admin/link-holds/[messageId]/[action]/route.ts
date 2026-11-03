import { NextResponse } from 'next/server'
import { proxyToManagerApi } from '@/lib/manager-api'
import { actorName } from '@/lib/link-scans'

// Releases (sends) or rejects a message held by the link check. The gateway does the work: it
// sends the message on, or rejects it with a REJECTD DLR and the refund the customer's settings say.
export async function POST(_request: Request, { params }: { params: Promise<{ messageId: string; action: string }> }) {
  const { messageId, action } = await params
  if (!/^\d+$/.test(messageId)) return NextResponse.json({ error: 'Invalid message ID' }, { status: 400 })
  if (action !== 'release' && action !== 'reject') return NextResponse.json({ error: 'Unknown action' }, { status: 404 })
  return proxyToManagerApi(`/link-scan-holds/${messageId}/${action}`, { method: 'POST', body: { by: await actorName() } })
}
