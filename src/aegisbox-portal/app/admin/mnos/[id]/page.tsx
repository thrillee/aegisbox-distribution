import Link from 'next/link'
import { ArrowLeft, Plus, Settings, RefreshCw, Radio, Globe } from 'lucide-react'
import { db } from '@/lib/db'
import { mnos, mnoConnections } from '@/lib/schema'
import { desc, eq } from 'drizzle-orm'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

async function getMNO(id: string) {
  try {
    const [mno] = await db.select({
      id: mnos.id,
      name: mnos.name,
      countryCode: mnos.countryCode,
      networkCode: mnos.networkCode,
      status: mnos.status,
      createdAt: mnos.createdAt,
    })
      .from(mnos)
      .where(eq(mnos.id, parseInt(id)))

    return mno
  } catch (error) {
    console.error('Error fetching MNO:', error)
    return null
  }
}

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

export default async function MNODetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const [mno, connections] = await Promise.all([
    getMNO(id),
    getMNOConnections(parseInt(id))
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
        title={mno.name}
        description={`+${mno.countryCode} ${mno.networkCode || ''}`}
        actions={
          <Link href={`/admin/mnos/${id}/connections/new`}>
            <Button className="bg-white text-black hover:bg-zinc-200">
              <Plus className="h-4 w-4" />
              Add Connection
            </Button>
          </Link>
        }
      />

      <div className="grid gap-4 md:grid-cols-3">
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-blue-500/20">
              <Radio className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Connections</p>
              <p className="text-2xl font-bold text-white">{connections.length}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-emerald-500/20">
              <RefreshCw className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Active</p>
              <p className="text-2xl font-bold text-emerald-400">{connections.filter(c => c.status === 'active').length}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-amber-500/20">
              <Settings className="h-5 w-5 text-amber-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Inactive</p>
              <p className="text-2xl font-bold text-white">{connections.filter(c => c.status !== 'active').length}</p>
            </div>
          </div>
        </DataCard>
      </div>

      <DataCard title="Connections" subtitle="MNO connection configurations" action={{ label: 'Add Connection', href: `/admin/mnos/${id}/connections/new` }}>
        {connections.length > 0 ? (
          <div className="divide-y divide-white/10">
            {connections.map((conn) => (
              <div key={conn.id} className="p-4 flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="p-2 bg-zinc-800 rounded-lg">
                    <Settings className="h-5 w-5 text-zinc-400" />
                  </div>
                  <div>
                    <p className="font-medium text-white">{conn.protocol.toUpperCase()}</p>
                    <p className="text-sm text-zinc-400">{conn.host}:{conn.port}</p>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  {getStatusBadge(conn.status)}
                  <Link href={`/admin/mnos/${id}/connections/${conn.id}`} className="text-sm text-white hover:underline">
                    Edit
                  </Link>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="p-8 text-center text-zinc-500">
            No connections configured
          </div>
        )}
      </DataCard>
    </div>
  )
}
