import Link from 'next/link'
import { Plus, Settings, ArrowLeft, RefreshCw } from 'lucide-react'
import { db } from '@/lib/db'
import { mnoConnections, mnos } from '@/lib/schema'
import { desc, eq } from 'drizzle-orm'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

async function getMNOConnections(mnoId: number) {
  try {
    return await db.select({
      id: mnoConnections.id,
      protocol: mnoConnections.protocol,
      status: mnoConnections.status,
      host: mnoConnections.host,
      port: mnoConnections.port,
      priority: mnoConnections.priority,
      createdAt: mnoConnections.createdAt,
    })
      .from(mnoConnections)
      .where(eq(mnoConnections.mnoId, mnoId))
      .orderBy(desc(mnoConnections.priority))
  } catch (error) {
    console.error('Error fetching connections:', error)
    return []
  }
}

async function getMNO(id: number) {
  try {
    const [mno] = await db.select({
      id: mnos.id,
      name: mnos.name,
      countryCode: mnos.countryCode,
    })
      .from(mnos)
      .where(eq(mnos.id, id))
    return mno
  } catch (error) {
    console.error('Error fetching MNO:', error)
    return null
  }
}

function getStatusBadge(status: string) {
  const styles: Record<string, string> = {
    active: 'bg-emerald-500/20 text-emerald-400',
    inactive: 'bg-zinc-500/20 text-zinc-400',
  }
  return (
    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.inactive}`}>
      {status}
    </span>
  )
}

export default async function MNOConnectionsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const mnoId = parseInt(id)
  
  const [mno, connections] = await Promise.all([
    getMNO(mnoId),
    getMNOConnections(mnoId)
  ])

  if (!mno) {
    return (
      <div className="p-8">
        <Link href="/admin/mnos" className="flex items-center gap-2 text-zinc-400">
          <ArrowLeft className="h-4 w-4" />
          Back to MNOs
        </Link>
        <div className="mt-8 text-center">
          <h1 className="text-2xl font-bold text-white">MNO Not Found</h1>
        </div>
      </div>
    )
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title={`${mno.name} Connections`}
        description={`+${mno.countryCode}`}
        actions={
          <Link href={`/admin/mnos/${id}/connections/new`}>
            <Button className="bg-white text-black hover:bg-zinc-200">
              <Plus className="h-4 w-4" />
              Add Connection
            </Button>
          </Link>
        }
      />

      <DataCard title="All Connections" subtitle="MNO connection configurations">
        {connections.length > 0 ? (
          <div className="divide-y divide-white/10">
            {connections.map((conn) => (
              <div key={conn.id} className="p-4 flex items-center justify-between hover:bg-white/5 transition-colors">
                <div className="flex items-center gap-4">
                  <div className="flex items-center justify-center w-12 h-12 bg-zinc-800 rounded-lg">
                    <Settings className="h-6 w-6 text-zinc-400" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-white">{conn.protocol.toUpperCase()}</span>
                      {getStatusBadge(conn.status)}
                    </div>
                    <p className="text-sm text-zinc-400">{conn.host}:{conn.port}</p>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-sm text-zinc-400">Priority: {conn.priority}</span>
                  <Link href={`/admin/mnos/${id}/connections/${conn.id}`} className="text-sm text-white hover:underline">
                    Edit
                  </Link>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="p-8 text-center">
            <Settings className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
            <p className="text-zinc-500">No connections configured</p>
            <Link href={`/admin/mnos/${id}/connections/new`} className="text-sm text-white hover:underline">
              Add your first connection
            </Link>
          </div>
        )}
      </DataCard>
    </div>
  )
}
