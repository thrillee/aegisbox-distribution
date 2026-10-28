import { proxyToManagerApi } from '@/lib/manager-api'

export async function GET() {
  return proxyToManagerApi('/routing/config', { method: 'GET' })
}

export async function POST(request: Request) {
  const body = await request.json().catch(() => ({}))
  return proxyToManagerApi('/routing/config', { method: 'POST', body })
}
