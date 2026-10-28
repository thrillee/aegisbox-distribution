import { NextResponse } from 'next/server'
import { db } from '@/lib/db'
import { spCredentials } from '@/lib/schema'
import { requireServiceProvider } from '@/lib/service-provider'
import { desc, eq } from 'drizzle-orm'

// The API key identifier is the bearer token the HTTP API accepts, so only a short prefix
// is ever returned here. Keys are issued and rotated by an administrator.
const VISIBLE_KEY_PREFIX = 8

export async function GET() {
  try {
    const sp = await requireServiceProvider()
    if (sp.response) return sp.response

    const credentials = await db
      .select({
        id: spCredentials.id,
        protocol: spCredentials.protocol,
        systemId: spCredentials.systemId,
        apiKeyIdentifier: spCredentials.apiKeyIdentifier,
        bindType: spCredentials.bindType,
        scope: spCredentials.scope,
        status: spCredentials.status,
        createdAt: spCredentials.createdAt,
      })
      .from(spCredentials)
      .where(eq(spCredentials.serviceProviderId, sp.serviceProviderId))
      .orderBy(desc(spCredentials.createdAt))

    return NextResponse.json({
      credentials: credentials.map(({ apiKeyIdentifier, ...credential }) => ({
        ...credential,
        apiKeyPrefix: apiKeyIdentifier ? apiKeyIdentifier.slice(0, VISIBLE_KEY_PREFIX) : null,
      })),
    })
  } catch (error) {
    console.error('Error fetching API keys:', error)
    return NextResponse.json({ error: 'Failed to fetch API keys' }, { status: 500 })
  }
}
