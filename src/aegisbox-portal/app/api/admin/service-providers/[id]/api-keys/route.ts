import { db } from '@/lib/db'
import { spCredentials } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq, and } from 'drizzle-orm'
import { randomBytes } from 'crypto'

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const spId = parseInt(resolvedParams.id)

    if (isNaN(spId)) {
      return NextResponse.json({ error: 'Invalid service provider ID' }, { status: 400 })
    }

    const body = await request.json()
    const { name, scope, routingGroupId } = body

    const apiKeyIdentifier = randomBytes(32).toString('hex')
    const apiKeyHash = randomBytes(64).toString('hex')
    const prefix = apiKeyIdentifier.substring(0, 8).toUpperCase()

    const [credential] = await db
      .insert(spCredentials)
      .values({
        serviceProviderId: spId,
        protocol: 'http',
        systemId: null,
        passwordHash: null,
        bindType: null,
        scope: scope || 'default',
        apiKeyIdentifier,
        apiKeyHash,
        httpConfig: { name: name || 'API Key' },
        routingGroupId: routingGroupId || null,
        status: 'active',
      })
      .returning()

    return NextResponse.json({
      success: true,
      apiKey: {
        id: credential.id,
        name: name || 'API Key',
        identifier: apiKeyIdentifier,
        prefix,
        routingGroupId: credential.routingGroupId,
        status: credential.status,
        createdAt: credential.createdAt,
      },
    })
  } catch (error) {
    console.error('[v0] Error creating API key:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const spId = parseInt(resolvedParams.id)

    if (isNaN(spId)) {
      return NextResponse.json({ error: 'Invalid service provider ID' }, { status: 400 })
    }

    const credentials = await db
      .select({
        id: spCredentials.id,
        systemId: spCredentials.systemId,
        apiKeyIdentifier: spCredentials.apiKeyIdentifier,
        httpConfig: spCredentials.httpConfig,
        status: spCredentials.status,
        scope: spCredentials.scope,
        routingGroupId: spCredentials.routingGroupId,
        createdAt: spCredentials.createdAt,
      })
      .from(spCredentials)
      .where(
        and(
          eq(spCredentials.serviceProviderId, spId),
          eq(spCredentials.protocol, 'http')
        )
      )

    return NextResponse.json({
      apiKeys: credentials.map(c => ({
        ...c,
        prefix: c.apiKeyIdentifier?.substring(0, 8).toUpperCase() || 'UNKNOWN',
      })),
    })
  } catch (error) {
    console.error('[v0] Error fetching API keys:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
