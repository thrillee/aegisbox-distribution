import { NextResponse } from 'next/server'
import { proxyToManagerApi } from '@/lib/manager-api'

export async function POST(
  request: Request,
  { params }: { params: Promise<{ senderId: string; mnoId: string }> }
) {
  const { senderId, mnoId } = await params
  if (!/^\d+$/.test(senderId) || !/^\d+$/.test(mnoId)) {
    return NextResponse.json({ error: 'Invalid sender or MNO ID' }, { status: 400 })
  }
  const body = await request.json().catch(() => ({}))
  return proxyToManagerApi(`/routing/health/senders/${senderId}/${mnoId}/unblock`, { method: 'POST', body })
}
