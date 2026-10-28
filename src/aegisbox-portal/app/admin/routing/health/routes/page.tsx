import { db } from '@/lib/db'
import { routeHealth, routingAssignments, routingGroups, msisdnPrefixGroups, mnos } from '@/lib/schema'
import { eq, desc, isNull } from 'drizzle-orm'
import { RouteHealthTable } from './route-health-table'
import { PageHeader } from '@/components/dashboard/page-header'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'

async function getRouteHealthData() {
  try {
    const routes = await db
      .select({
        assignmentId: routeHealth.routingAssignmentId,
        healthScore: routeHealth.healthScore,
        successCount: routeHealth.successCount,
        failCount: routeHealth.failCount,
        totalAttempts: routeHealth.totalAttempts,
        consecutiveFailures: routeHealth.consecutiveFailures,
        isAvailable: routeHealth.isAvailable,
        lastAttemptAt: routeHealth.lastAttemptAt,
        updatedAt: routeHealth.updatedAt,
        routingGroupId: routingAssignments.routingGroupId,
        routingGroupName: routingGroups.name,
        prefixGroupName: msisdnPrefixGroups.name,
        mnoName: mnos.name,
        priority: routingAssignments.priority,
        weight: routingAssignments.weight,
      })
      .from(routeHealth)
      .leftJoin(routingAssignments, eq(routeHealth.routingAssignmentId, routingAssignments.id))
      .leftJoin(routingGroups, eq(routingAssignments.routingGroupId, routingGroups.id))
      .leftJoin(msisdnPrefixGroups, eq(routingAssignments.msisdnPrefixGroupId, msisdnPrefixGroups.id))
      .leftJoin(mnos, eq(routingAssignments.mnoId, mnos.id))
      // Route-wide rows only: the ones the router checks and "mark available" changes
      .where(isNull(routeHealth.senderId))
      .orderBy(desc(routeHealth.lastAttemptAt))

    return routes
  } catch (error) {
    console.error('Error fetching route health data:', error)
    return []
  }
}

async function getRoutingGroupsForFilter() {
  try {
    const groups = await db
      .select({
        id: routingGroups.id,
        name: routingGroups.name,
      })
      .from(routingGroups)
      .orderBy(routingGroups.name)

    return groups
  } catch (error) {
    console.error('Error fetching routing groups:', error)
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

export default async function RouteHealthPage() {
  const [routes, routingGroupsList, mnosList] = await Promise.all([
    getRouteHealthData(),
    getRoutingGroupsForFilter(),
    getMNOsForFilter(),
  ])

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Route Health"
        description="Monitor health metrics for all routing assignments"
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

      <RouteHealthTable
        initialData={routes}
        routingGroups={routingGroupsList}
        mnos={mnosList}
      />
    </div>
  )
}
