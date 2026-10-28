import { db } from '@/lib/db'
import { otpAlternativeSenders, serviceProviders, mnos } from '@/lib/schema'
import { desc, eq } from 'drizzle-orm'
import { AlternativeSendersPage } from './client-page'

async function getAlternativeSenders() {
  try {
    const senders = await db.select({
      id: otpAlternativeSenders.id,
      senderIdString: otpAlternativeSenders.senderIdString,
      serviceProviderId: otpAlternativeSenders.serviceProviderId,
      mnoId: otpAlternativeSenders.mnoId,
      status: otpAlternativeSenders.status,
      currentUsageCount: otpAlternativeSenders.currentUsageCount,
      maxUsageCount: otpAlternativeSenders.maxUsageCount,
      resetIntervalHours: otpAlternativeSenders.resetIntervalHours,
      lastResetAt: otpAlternativeSenders.lastResetAt,
      lastUsedAt: otpAlternativeSenders.lastUsedAt,
      notes: otpAlternativeSenders.notes,
      createdAt: otpAlternativeSenders.createdAt,
      updatedAt: otpAlternativeSenders.updatedAt,
    })
      .from(otpAlternativeSenders)
      .orderBy(desc(otpAlternativeSenders.createdAt))

    return senders
  } catch (error) {
    console.error('Error fetching OTP alternative senders:', error)
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

export default async function Page() {
  const [initialSenders, serviceProvidersList, mnosList] = await Promise.all([
    getAlternativeSenders(),
    getServiceProviders(),
    getMnos(),
  ])

  return (
    <AlternativeSendersPage
      initialSenders={initialSenders}
      serviceProviders={serviceProvidersList}
      mnos={mnosList}
    />
  )
}
