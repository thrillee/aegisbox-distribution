import { NextResponse } from 'next/server'
import { proxyToManagerApi } from '@/lib/manager-api'

export async function POST(request: Request) {
  const body = await request.json().catch(() => null)
  if (!body) {
    return NextResponse.json({ error: 'Invalid request body' }, { status: 400 })
  }
  return proxyToManagerApi('/messages/dlr/bulk', { method: 'POST', body })
}
