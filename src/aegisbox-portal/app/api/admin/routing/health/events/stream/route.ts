import { NextResponse } from 'next/server'
import { auth } from '@clerk/nextjs/server'
import { MANAGER_API_URL, managerApiHeaders } from '@/lib/manager-api'

export const dynamic = 'force-dynamic'

// Relays the manager API's health event stream (server-sent events) to the browser,
// so the browser never needs the manager API address or key.
export async function GET(request: Request) {
  const { userId } = await auth()
  if (!userId) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  try {
    const upstream = await fetch(`${MANAGER_API_URL}/api/v1/routing/health/events/stream`, {
      headers: { ...managerApiHeaders(), Accept: 'text/event-stream' },
      cache: 'no-store',
      signal: request.signal,
    })
    if (!upstream.ok || !upstream.body) {
      console.error('Manager API health stream error:', upstream.status)
      return NextResponse.json({ error: 'Health stream unavailable' }, { status: upstream.status || 502 })
    }
    return new Response(upstream.body, {
      headers: {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache, no-transform',
        Connection: 'keep-alive',
        'X-Accel-Buffering': 'no',
      },
    })
  } catch (error) {
    console.error('Error connecting to manager API health stream:', error)
    return NextResponse.json({ error: 'Manager API is unreachable' }, { status: 502 })
  }
}
