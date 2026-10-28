import Link from 'next/link'
import { Activity, TrendingUp, AlertTriangle, XCircle, Users, Ban } from 'lucide-react'
import { db } from '@/lib/db'
import { routeHealth, senderIdHealth, routeDeliveryAttempts, mnos, routingAssignments, routingGroups, msisdnPrefixGroups } from '@/lib/schema'
import { sql, desc, gte, eq, and } from 'drizzle-orm'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { HealthScoreBadge } from '@/components/routing/health-score-badge'
import { RouteStatusBadge } from '@/components/routing/route-status-badge'
import { SenderStatusBadge } from '@/components/routing/sender-status-badge'

async function getHealthStats() {
  try {
    // Get route health statistics
    const routeStats = await db
      .select({
        total: sql<number>`count(*)`.mapWith(Number),
        // health_score is a percentage (0-100)
        healthy: sql<number>`count(*) filter (where ${routeHealth.healthScore}::numeric > 70)`.mapWith(Number),
        warning: sql<number>`count(*) filter (where ${routeHealth.healthScore}::numeric between 50 and 70)`.mapWith(Number),
        critical: sql<number>`count(*) filter (where ${routeHealth.healthScore}::numeric < 50)`.mapWith(Number),
        unavailable: sql<number>`count(*) filter (where ${routeHealth.isAvailable} = false)`.mapWith(Number),
      })
      .from(routeHealth)

    // Get sender health statistics
    const senderStats = await db
      .select({
        total: sql<number>`count(*)`.mapWith(Number),
        blocked: sql<number>`count(*) filter (where ${senderIdHealth.isBlocked} = true)`.mapWith(Number),
        avgDeliveryRate: sql<number>`coalesce(avg(${senderIdHealth.deliveryRate}::numeric), 0)`.mapWith(Number),
      })
      .from(senderIdHealth)

    return {
      routes: routeStats[0] || { total: 0, healthy: 0, warning: 0, critical: 0, unavailable: 0 },
      senders: senderStats[0] || { total: 0, blocked: 0, avgDeliveryRate: 0 },
    }
  } catch (error) {
    console.error('Error fetching health stats:', error)
    return {
      routes: { total: 0, healthy: 0, warning: 0, critical: 0, unavailable: 0 },
      senders: { total: 0, blocked: 0, avgDeliveryRate: 0 },
    }
  }
}

async function getCriticalRoutes() {
  try {
    const criticalRoutes = await db
      .select({
        assignmentId: routeHealth.routingAssignmentId,
        healthScore: routeHealth.healthScore,
        consecutiveFailures: routeHealth.consecutiveFailures,
        isAvailable: routeHealth.isAvailable,
        routingGroupName: routingGroups.name,
        prefixGroupName: msisdnPrefixGroups.name,
        mnoName: mnos.name,
      })
      .from(routeHealth)
      .leftJoin(routingAssignments, eq(routeHealth.routingAssignmentId, routingAssignments.id))
      .leftJoin(routingGroups, eq(routingAssignments.routingGroupId, routingGroups.id))
      .leftJoin(msisdnPrefixGroups, eq(routingAssignments.msisdnPrefixGroupId, msisdnPrefixGroups.id))
      .leftJoin(mnos, eq(routingAssignments.mnoId, mnos.id))
      .where(
        gte(routeHealth.consecutiveFailures, 3)
      )
      .orderBy(desc(routeHealth.consecutiveFailures))
      .limit(10)

    return criticalRoutes
  } catch (error) {
    console.error('Error fetching critical routes:', error)
    return []
  }
}

async function getBlockedSenders() {
  try {
    const blockedSenders = await db
      .select({
        id: senderIdHealth.id,
        senderId: senderIdHealth.senderId,
        mnoName: mnos.name,
        healthScore: senderIdHealth.healthScore,
        deliveryRate: senderIdHealth.deliveryRate,
        blockedAt: senderIdHealth.blockedAt,
        blockedReason: senderIdHealth.blockedReason,
      })
      .from(senderIdHealth)
      .leftJoin(mnos, eq(senderIdHealth.mnoId, mnos.id))
      .where(eq(senderIdHealth.isBlocked, true))
      .orderBy(desc(senderIdHealth.blockedAt))
      .limit(10)

    return blockedSenders
  } catch (error) {
    console.error('Error fetching blocked senders:', error)
    return []
  }
}

async function getRecentHealthEvents() {
  try {
    const recentEvents = await db
      .select({
        id: routeDeliveryAttempts.id,
        assignmentId: routeDeliveryAttempts.routingAssignmentId,
        success: routeDeliveryAttempts.success,
        attemptedAt: routeDeliveryAttempts.attemptedAt,
        routingGroupName: routingGroups.name,
        mnoName: mnos.name,
      })
      .from(routeDeliveryAttempts)
      .leftJoin(routingAssignments, eq(routeDeliveryAttempts.routingAssignmentId, routingAssignments.id))
      .leftJoin(routingGroups, eq(routingAssignments.routingGroupId, routingGroups.id))
      .leftJoin(mnos, eq(routingAssignments.mnoId, mnos.id))
      .orderBy(desc(routeDeliveryAttempts.attemptedAt))
      .limit(30)

    return recentEvents
  } catch (error) {
    console.error('Error fetching recent events:', error)
    return []
  }
}

export default async function HealthMonitoringPage() {
  const [stats, criticalRoutes, blockedSenders, recentEvents] = await Promise.all([
    getHealthStats(),
    getCriticalRoutes(),
    getBlockedSenders(),
    getRecentHealthEvents(),
  ])

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Health Monitoring"
        description="Monitor route and sender ID health metrics across your routing infrastructure"
        actions={
          <div className="flex items-center gap-2">
            <Link
              href="/admin/routing/health/routes"
              className="flex items-center gap-2 px-4 py-2 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
            >
              <TrendingUp className="h-4 w-4" />
              Route Health
            </Link>
            <Link
              href="/admin/routing/health/senders"
              className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
            >
              <Users className="h-4 w-4" />
              Sender Health
            </Link>
          </div>
        }
      />

      {/* Statistics Cards */}
      <div className="grid gap-4 md:grid-cols-3 lg:grid-cols-6">
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-blue-500/20">
              <Activity className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Total Routes</p>
              <p className="text-2xl font-bold text-white">{stats.routes.total}</p>
            </div>
          </div>
        </DataCard>

        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-emerald-500/20">
              <TrendingUp className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Healthy</p>
              <p className="text-2xl font-bold text-white">{stats.routes.healthy}</p>
            </div>
          </div>
        </DataCard>

        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-amber-500/20">
              <AlertTriangle className="h-5 w-5 text-amber-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Warning</p>
              <p className="text-2xl font-bold text-white">{stats.routes.warning}</p>
            </div>
          </div>
        </DataCard>

        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-red-500/20">
              <XCircle className="h-5 w-5 text-red-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Critical</p>
              <p className="text-2xl font-bold text-white">{stats.routes.critical}</p>
            </div>
          </div>
        </DataCard>

        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-purple-500/20">
              <Users className="h-5 w-5 text-purple-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Total Senders</p>
              <p className="text-2xl font-bold text-white">{stats.senders.total}</p>
            </div>
          </div>
        </DataCard>

        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-red-500/20">
              <Ban className="h-5 w-5 text-red-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Blocked</p>
              <p className="text-2xl font-bold text-white">{stats.senders.blocked}</p>
            </div>
          </div>
        </DataCard>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Critical Routes Alert */}
        <DataCard title="Critical Routes" subtitle={`${criticalRoutes.length} routes need attention`}>
          {criticalRoutes.length === 0 ? (
            <div className="py-8 text-center text-zinc-500">
              <TrendingUp className="h-12 w-12 mx-auto mb-2 opacity-50" />
              <p>All routes are healthy</p>
            </div>
          ) : (
            <div className="space-y-3">
              {criticalRoutes.map((route) => (
                <div
                  key={route.assignmentId}
                  className="flex items-center justify-between p-3 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <p className="text-sm font-medium text-white">
                        {route.routingGroupName || 'Unknown'} → {route.mnoName || 'Unknown'}
                      </p>
                      <RouteStatusBadge
                        isAvailable={route.isAvailable}
                        consecutiveFailures={route.consecutiveFailures}
                      />
                    </div>
                    <p className="text-xs text-zinc-400">
                      Prefix: {route.prefixGroupName || 'Unknown'} • {route.consecutiveFailures} consecutive failures
                    </p>
                  </div>
                  <HealthScoreBadge score={route.healthScore} />
                </div>
              ))}
              <Link
                href="/admin/routing/health/routes"
                className="block text-center text-sm text-blue-400 hover:text-blue-300 transition-colors mt-4"
              >
                View all routes →
              </Link>
            </div>
          )}
        </DataCard>

        {/* Blocked Senders */}
        <DataCard title="Blocked Senders" subtitle={`${blockedSenders.length} senders blocked`}>
          {blockedSenders.length === 0 ? (
            <div className="py-8 text-center text-zinc-500">
              <Users className="h-12 w-12 mx-auto mb-2 opacity-50" />
              <p>No blocked senders</p>
            </div>
          ) : (
            <div className="space-y-3">
              {blockedSenders.map((sender) => (
                <div
                  key={sender.id}
                  className="flex items-center justify-between p-3 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <p className="text-sm font-medium text-white">{sender.senderId}</p>
                      <SenderStatusBadge isBlocked={true} healthScore={sender.healthScore} />
                    </div>
                    <p className="text-xs text-zinc-400">
                      MNO: {sender.mnoName || 'Unknown'} • Blocked: {sender.blockedAt ? new Date(sender.blockedAt).toLocaleString() : 'Unknown'}
                    </p>
                    {sender.blockedReason && (
                      <p className="text-xs text-zinc-500 mt-1">{sender.blockedReason}</p>
                    )}
                  </div>
                  <HealthScoreBadge score={sender.healthScore} />
                </div>
              ))}
              <Link
                href="/admin/routing/health/senders"
                className="block text-center text-sm text-blue-400 hover:text-blue-300 transition-colors mt-4"
              >
                View all senders →
              </Link>
            </div>
          )}
        </DataCard>
      </div>

      {/* Recent Health Events */}
      <DataCard title="Recent Delivery Attempts" subtitle="Last 30 route delivery attempts">
        {recentEvents.length === 0 ? (
          <div className="py-8 text-center text-zinc-500">
            <Activity className="h-12 w-12 mx-auto mb-2 opacity-50" />
            <p>No recent delivery attempts</p>
          </div>
        ) : (
          <div className="space-y-2">
            {recentEvents.map((event) => (
              <div
                key={event.id}
                className="flex items-center gap-3 p-2 rounded-lg hover:bg-white/5 transition-colors"
              >
                <div
                  className={`w-2 h-2 rounded-full flex-shrink-0 ${
                    event.success ? 'bg-emerald-400' : 'bg-red-400'
                  }`}
                />
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-white truncate">
                    {event.routingGroupName || 'Unknown'} → {event.mnoName || 'Unknown'}
                  </p>
                </div>
                <p className="text-xs text-zinc-500 whitespace-nowrap">
                  {event.attemptedAt ? new Date(event.attemptedAt).toLocaleTimeString() : 'Unknown'}
                </p>
                <span
                  className={`px-2 py-0.5 rounded text-xs font-medium ${
                    event.success
                      ? 'bg-emerald-500/20 text-emerald-400'
                      : 'bg-red-500/20 text-red-400'
                  }`}
                >
                  {event.success ? 'Success' : 'Failed'}
                </span>
              </div>
            ))}
          </div>
        )}
      </DataCard>

      {/* How It Works */}
      <DataCard title="About Health Monitoring" subtitle="Understanding the intelligent routing system">
        <div className="grid gap-6 md:grid-cols-2">
          <div>
            <h3 className="text-sm font-semibold text-white mb-3">Route Health Tracking</h3>
            <ul className="space-y-2 text-sm text-zinc-400">
              <li className="flex items-start gap-2">
                <span className="text-emerald-400 mt-0.5">•</span>
                <span>Tracks success/failure rates for each routing assignment</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-emerald-400 mt-0.5">•</span>
                <span>Automatically marks routes unavailable after 5 consecutive failures</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-emerald-400 mt-0.5">•</span>
                <span>Health scores calculated from last 1000 delivery attempts</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-emerald-400 mt-0.5">•</span>
                <span>Routes with health score below 70% trigger failover</span>
              </li>
            </ul>
          </div>
          <div>
            <h3 className="text-sm font-semibold text-white mb-3">Sender ID Health</h3>
            <ul className="space-y-2 text-sm text-zinc-400">
              <li className="flex items-start gap-2">
                <span className="text-blue-400 mt-0.5">•</span>
                <span>Monitors delivery rates and DLR response rates per MNO</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-blue-400 mt-0.5">•</span>
                <span>Automatically blocks senders with delivery rate below 50%</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-blue-400 mt-0.5">•</span>
                <span>Auto-unblock when delivery rate exceeds 70%</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-blue-400 mt-0.5">•</span>
                <span>Manual unblock available for false positives</span>
              </li>
            </ul>
          </div>
        </div>
      </DataCard>
    </div>
  )
}
