import Link from 'next/link'
import { db } from '@/lib/db'
import { mnos, mnoConnections } from '@/lib/schema'
import { desc, eq } from 'drizzle-orm'
import { MNOsList } from './mnos-list'

async function getMNOs() {
  try {
    const mnoList = await db.select({
      id: mnos.id,
      name: mnos.name,
      countryCode: mnos.countryCode,
      status: mnos.status,
      createdAt: mnos.createdAt,
    })
      .from(mnos)
      .orderBy(desc(mnos.name))

    const results = []
    for (const mno of mnoList) {
      const connections = await db.select({
        id: mnoConnections.id,
        status: mnoConnections.status,
      })
        .from(mnoConnections)
        .where(eq(mnoConnections.mnoId, mno.id))

      const activeConnections = connections.filter(c => c.status === 'active').length

      results.push({
        id: mno.id,
        name: mno.name,
        countryCode: mno.countryCode,
        status: mno.status,
        connectionsCount: connections.length,
        activeConnections,
        lastActivity: 'N/A',
        dailyMessages: 0,
        deliveryRate: 0,
      })
    }

    return results
  } catch (error) {
    console.error('Error fetching MNOs:', error)
    return []
  }
}

export default async function MNOsPage() {
  const initialMNOs = await getMNOs()

  return <MNOsList initialMNOs={initialMNOs} />
}
