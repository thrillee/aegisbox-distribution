import { db } from '@/lib/db'
import { senderIdHealth, mnos } from '@/lib/schema'
import { eq, desc } from 'drizzle-orm'
import { SenderHealthTable } from './sender-health-table'
import { PageHeader } from '@/components/dashboard/page-header'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'

async function getSenderHealthData() {
  try {
    const senders = await db
      .select({
        id: senderIdHealth.id,
        senderId: senderIdHealth.senderId,
        mnoId: senderIdHealth.mnoId,
        mnoName: mnos.name,
        deliveryRate: senderIdHealth.deliveryRate,
        dlrResponseRate: senderIdHealth.dlrResponseRate,
        healthScore: senderIdHealth.healthScore,
        totalMessages: senderIdHealth.totalSent,
        isBlocked: senderIdHealth.isBlocked,
        blockedAt: senderIdHealth.blockedAt,
        blockedReason: senderIdHealth.blockedReason,
        lastMessageAt: senderIdHealth.lastAttemptAt,
        updatedAt: senderIdHealth.updatedAt,
      })
      .from(senderIdHealth)
      .leftJoin(mnos, eq(senderIdHealth.mnoId, mnos.id))
      .orderBy(desc(senderIdHealth.updatedAt))

    return senders
  } catch (error) {
    console.error('Error fetching sender health data:', error)
    return []
  }
}

async function getMNOsForFilter() {
  try {
    const mnosData = await db
      .select({
        id: mnos.id,
        name: mnos.name,
      })
      .from(mnos)
      .orderBy(mnos.name)

    return mnosData
  } catch (error) {
    console.error('Error fetching MNOs:', error)
    return []
  }
}

export default async function SenderHealthPage() {
  const [senders, mnosList] = await Promise.all([
    getSenderHealthData(),
    getMNOsForFilter(),
  ])

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Sender ID Health"
        description="Monitor delivery performance for all sender IDs across MNOs"
        actions={
          <Link
            href="/admin/routing/health"
            className="flex items-center gap-2 px-4 py-2 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to Overview
          </Link>
        }
      />

      <SenderHealthTable
        initialData={senders}
        mnos={mnosList}
      />
    </div>
  )
}
