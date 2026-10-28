import { db } from '@/lib/db'
import { senderIds, serviceProviders, mnos, mnoConnections } from '@/lib/schema'
import { desc, eq } from 'drizzle-orm'
import { SenderIdsPage } from './client-page'

async function getSenderIdsList() {
  try {
    const senders = await db.select({
      id: senderIds.id,
      senderIdString: senderIds.senderIdString,
      serviceProviderId: senderIds.serviceProviderId,
      mnoConnectionId: senderIds.mnoConnectionId,
      status: senderIds.status,
      providerId: senderIds.providerId,
      rateLimit: senderIds.rateLimit,
      otpTemplate: senderIds.otpTemplate,
      networks: senderIds.networks,
      maxUsageCount: senderIds.maxUsageCount,
      currentUsageCount: senderIds.currentUsageCount,
      resetIntervalHours: senderIds.resetIntervalHours,
      createdAt: senderIds.createdAt,
      updatedAt: senderIds.updatedAt,
    })
      .from(senderIds)
      .orderBy(desc(senderIds.createdAt))

    return senders
  } catch (error) {
    console.error('Error fetching sender IDs:', error)
    return []
  }
}

async function getServiceProviders() {
  try {
    const spList = await db.select({
      id: serviceProviders.id,
      name: serviceProviders.name,
    })
      .from(serviceProviders)
      .orderBy(serviceProviders.name)

    return spList
  } catch (error) {
    console.error('Error fetching service providers:', error)
    return []
  }
}

async function getMnos() {
  try {
    const mnoList = await db.select({
      id: mnos.id,
      name: mnos.name,
      countryCode: mnos.countryCode,
    })
      .from(mnos)
      .where(eq(mnos.status, 'active'))
      .orderBy(mnos.name)

    return mnoList
  } catch (error) {
    console.error('Error fetching MNOs:', error)
    return []
  }
}

async function getMnoConnections() {
  try {
    const connections = await db.select({
      id: mnoConnections.id,
      mnoId: mnoConnections.mnoId,
      protocol: mnoConnections.protocol,
      status: mnoConnections.status,
      host: mnoConnections.host,
    })
      .from(mnoConnections)
      .orderBy(mnoConnections.id)

    return connections
  } catch (error) {
    console.error('Error fetching MNO connections:', error)
    return []
  }
}

export default async function Page() {
  const [initialSenders, serviceProvidersList, mnosList, mnoConnectionsList] = await Promise.all([
    getSenderIdsList(),
    getServiceProviders(),
    getMnos(),
    getMnoConnections(),
  ])

  return (
    <SenderIdsPage
      initialSenders={initialSenders}
      serviceProviders={serviceProvidersList}
      mnos={mnosList}
      mnoConnections={mnoConnectionsList}
    />
  )
}
