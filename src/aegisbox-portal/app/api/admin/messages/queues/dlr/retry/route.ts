import { NextResponse } from 'next/server'
import { proxyToManagerApi } from '@/lib/manager-api'

export async function POST(request: Request) {
  const body = await request.json().catch(() => null)
  if (!body || !Array.isArray(body.ids)) {
    return NextResponse.json({ error: 'Invalid request body' }, { status: 400 })
  }
  return proxyToManagerApi('/dlr-forwarding/retry', {
    method: 'POST',
    body: { ids: body.ids.map(Number) },
  })
}
