import { db } from '@/lib/db'
import { spCredentials } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'
import { randomBytes } from 'crypto'
import { generateSmppPassword, hashSmppPassword, systemIdTaken, validateSmppLogin } from '@/lib/smpp-credentials'
import { smppConnectionInfo } from '@/lib/connection-info'

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
    const { systemId, password, bindType, scope, routingGroupId } = body

    const smppSystemId = systemId || `sp_${randomBytes(4).toString('hex')}`
    const smppPassword = password || generateSmppPassword()
    const invalid = validateSmppLogin(smppSystemId, smppPassword)
    if (invalid) {
      return NextResponse.json({ error: invalid }, { status: 400 })
    }
    if (await systemIdTaken(smppSystemId)) {
      return NextResponse.json({ error: `System ID ${smppSystemId} is already in use` }, { status: 400 })
    }

    const [credential] = await db
      .insert(spCredentials)
      .values({
        serviceProviderId: spId,
        protocol: 'smpp',
        systemId: smppSystemId,
        passwordHash: await hashSmppPassword(smppPassword),
        bindType: bindType || 'trx',
        scope: scope || 'default',
        routingGroupId: routingGroupId || null,
        status: 'pending',
      })
      .returning()

    return NextResponse.json({
      success: true,
      credential: {
        id: credential.id,
        systemId: credential.systemId,
        password: smppPassword,
        bindType: credential.bindType,
        scope: credential.scope,
        routingGroupId: credential.routingGroupId,
      },
      connection: smppConnectionInfo(),
    })
  } catch (error) {
    console.error('[v0] Error creating SMPP credential:', error)
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
        bindType: spCredentials.bindType,
        status: spCredentials.status,
        scope: spCredentials.scope,
        routingGroupId: spCredentials.routingGroupId,
        createdAt: spCredentials.createdAt,
      })
      .from(spCredentials)
      .where(eq(spCredentials.serviceProviderId, spId))

    return NextResponse.json({ credentials, connection: smppConnectionInfo() })
  } catch (error) {
    console.error('[v0] Error fetching SMPP credentials:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
