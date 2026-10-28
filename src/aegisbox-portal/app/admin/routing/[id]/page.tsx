import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { db } from '@/lib/db'
import { routingGroups, routingAssignments, msisdnPrefixGroups, mnos, msisdnPrefixEntries } from '@/lib/schema'
import { eq, desc, sql } from 'drizzle-orm'
import { RoutingGroupPage } from './routing-group-page'

async function getRoutingGroup(id: number) {
  const groups = await db.select({
    id: routingGroups.id,
    name: routingGroups.name,
    reference: routingGroups.reference,
    description: routingGroups.description,
    createdAt: routingGroups.createdAt,
    updatedAt: routingGroups.updatedAt,
  })
    .from(routingGroups)
    .where(eq(routingGroups.id, id))

  const group = groups[0]
  if (!group) return null
  
  const assignments = await db.select({
    id: routingAssignments.id,
    msisdnPrefixGroupId: routingAssignments.msisdnPrefixGroupId,
    mnoId: routingAssignments.mnoId,
    status: routingAssignments.status,
    priority: routingAssignments.priority,
    comment: routingAssignments.comment,
    createdAt: routingAssignments.createdAt,
    msisdnPrefixGroupName: msisdnPrefixGroups.name,
    mnoName: mnos.name,
    prefixes: sql<{ id: number; prefix: string }[]>`cast(json_agg(json_build_object('id', ${msisdnPrefixEntries.id}, 'prefix', ${msisdnPrefixEntries.msisdnPrefix})) as json)`.mapWith(Object),
  })
    .from(routingAssignments)
    .innerJoin(msisdnPrefixGroups, eq(routingAssignments.msisdnPrefixGroupId, msisdnPrefixGroups.id))
    .innerJoin(mnos, eq(routingAssignments.mnoId, mnos.id))
    .leftJoin(msisdnPrefixEntries, eq(msisdnPrefixEntries.msisdnPrefixGroupId, msisdnPrefixGroups.id))
    .where(eq(routingAssignments.routingGroupId, id))
    .groupBy(routingAssignments.id, msisdnPrefixGroups.id, mnos.id)
    .orderBy(routingAssignments.priority)

  return { ...group, status: 'active', assignments: assignments as any }
}

export default async function RoutingGroupPageServer({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const groupId = parseInt(id)
  
  if (isNaN(groupId)) {
    return (
      <div className="p-8">
        <Link href="/admin/routing" className="flex items-center gap-2 text-zinc-400 hover:text-white">
          <ArrowLeft className="h-4 w-4" />
          Back to Routing
        </Link>
        <div className="mt-8 text-center">
          <h1 className="text-2xl font-bold text-white">Invalid Routing Group ID</h1>
        </div>
      </div>
    )
  }

  const group = await getRoutingGroup(groupId)

  if (!group) {
    return (
      <div className="p-8">
        <Link href="/admin/routing" className="flex items-center gap-2 text-zinc-400 hover:text-white">
          <ArrowLeft className="h-4 w-4" />
          Back to Routing
        </Link>
        <div className="mt-8 text-center">
          <h1 className="text-2xl font-bold text-white">Routing Group Not Found</h1>
        </div>
      </div>
    )
  }

  return <RoutingGroupPage initialGroup={group} />
}
