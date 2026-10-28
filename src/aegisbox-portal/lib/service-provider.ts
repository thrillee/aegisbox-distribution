import { NextResponse } from 'next/server'
import { auth, clerkClient } from '@clerk/nextjs/server'
import { eq } from 'drizzle-orm'
import { db } from '@/lib/db'
import { orgMappings, serviceProviders } from '@/lib/schema'

export const NOT_LINKED_MESSAGE =
  'This organization is not linked to a service provider account. Please contact support.'

// Resolves the service provider a Clerk organization belongs to. The org_mappings table is
// the source of truth; when a mapping is missing but the Clerk organization carries
// publicMetadata.serviceProviderId (set at signup), the mapping is backfilled from it.
export async function resolveServiceProviderId(orgId: string): Promise<number | null> {
  const [mapping] = await db
    .select({ serviceProviderId: orgMappings.serviceProviderId })
    .from(orgMappings)
    .where(eq(orgMappings.clerkOrgId, orgId))
    .limit(1)
  if (mapping) return mapping.serviceProviderId

  let metadataId: number
  try {
    const client = await clerkClient()
    const org = await client.organizations.getOrganization({ organizationId: orgId })
    metadataId = Number(org.publicMetadata?.serviceProviderId)
  } catch (error) {
    console.error('Failed to read Clerk organization metadata:', error)
    return null
  }
  if (!Number.isInteger(metadataId) || metadataId <= 0) return null

  const [provider] = await db
    .select({ id: serviceProviders.id })
    .from(serviceProviders)
    .where(eq(serviceProviders.id, metadataId))
    .limit(1)
  if (!provider) return null

  await db
    .insert(orgMappings)
    .values({ clerkOrgId: orgId, serviceProviderId: provider.id })
    .onConflictDoNothing()
  return provider.id
}

// Returns the service provider of the signed in user's active organization, or null when
// there is no active organization or it is not linked to a provider.
export async function getCurrentServiceProviderId(): Promise<number | null> {
  const { orgId } = await auth()
  if (!orgId) return null
  return resolveServiceProviderId(orgId)
}

type SPAuthResult =
  | { serviceProviderId: number; response?: undefined }
  | { serviceProviderId?: undefined; response: NextResponse }

// For /api/sp routes: the caller's service provider, or the error response to return.
export async function requireServiceProvider(): Promise<SPAuthResult> {
  const { userId, orgId } = await auth()
  if (!userId || !orgId) {
    return { response: NextResponse.json({ error: 'Unauthorized' }, { status: 401 }) }
  }
  const serviceProviderId = await resolveServiceProviderId(orgId)
  if (!serviceProviderId) {
    return { response: NextResponse.json({ error: NOT_LINKED_MESSAGE }, { status: 403 }) }
  }
  return { serviceProviderId }
}

// Links a Clerk organization to a service provider: records the mapping (replacing any
// previous provider for that organization) and stores the ID in the organization's public
// metadata. The mapping is what access checks use, so a metadata failure is only logged.
export async function linkOrganizationToProvider(clerkOrgId: string, serviceProviderId: number) {
  await db
    .insert(orgMappings)
    .values({ clerkOrgId, serviceProviderId })
    .onConflictDoUpdate({
      target: orgMappings.clerkOrgId,
      set: { serviceProviderId, updatedAt: new Date() },
    })

  try {
    const client = await clerkClient()
    await client.organizations.updateOrganizationMetadata(clerkOrgId, {
      publicMetadata: { serviceProviderId },
    })
  } catch (error) {
    console.error('Failed to store serviceProviderId in Clerk organization metadata:', error)
  }
}
