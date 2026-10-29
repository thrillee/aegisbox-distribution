import { Suspense } from 'react'
import { RefreshCw } from 'lucide-react'
import { db } from '@/lib/db'
import { reportDate, reportDayStart } from '@/lib/time'
import { serviceProviders, messages, mnoConnections, wallets, walletTransactions } from '@/lib/schema'
import { desc, gte, lt, eq, and, sql, count, sum } from 'drizzle-orm'
import { PageHeader } from '@/components/dashboard/page-header'
import { StatCard } from '@/components/dashboard/stat-card'
import { DataCard } from '@/components/dashboard/data-card'
import { Users, MessageSquare, TrendingUp, DollarSign } from 'lucide-react'
import Link from 'next/link'
import { formatMoney } from '@/lib/money'

async function getDashboardStats() {
  try {
    const [totalSPs] = await db.select({ count: count() })
      .from(serviceProviders)
      .where(eq(serviceProviders.status, 'active'))

    const today = reportDayStart()

    const [messagesToday] = await db.select({ count: count() })
      .from(messages)
      .where(gte(messages.receivedAt, today))

    const [totalMessages] = await db.select({ count: count() })
      .from(messages)

    const [deliveredMessages] = await db.select({ count: count() })
      .from(messages)
      .where(eq(messages.finalStatus, 'delivered'))

    // Revenue is what was charged for messages today (debits less refunds), per wallet currency.
    // Top-up credits are not revenue, and amounts in different currencies cannot be added up.
    const revenueRows = await db.select({
      currency: wallets.currencyCode,
      total: sql<string>`COALESCE(SUM(CASE ${walletTransactions.transactionType}
        WHEN 'debit' THEN ${walletTransactions.amount}
        WHEN 'reversal' THEN -${walletTransactions.amount}
        ELSE 0 END), 0)`,
    })
      .from(walletTransactions)
      .innerJoin(wallets, eq(wallets.id, walletTransactions.walletId))
      .where(gte(walletTransactions.transactionDate, today))
      .groupBy(wallets.currencyCode)

    return {
      totalSPs: totalSPs?.count || 0,
      messagesToday: messagesToday?.count || 0,
      deliveryRate: totalMessages?.count > 0
        ? Math.round((deliveredMessages?.count || 0) / totalMessages.count * 100 * 10) / 10
        : 0,
      revenueToday: revenueRows.length > 0
        ? revenueRows.map((r) => formatMoney(r.total, r.currency)).join(' · ')
        : formatMoney(0, 'NGN'),
      messagesTrend: 0,
      deliveryTrend: 0,
      revenueTrend: 0,
    }
  } catch (error) {
    console.error('Error fetching dashboard stats:', error)
    return {
      totalSPs: 0,
      messagesToday: 0,
      deliveryRate: 0,
      revenueToday: '-',
      messagesTrend: 0,
      deliveryTrend: 0,
      revenueTrend: 0,
    }
  }
}

async function get7DayAnalytics() {
  try {
    const days = []

    for (let i = 6; i >= 0; i--) {
      const date = reportDayStart(i)
      const nextDate = reportDayStart(i - 1)

      const [dayStats] = await db.select({
        processed: count(),
        delivered: sql<number>`sum(case when ${messages.finalStatus} = 'delivered' then 1 else 0 end)`,
        failed: sql<number>`sum(case when ${messages.finalStatus} = 'failed' then 1 else 0 end)`,
      })
        .from(messages)
        .where(and(gte(messages.receivedAt, date), lt(messages.receivedAt, nextDate)))

      const [revenue] = await db.select({
        total: sum(walletTransactions.amount),
      })
        .from(walletTransactions)
        .innerJoin(wallets, eq(walletTransactions.walletId, wallets.id))
        .where(and(gte(walletTransactions.transactionDate, date), lt(walletTransactions.transactionDate, nextDate)))

      days.push({
        date: date.toISOString(),
        dateOnly: reportDate(date),
        processed: parseInt(String(dayStats?.processed || '0')),
        delivered: parseInt(String(dayStats?.delivered || '0')),
        failed: parseInt(String(dayStats?.failed || '0')),
        revenue: parseFloat(String(revenue?.total || '0')),
      })
    }

    return days
  } catch (error) {
    console.error('Error fetching 7-day analytics:', error)
    const days = []
    for (let i = 6; i >= 0; i--) {
      const date = reportDayStart(i)
      days.push({
        date: date.toISOString(),
        dateOnly: reportDate(date),
        processed: 0,
        delivered: 0,
        failed: 0,
        revenue: 0,
      })
    }
    return days
  }
}

async function getTopSPs() {
  try {
    const topSPs = await db.select({
      id: serviceProviders.id,
      name: serviceProviders.name,
    })
      .from(serviceProviders)
      .orderBy(desc(serviceProviders.createdAt))
      .limit(5)

    const results = []
    for (const sp of topSPs) {
      const [messageCount] = await db.select({ count: count() })
        .from(messages)
        .where(eq(messages.serviceProviderId, sp.id))

      const [deliveredCount] = await db.select({ count: count() })
        .from(messages)
        .where(eq(messages.serviceProviderId, sp.id))

      const [revenue] = await db.select({
        total: sum(walletTransactions.amount),
      })
        .from(walletTransactions)
        .innerJoin(wallets, eq(walletTransactions.walletId, wallets.id))
        .where(eq(wallets.serviceProviderId, sp.id))

      results.push({
        id: sp.id,
        name: sp.name,
        messages: parseInt(String(messageCount?.count || '0')),
        deliveryRate: 0,
        revenue: parseFloat(String(revenue?.total || '0')),
      })
    }
    return results
  } catch (error) {
    console.error('Error fetching top SPs:', error)
    return []
  }
}

async function getMNOHealth() {
  try {
    const connections = await db.select({
      id: mnoConnections.id,
      mnoId: mnoConnections.mnoId,
      status: mnoConnections.status,
    })
      .from(mnoConnections)

    const connected = connections.filter(c => c.status === 'active').length
    const total = connections.length

    return {
      connected,
      total,
      connections: connections.map(c => ({
        id: c.id,
        status: c.status,
        lastActivity: 'N/A',
      })),
    }
  } catch (error) {
    console.error('Error fetching MNO health:', error)
    return {
      connected: 0,
      total: 0,
      connections: [],
    }
  }
}

function LoadingCard() {
  return (
    <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
      <div className="flex items-center justify-between">
        <div className="h-4 w-24 bg-zinc-800 rounded animate-pulse" />
      </div>
      <div className="mt-2">
        <div className="h-8 w-32 bg-zinc-800 rounded animate-pulse" />
      </div>
    </div>
  )
}

function ChartLegend() {
  return (
    <div className="flex items-center justify-center gap-6 mt-4">
      <div className="flex items-center gap-2">
        <div className="w-3 h-3 rounded bg-emerald-500" />
        <span className="text-sm text-zinc-400">Delivered</span>
      </div>
      <div className="flex items-center gap-2">
        <div className="w-3 h-3 rounded bg-amber-500" />
        <span className="text-sm text-zinc-400">Pending</span>
      </div>
      <div className="flex items-center gap-2">
        <div className="w-3 h-3 rounded bg-red-500" />
        <span className="text-sm text-zinc-400">Failed</span>
      </div>
    </div>
  )
}

export default async function AdminDashboardPage() {
  const [stats, analytics, topSPs, mnoHealth] = await Promise.all([
    getDashboardStats(),
    get7DayAnalytics(),
    getTopSPs(),
    getMNOHealth(),
  ])

  const maxMessages = Math.max(...analytics.map(d => d.processed), 1)

  return (
    <div className="p-8 space-y-8">
      <PageHeader
        title="Dashboard"
        description="Welcome back! Here's what's happening with your platform."
        actions={
          <button className="flex items-center gap-2 px-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-zinc-300 hover:bg-zinc-700 hover:text-white transition-colors">
            <RefreshCw className="h-4 w-4" />
            Refresh
          </button>
        }
      />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Service Providers"
          value={stats.totalSPs}
          icon={Users}
          trend="up"
          trendValue="+12%"
        />
        <StatCard
          title="Messages Today"
          value={stats.messagesToday}
          icon={MessageSquare}
          trend={stats.messagesTrend > 0 ? 'up' : 'down'}
          trendValue={`${stats.messagesTrend > 0 ? '+' : ''}${stats.messagesTrend}%`}
        />
        <StatCard
          title="Delivery Rate"
          value={stats.deliveryRate}
          icon={TrendingUp}
          suffix="%"
          trend={stats.deliveryTrend > 0 ? 'up' : 'down'}
          trendValue={`${stats.deliveryTrend > 0 ? '+' : ''}${stats.deliveryTrend}%`}
        />
        <StatCard
          title="Revenue Today"
          value={stats.revenueToday}
          icon={DollarSign}
          trend={stats.revenueTrend > 0 ? 'up' : 'down'}
          trendValue={`${stats.revenueTrend > 0 ? '+' : ''}${stats.revenueTrend}%`}
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-7">
        <DataCard
          title="Messages Overview (Last 7 Days)"
          subtitle="Daily message volume by status"
          className="col-span-4"
        >
          <div className="h-[300px] flex items-end gap-2">
            {analytics.map((day, i) => (
              <div key={i} className="flex-1 flex flex-col items-center gap-2">
                <div className="w-full flex flex-col gap-1" style={{ height: '260px' }}>
                  <div
                    className="w-full bg-emerald-500 rounded-t transition-all duration-300 hover:bg-emerald-400"
                    style={{ height: `${(day.delivered / maxMessages) * 100}%` }}
                    title={`Delivered: ${day.delivered.toLocaleString()}`}
                  />
                  <div
                    className="w-full bg-amber-500 transition-all duration-300 hover:bg-amber-400"
                    style={{ height: `${Math.max(0, ((day.processed - day.delivered - day.failed) / maxMessages) * 100)}%` }}
                  />
                  <div
                    className="w-full bg-red-500 transition-all duration-300 hover:bg-red-400"
                    style={{ height: `${(day.failed / maxMessages) * 100}%` }}
                  />
                </div>
                <span className="text-xs text-zinc-500">
                  {new Date(`${day.dateOnly}T00:00:00Z`).toLocaleDateString('en-US', { weekday: 'short', timeZone: 'UTC' })}
                </span>
              </div>
            ))}
          </div>
          <ChartLegend />
        </DataCard>

        <DataCard
          title="MNO Connection Status"
          subtitle="Network connectivity health"
          className="col-span-3"
        >
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-zinc-800/50 rounded-lg border border-white/10">
              <div>
                <p className="text-sm text-zinc-400">Connected</p>
                <p className="text-3xl font-bold text-emerald-500">{mnoHealth.connected}</p>
              </div>
              <div>
                <p className="text-sm text-zinc-400">Total</p>
                <p className="text-3xl font-bold text-white">{mnoHealth.total}</p>
              </div>
              <div className="text-right">
                <p className="text-sm text-zinc-400">Health</p>
                <p className="text-3xl font-bold text-white">
                  {mnoHealth.total > 0 ? Math.round((mnoHealth.connected / mnoHealth.total) * 100) : 0}%
                </p>
              </div>
            </div>
            <div className="space-y-2">
              {mnoHealth.connections.map((conn: any, i: number) => (
                <div key={i} className="flex items-center justify-between p-3 border border-white/10 rounded-lg">
                  <div className="flex items-center gap-3">
                    <div className={`w-2 h-2 rounded-full ${
                      conn.status === 'connected' ? 'bg-emerald-500' : 'bg-red-500'
                    }`} />
                    <span className="font-medium text-white">Connection #{conn.id}</span>
                  </div>
                  <span className="text-sm text-zinc-500">{conn.lastActivity}</span>
                </div>
              ))}
              {mnoHealth.connections.length === 0 && (
                <p className="text-sm text-zinc-500 text-center py-4">No MNO connections configured</p>
              )}
            </div>
          </div>
        </DataCard>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <DataCard
          title="Top Service Providers"
          subtitle="By message volume"
        >
          <div className="space-y-4">
            {topSPs.length > 0 ? topSPs.map((sp: any, i: number) => (
              <div key={sp.id} className="flex items-center justify-between p-3 border border-white/10 rounded-lg">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-white/10 flex items-center justify-center text-white font-semibold">
                    {i + 1}
                  </div>
                  <span className="font-medium text-white">{sp.name}</span>
                </div>
                <div className="text-right">
                  <p className="font-medium text-white">{sp.messages.toLocaleString()} msgs</p>
                  <p className="text-sm text-emerald-500">{sp.deliveryRate}% delivery</p>
                </div>
              </div>
            )) : (
              <p className="text-sm text-zinc-500 text-center py-4">No service providers found</p>
            )}
          </div>
        </DataCard>

        <DataCard
          title="Quick Actions"
          subtitle="Common administrative tasks"
        >
          <div className="grid grid-cols-2 gap-4">
            <Link
              href="/admin/service-providers"
              className="p-4 border border-white/10 rounded-lg hover:bg-white/5 transition-colors"
            >
              <Users className="h-8 w-8 text-emerald-500 mb-2" />
              <p className="font-medium text-white">Service Providers</p>
              <p className="text-sm text-zinc-400">Manage SP accounts</p>
            </Link>
            <Link
              href="/admin/mnos"
              className="p-4 border border-white/10 rounded-lg hover:bg-white/5 transition-colors"
            >
              <MessageSquare className="h-8 w-8 text-emerald-500 mb-2" />
              <p className="font-medium text-white">MNO Networks</p>
              <p className="text-sm text-zinc-400">Configure networks</p>
            </Link>
            <Link
              href="/admin/routing"
              className="p-4 border border-white/10 rounded-lg hover:bg-white/5 transition-colors"
            >
              <TrendingUp className="h-8 w-8 text-purple-500 mb-2" />
              <p className="font-medium text-white">Routing</p>
              <p className="text-sm text-zinc-400">Configure routes</p>
            </Link>
            <Link
              href="/admin/pricing"
              className="p-4 border border-white/10 rounded-lg hover:bg-white/5 transition-colors"
            >
              <DollarSign className="h-8 w-8 text-amber-500 mb-2" />
              <p className="font-medium text-white">Pricing</p>
              <p className="text-sm text-zinc-400">Set rates</p>
            </Link>
          </div>
        </DataCard>
      </div>
    </div>
  )
}
