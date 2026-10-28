import { NextResponse } from 'next/server'
import { auth } from '@clerk/nextjs/server'

export const MANAGER_API_URL = process.env.MANAGER_API_URL || 'http://localhost:8081'

// Server side only: MANAGER_API_KEY must never reach browser code.
export function managerApiHeaders(): Record<string, string> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (process.env.MANAGER_API_KEY) {
    headers['X-API-Key'] = process.env.MANAGER_API_KEY
  }
  return headers
}

export async function proxyToManagerApi(
  path: string,
  init: { method: 'GET' | 'POST' | 'PUT'; body?: unknown },
): Promise<NextResponse> {
  const { userId } = await auth()
  if (!userId) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  try {
    const response = await fetch(`${MANAGER_API_URL}/api/v1${path}`, {
      method: init.method,
      headers: managerApiHeaders(),
      body: init.body === undefined ? undefined : JSON.stringify(init.body),
      cache: 'no-store',
    })
    const text = await response.text()
    let data: unknown = {}
    try {
      data = text ? JSON.parse(text) : {}
    } catch {
      data = { error: text || 'Unexpected response from Manager API' }
    }
    if (!response.ok) {
      console.error('Manager API error:', path, response.status, text)
    }
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error('Error calling Manager API:', path, error)
    return NextResponse.json({ error: 'Manager API is unreachable' }, { status: 502 })
  }
}
