import { NextResponse } from 'next/server'
import { proxyToManagerApi } from '@/lib/manager-api'

// The sender IDs this MNO connection accepts, used for providers on the route sender pool.
export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  if (!/^\d+$/.test(id)) {
    return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
  }
  return proxyToManagerApi(`/mno-connections/${id}/route-senders`, { method: 'GET' })
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  if (!/^\d+$/.test(id)) {
    return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
  }
  const body = await request.json().catch(() => null)
  if (!body) {
    return NextResponse.json({ error: 'Invalid request body' }, { status: 400 })
  }
  return proxyToManagerApi(`/mno-connections/${id}/route-senders`, { method: 'POST', body })
}
