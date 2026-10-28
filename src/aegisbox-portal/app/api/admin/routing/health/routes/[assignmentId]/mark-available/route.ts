import { NextResponse } from 'next/server'
import { proxyToManagerApi } from '@/lib/manager-api'

export async function POST(
  request: Request,
  { params }: { params: Promise<{ assignmentId: string }> }
) {
  const { assignmentId } = await params
  if (!/^\d+$/.test(assignmentId)) {
    return NextResponse.json({ error: 'Invalid assignment ID' }, { status: 400 })
  }
  const body = await request.json().catch(() => null)
  if (!body) {
    return NextResponse.json({ error: 'Invalid request body' }, { status: 400 })
  }
  return proxyToManagerApi(`/routing/health/routes/${assignmentId}/mark-available`, { method: 'POST', body })
}
