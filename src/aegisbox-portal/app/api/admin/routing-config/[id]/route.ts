import { NextResponse } from 'next/server'
import { proxyToManagerApi } from '@/lib/manager-api'

// [id] is the routing group ID
export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  if (!/^\d+$/.test(id)) {
    return NextResponse.json({ error: 'Invalid routing group ID' }, { status: 400 })
  }
  return proxyToManagerApi(`/routing/config/${id}`, { method: 'GET' })
}

export async function PUT(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  if (!/^\d+$/.test(id)) {
    return NextResponse.json({ error: 'Invalid routing group ID' }, { status: 400 })
  }
  const body = await request.json().catch(() => null)
  if (!body) {
    return NextResponse.json({ error: 'Invalid request body' }, { status: 400 })
  }
  return proxyToManagerApi(`/routing/config/${id}`, { method: 'PUT', body })
}
