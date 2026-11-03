import { NextResponse } from 'next/server'
import { auth } from '@clerk/nextjs/server'
import { MANAGER_API_URL, managerApiHeaders } from '@/lib/manager-api'

// Streams a scan screenshot from the gateway (the screenshot volume lives with the gateway).
export async function GET(_request: Request, { params }: { params: Promise<{ scanId: string }> }) {
  const { userId } = await auth()
  if (!userId) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  const { scanId } = await params
  if (!/^\d+$/.test(scanId)) return NextResponse.json({ error: 'Invalid scan ID' }, { status: 400 })
  try {
    const res = await fetch(`${MANAGER_API_URL}/api/v1/link-scans/${scanId}/screenshot`, {
      headers: managerApiHeaders(),
      cache: 'no-store',
    })
    if (!res.ok || res.headers.get('content-type') !== 'image/jpeg') {
      return NextResponse.json({ error: 'No screenshot' }, { status: 404 })
    }
    const body = await res.arrayBuffer()
    return new NextResponse(body, {
      headers: {
        'Content-Type': 'image/jpeg',
        'Cache-Control': 'private, no-store',
        'X-Content-Type-Options': 'nosniff',
      },
    })
  } catch {
    return NextResponse.json({ error: 'Gateway unreachable' }, { status: 502 })
  }
}
