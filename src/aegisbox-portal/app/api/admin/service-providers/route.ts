import { db } from '@/lib/db'
import { serviceProviders, wallets, spCredentials } from '@/lib/schema'
import { NextResponse } from 'next/server'
import crypto from 'crypto'
import { generateSmppPassword, hashSmppPassword, systemIdTaken, validateSmppLogin } from '@/lib/smpp-credentials'
import { smppConnectionInfo } from '@/lib/connection-info'
import { linkOrganizationToProvider } from '@/lib/service-provider'

export async function GET() {
  try {
    const providers = await db
      .select({
        id: serviceProviders.id,
        name: serviceProviders.name,
        email: serviceProviders.email,
        status: serviceProviders.status,
        defaultCurrencyCode: serviceProviders.defaultCurrencyCode,
        createdAt: serviceProviders.createdAt,
      })
      .from(serviceProviders)
      .orderBy(serviceProviders.createdAt)

    return NextResponse.json({ providers })
  } catch (error) {
    console.error('[v0] Error fetching providers:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

async function createClerkOrganization(name: string, adminEmail: string) {
  const clerkSecretKey = process.env.CLERK_SECRET_KEY
  if (!clerkSecretKey) {
    throw new Error('CLERK_SECRET_KEY is not configured')
  }

  const baseSlug = name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '')
  const timestamp = Date.now().toString(36)
  const slug = `${baseSlug}-${timestamp}`

  const response = await fetch('https://api.clerk.com/v1/organizations', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${clerkSecretKey}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      name,
      slug,
    }),
  })

  if (!response.ok) {
    const error = await response.json()
    console.error('[Clerk] Error creating organization:', error)
    throw new Error(error.errors?.[0]?.message || 'Failed to create Clerk organization')
  }

  const org = await response.json()
  return org.id
}

async function inviteUserToOrganization(orgId: string, email: string, firstName?: string, lastName?: string) {
  const clerkSecretKey = process.env.CLERK_SECRET_KEY
  if (!clerkSecretKey) {
    throw new Error('CLERK_SECRET_KEY is not configured')
  }

  const response = await fetch('https://api.clerk.com/v1/organization_invitations', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${clerkSecretKey}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      organization_id: orgId,
      email_address: email,
      role: 'org:admin',
      inviter_name: firstName || 'Admin',
      redirect_url: `${process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000'}/sp`,
    }),
  })

  if (!response.ok) {
    const error = await response.json()
    console.error('[Clerk] Error creating invitation:', error)
    throw new Error(error.errors?.[0]?.message || 'Failed to create invitation')
  }

  const invitation = await response.json()
  return invitation
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { name, email, currency, adminFirstName, adminLastName, credentials } = body

    if (!name || !email) {
      return NextResponse.json({ error: 'Name and email are required' }, { status: 400 })
    }

    // Check the SMPP login before creating anything, so a bad value leaves nothing half made.
    const smppSystemId = credentials?.smpp?.systemId || `sp_${crypto.randomBytes(4).toString('hex')}`
    const smppPassword = credentials?.smpp?.password || generateSmppPassword()
    if (credentials?.smpp) {
      const invalid = validateSmppLogin(smppSystemId, smppPassword)
      if (invalid) {
        return NextResponse.json({ error: invalid }, { status: 400 })
      }
      if (await systemIdTaken(smppSystemId)) {
        return NextResponse.json({ error: `System ID ${smppSystemId} is already in use` }, { status: 400 })
      }
    }

    let clerkOrgId = null
    let clerkInvitationId = null

    try {
      clerkOrgId = await createClerkOrganization(name, email)
      console.log(`[v0] Created Clerk organization: ${clerkOrgId}`)

      const invitation = await inviteUserToOrganization(clerkOrgId, email, adminFirstName, adminLastName)
      clerkInvitationId = invitation.id
      console.log(`[v0] Created Clerk invitation: ${clerkInvitationId}`)
    } catch (clerkError: any) {
      console.error('[v0] Clerk integration error:', clerkError.message)
    }

    const providerResult = await db
      .insert(serviceProviders)
      .values({
        name,
        email,
        defaultCurrencyCode: currency || 'NGN',
        status: 'pending',
      })
      .returning()

    const provider = providerResult[0]

    await db.insert(wallets).values({
      serviceProviderId: provider.id,
      currencyCode: currency || 'NGN',
      balance: '0.000000',
    })

    if (clerkOrgId) {
      await linkOrganizationToProvider(clerkOrgId, provider.id)
    }

    const createdCredentials = []

    if (credentials?.smpp) {
      const { bindType, routingGroupId } = credentials.smpp

      const [smppCred] = await db.insert(spCredentials).values({
        serviceProviderId: provider.id,
        protocol: 'smpp',
        status: 'pending',
        systemId: smppSystemId,
        passwordHash: await hashSmppPassword(smppPassword),
        bindType: bindType || 'trx',
        routingGroupId: routingGroupId || null,
        scope: 'default',
      }).returning()

      createdCredentials.push({
        type: 'smpp',
        id: smppCred.id,
        systemId: smppSystemId,
        password: smppPassword, // shown once; only the hash is stored
        bindType: bindType || 'trx',
      })
    }

    if (credentials?.http) {
      const { name: httpName, routingGroupId } = credentials.http
      const apiKeyIdentifier = crypto.randomBytes(32).toString('hex')
      const apiKeyHash = crypto.randomBytes(64).toString('hex')

      const [httpCred] = await db.insert(spCredentials).values({
        serviceProviderId: provider.id,
        protocol: 'http',
        status: 'active',
        systemId: null,
        passwordHash: null,
        apiKeyHash,
        apiKeyIdentifier,
        httpConfig: { name: httpName || name + ' API Key' },
        routingGroupId: routingGroupId || null,
        scope: 'default',
      }).returning()

      createdCredentials.push({
        type: 'http',
        id: httpCred.id,
        identifier: apiKeyIdentifier,
        prefix: apiKeyIdentifier.substring(0, 8).toUpperCase(),
      })
    }

    return NextResponse.json({
      success: true,
      provider: {
        id: provider.id,
        name: provider.name,
        email: provider.email,
      },
      credentials: createdCredentials,
      connection: smppConnectionInfo(),
      clerk: {
        organizationId: clerkOrgId,
        invitationId: clerkInvitationId,
      },
    })
  } catch (error: any) {
    console.error('[v0] Error creating provider:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'A provider with this email already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: error.message || 'Failed to create service provider' }, { status: 500 })
  }
}
