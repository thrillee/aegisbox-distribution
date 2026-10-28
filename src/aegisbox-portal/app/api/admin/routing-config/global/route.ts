import { NextResponse } from 'next/server'
import { proxyToManagerApi } from '@/lib/manager-api'

export async function GET() {
  return proxyToManagerApi('/routing/config/global', { method: 'GET' })
}

export async function PUT(request: Request) {
  const body = await request.json().catch(() => null)
  if (!body) {
    return NextResponse.json({ error: 'Invalid request body' }, { status: 400 })
  }
  return proxyToManagerApi('/routing/config/global', { method: 'PUT', body })
}
