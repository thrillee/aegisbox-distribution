import { NextRequest, NextResponse } from 'next/server'
import { clerkClient } from '@clerk/nextjs/server'
import { db } from '@/lib/db'
import { orgMappings, serviceProviders } from '@/lib/schema'
import { linkOrganizationToProvider } from '@/lib/service-provider'
import { and, eq } from 'drizzle-orm'

async function parseProviderId(params: Promise<{ id: string }>) {
  const id = parseInt((await params).id)
  return Number.isInteger(id) && id > 0 ? id : null
}

// Lists the Clerk organizations whose members sign in to the portal as this provider.
export async function GET(_request: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  try {
    const providerId = await parseProviderId(params)
    if (!providerId) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const links = await db
      .select({ clerkOrgId: orgMappings.clerkOrgId, createdAt: orgMappings.createdAt })
      .from(orgMappings)
      .where(eq(orgMappings.serviceProviderId, providerId))

    const client = await clerkClient()
    const organizations = await Promise.all(
      links.map(async (link) => {
        try {
          const org = await client.organizations.getOrganization({ organizationId: link.clerkOrgId })
          return { ...link, name: org.name, slug: org.slug }
        } catch {
          return { ...link, name: null, slug: null }
        }
      })
    )

    return NextResponse.json({ organizations })
  } catch (error) {
    console.error('Error fetching provider organizations:', error)
    return NextResponse.json({ error: 'Failed to fetch linked organizations' }, { status: 500 })
  }
}

// Links a Clerk organization to this provider, e.g. for providers created before the
// mapping was recorded automatically. An organization belongs to one provider only, so
// linking it here moves it away from any provider it was linked to before.
export async function PUT(request: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  try {
    const providerId = await parseProviderId(params)
    if (!providerId) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const body = await request.json().catch(() => ({}))
    const clerkOrgId = typeof body.clerk_org_id === 'string' ? body.clerk_org_id.trim() : ''
    if (!clerkOrgId.startsWith('org_')) {
      return NextResponse.json({ error: 'clerk_org_id must be a Clerk organization ID (org_...)' }, { status: 400 })
    }
    if (clerkOrgId === process.env.CLERK_ADMIN_ORG_ID) {
      return NextResponse.json({ error: 'The admin organization cannot be linked to a provider' }, { status: 400 })
    }

    const [provider] = await db
      .select({ id: serviceProviders.id })
      .from(serviceProviders)
      .where(eq(serviceProviders.id, providerId))
      .limit(1)
    if (!provider) {
      return NextResponse.json({ error: 'Service provider not found' }, { status: 404 })
    }

    try {
      const client = await clerkClient()
      await client.organizations.getOrganization({ organizationId: clerkOrgId })
    } catch {
      return NextResponse.json({ error: 'Clerk organization not found' }, { status: 404 })
    }

    await linkOrganizationToProvider(clerkOrgId, providerId)
    return NextResponse.json({ success: true, clerk_org_id: clerkOrgId })
  } catch (error) {
    console.error('Error linking organization:', error)
    return NextResponse.json({ error: 'Failed to link organization' }, { status: 500 })
  }
}

// Unlinks a Clerk organization from this provider; its members lose portal access to it.
export async function DELETE(request: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  try {
    const providerId = await parseProviderId(params)
    if (!providerId) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const clerkOrgId = new URL(request.url).searchParams.get('clerk_org_id')
    if (!clerkOrgId) {
      return NextResponse.json({ error: 'clerk_org_id is required' }, { status: 400 })
    }

    const [link] = await db
      .select({ id: orgMappings.id })
      .from(orgMappings)
      .where(and(eq(orgMappings.serviceProviderId, providerId), eq(orgMappings.clerkOrgId, clerkOrgId)))
      .limit(1)
    if (!link) {
      return NextResponse.json({ error: 'Organization is not linked to this provider' }, { status: 404 })
    }

    // Clear the metadata first: the portal backfills a missing mapping from it, so leaving it
    // behind would silently relink the organization.
    try {
      const client = await clerkClient()
      await client.organizations.updateOrganizationMetadata(clerkOrgId, {
        publicMetadata: { serviceProviderId: null },
      })
    } catch (error) {
      console.error('Failed to clear serviceProviderId from Clerk organization metadata:', error)
      return NextResponse.json({ error: 'Failed to update the Clerk organization; nothing was unlinked' }, { status: 502 })
    }

    await db.delete(orgMappings).where(eq(orgMappings.id, link.id))

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('Error unlinking organization:', error)
    return NextResponse.json({ error: 'Failed to unlink organization' }, { status: 500 })
  }
}
