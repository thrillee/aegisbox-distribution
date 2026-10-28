import { db } from '@/lib/db'
import { serviceProviders, wallets, spCredentials } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const id = parseInt(resolvedParams.id)

    if (isNaN(id)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const [provider] = await db
      .select({
        id: serviceProviders.id,
        name: serviceProviders.name,
        email: serviceProviders.email,
        status: serviceProviders.status,
        defaultCurrencyCode: serviceProviders.defaultCurrencyCode,
        createdAt: serviceProviders.createdAt,
        updatedAt: serviceProviders.updatedAt,
      })
      .from(serviceProviders)
      .where(eq(serviceProviders.id, id))
      .limit(1)

    if (!provider) {
      return NextResponse.json({ error: 'Service provider not found' }, { status: 404 })
    }

    const [wallet] = await db
      .select({
        id: wallets.id,
        balance: wallets.balance,
        currencyCode: wallets.currencyCode,
      })
      .from(wallets)
      .where(eq(wallets.serviceProviderId, id))
      .limit(1)

    const credentials = await db
      .select({
        id: spCredentials.id,
        systemId: spCredentials.systemId,
        protocol: spCredentials.protocol,
        status: spCredentials.status,
        bindType: spCredentials.bindType,
        apiKeyIdentifier: spCredentials.apiKeyIdentifier,
        httpConfig: spCredentials.httpConfig,
        scope: spCredentials.scope,
        routingGroupId: spCredentials.routingGroupId,
        createdAt: spCredentials.createdAt,
      })
      .from(spCredentials)
      .where(eq(spCredentials.serviceProviderId, id))

    const apiKeys = credentials.filter(c => c.protocol === 'http' || !c.systemId)
    const smppCreds = credentials.filter(c => c.protocol === 'smpp' && c.systemId)

    return NextResponse.json({
      provider: {
        ...provider,
        walletBalance: wallet?.balance || 0,
        walletCurrency: wallet?.currencyCode || provider.defaultCurrencyCode,
      },
      credentials: {
        apiKeys,
        smpps: smppCreds,
      },
    })
  } catch (error) {
    console.error('[v0] Error fetching service provider:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
