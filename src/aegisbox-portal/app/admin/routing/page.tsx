import Link from 'next/link'
import { Plus, Settings, MapPin, TrendingUp, Star } from 'lucide-react'
import { db } from '@/lib/db'
import { routingGroups, msisdnPrefixGroups, mnos, routingAssignments } from '@/lib/schema'
import { eq, desc } from 'drizzle-orm'
import { RoutingGroupsList } from './routing-groups-list'
import { RoutingRulesList } from './routing-rules-list'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

async function getRoutingGroupsData() {
  try {
    const groups = await db.select({
      id: routingGroups.id,
      name: routingGroups.name,
      reference: routingGroups.reference,
      description: routingGroups.description,
      createdAt: routingGroups.createdAt,
    })
      .from(routingGroups)
      .orderBy(desc(routingGroups.createdAt))
    return groups
  } catch (error) {
    console.error('Error fetching routing groups:', error)
    return []
  }
}

async function getAssignmentsByGroup(groupId: number) {
  try {
    const assignments = await db
      .select({
        id: routingAssignments.id,
        routingGroupId: routingAssignments.routingGroupId,
        msisdnPrefixGroupId: routingAssignments.msisdnPrefixGroupId,
        mnoId: routingAssignments.mnoId,
        priority: routingAssignments.priority,
        status: routingAssignments.status,
        comment: routingAssignments.comment,
        createdAt: routingAssignments.createdAt,
        mnoName: mnos.name,
        mnoCountryCode: mnos.countryCode,
        prefixGroupName: msisdnPrefixGroups.name,
        prefixGroupReference: msisdnPrefixGroups.reference,
        groupName: routingGroups.name,
      })
      .from(routingAssignments)
      .leftJoin(mnos, eq(routingAssignments.mnoId, mnos.id))
      .leftJoin(msisdnPrefixGroups, eq(routingAssignments.msisdnPrefixGroupId, msisdnPrefixGroups.id))
      .leftJoin(routingGroups, eq(routingAssignments.routingGroupId, routingGroups.id))
      .where(eq(routingAssignments.routingGroupId, groupId))
      .orderBy(routingAssignments.priority)
    return assignments
  } catch (error) {
    console.error('Error fetching assignments:', error)
    return []
  }
}

async function getAssignmentsCount() {
  try {
    const { sql } = await import('drizzle-orm')
    const result = await db.select({
      routingGroupId: routingAssignments.routingGroupId,
      count: sql<number>`count(*)`.mapWith(Number),
    })
      .from(routingAssignments)
      .groupBy(routingAssignments.routingGroupId)
    return result
  } catch (error) {
    console.error('Error fetching assignments count:', error)
    return []
  }
}

async function getPrefixGroupsData() {
  try {
    const groups = await db.select({
      id: msisdnPrefixGroups.id,
      name: msisdnPrefixGroups.name,
      reference: msisdnPrefixGroups.reference,
    })
      .from(msisdnPrefixGroups)
      .orderBy(msisdnPrefixGroups.name)
    return groups
  } catch (error) {
    console.error('Error fetching prefix groups:', error)
    return []
  }
}

async function getMNOsData() {
  try {
    const data = await db.select({
      id: mnos.id,
      name: mnos.name,
      countryCode: mnos.countryCode,
    })
      .from(mnos)
      .orderBy(mnos.name)
    return data
  } catch (error) {
    console.error('Error fetching MNOs:', error)
    return []
  }
}

interface PageProps {
  searchParams: Promise<{ selected?: string }>
}

export default async function RoutingPage({ searchParams }: PageProps) {
  const resolvedSearchParams = await searchParams
  const selectedGroupId = resolvedSearchParams.selected ? parseInt(resolvedSearchParams.selected) : undefined

  const [routingGroupsData, prefixGroups, mnosData, assignmentsCountResult, selectedGroupAssignments] = await Promise.all([
    getRoutingGroupsData(),
    getPrefixGroupsData(),
    getMNOsData(),
    getAssignmentsCount(),
    selectedGroupId ? getAssignmentsByGroup(selectedGroupId) : Promise.resolve([]),
  ])

  const assignmentsMap = new Map(assignmentsCountResult.map((r: any) => [r.routingGroupId, r.count]))
  const groupsWithCounts = routingGroupsData.map((g: any) => ({
    ...g,
    assignmentCount: assignmentsMap.get(g.id) || 0,
  }))
  const activeRulesCount = groupsWithCounts.reduce((acc: number, g: any) => acc + (g.assignmentCount || 0), 0)

  const selectedGroup = groupsWithCounts.find((g: any) => g.id === selectedGroupId)

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Routing Configuration"
        description="Configure message routing rules between prefixes and MNOs"
        actions={
          <div className="flex items-center gap-2">
            <Link
              href="/admin/routing/prefix-groups"
              className="flex items-center gap-2 px-4 py-2 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
            >
              <MapPin className="h-4 w-4" />
              Prefix Groups
            </Link>
            <Link
              href="/admin/routing/new"
              className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
            >
              <Plus className="h-4 w-4" />
              Create Routing Group
            </Link>
          </div>
        }
      />

      <div className="grid gap-4 md:grid-cols-4">
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-blue-500/20">
              <Settings className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Routing Groups</p>
              <p className="text-2xl font-bold text-white">{routingGroupsData.length}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-emerald-500/20">
              <TrendingUp className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Active Rules</p>
              <p className="text-2xl font-bold text-white">{activeRulesCount}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-purple-500/20">
              <MapPin className="h-5 w-5 text-purple-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Prefix Groups</p>
              <p className="text-2xl font-bold text-white">{prefixGroups.length}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-amber-500/20">
              <Star className="h-5 w-5 text-amber-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">MNOs</p>
              <p className="text-2xl font-bold text-white">{mnosData.length}</p>
            </div>
          </div>
        </DataCard>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-1">
          <RoutingGroupsList groups={groupsWithCounts} selectedGroupId={selectedGroupId} />
        </div>

        <div className="lg:col-span-2">
          <DataCard title="" className="!p-0">
            <div className="p-4 border-b border-white/10 flex items-center justify-between">
              <h2 className="font-semibold text-white">
                {selectedGroup ? `${selectedGroup.name} - Routing Rules` : 'Select a Routing Group'}
              </h2>
              {selectedGroupId && (
                <Link href={`/admin/routing/${selectedGroupId}/assignments/new`} className="flex items-center gap-2 px-3 py-1.5 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors text-sm">
                  <Plus className="h-4 w-4" />
                  Add Rule
                </Link>
              )}
            </div>
            
            {!selectedGroup ? (
              <div className="p-8 text-center">
                <Settings className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
                <p className="text-zinc-500">
                  Select a routing group from the list to view its rules
                </p>
              </div>
            ) : (
              <RoutingRulesList 
                assignments={selectedGroupAssignments} 
                groupName={selectedGroup.name}
                groupId={selectedGroupId!}
              />
            )}
          </DataCard>

          <DataCard title="How Routing Works" subtitle="Understanding the routing process">
            <div className="space-y-4">
              <div className="flex items-start gap-4">
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-white/10 flex items-center justify-center text-white font-semibold">
                  1
                </div>
                <div>
                  <p className="font-medium text-white">Incoming Message</p>
                  <p className="text-sm text-zinc-400">
                    Message arrives with destination MSISDN (e.g., +2347031234567)
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-4">
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-white/10 flex items-center justify-center text-white font-semibold">
                  2
                </div>
                <div>
                  <p className="font-medium text-white">Prefix Matching</p>
                  <p className="text-sm text-zinc-400">
                    System extracts prefix (+234703) and matches against prefix groups
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-4">
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-white/10 flex items-center justify-center text-white font-semibold">
                  3
                </div>
                <div>
                  <p className="font-medium text-white">Route Selection</p>
                  <p className="text-sm text-zinc-400">
                    Routes to assigned MNO based on priority order
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-4">
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-white/10 flex items-center justify-center text-white font-semibold">
                  4
                </div>
                <div>
                  <p className="font-medium text-white">Failover</p>
                  <p className="text-sm text-zinc-400">
                    If primary fails, message routes to failover MNO (if configured)
                  </p>
                </div>
              </div>
            </div>
          </DataCard>
        </div>
      </div>
    </div>
  )
}
