import Link from 'next/link'
import {
  Wallet,
  Send,
  TrendingUp,
  TrendingDown,
  ArrowUpRight,
  MessageSquare,
  Clock,
  CheckCircle,
  AlertCircle
} from 'lucide-react'
import { db } from '@/lib/db'
import { REPORT_TIME_ZONE_SQL } from '@/lib/time-sql'
import { reportDayStart, reportMonthStart } from '@/lib/time'
import { getCurrentServiceProviderId, NOT_LINKED_MESSAGE } from '@/lib/service-provider'
import { wallets, messages } from '@/lib/schema'
import { and, desc, eq, gte, sql } from 'drizzle-orm'
import { formatMoney } from '@/lib/money'
import { PageHeader } from '@/components/dashboard/page-header'
import { StatCard } from '@/components/dashboard/stat-card'
import { DataCard } from '@/components/dashboard/data-card'
import { StatusBadge } from '@/components/dashboard/status-badge'

async function getSPStats(serviceProviderId: number) {

  const wallet = await db.query.wallets.findFirst({
    where: (wallets, { eq }) => eq(wallets.serviceProviderId, serviceProviderId),
  })

  const today = reportDayStart()
  const monthStart = reportMonthStart()

  const [messageStats] = await db
    .select({
      totalMessages: sql<number>`count(*)`,
      deliveredMessages: sql<number>`SUM(CASE WHEN ${messages.finalStatus} = 'delivered' THEN 1 ELSE 0 END)`,
    })
    .from(messages)
    .where(eq(messages.serviceProviderId, serviceProviderId))

  // Spend is what the wallet was charged this month, less refunds (same basis as Analytics).
  const monthSpend = wallet
    ? await db.execute(sql`
        SELECT COALESCE(SUM(CASE transaction_type WHEN 'debit' THEN amount
                                                  WHEN 'reversal' THEN -amount ELSE 0 END), 0) AS spent
        FROM wallet_transactions
        WHERE wallet_id = ${wallet.id} AND transaction_date >= ${monthStart}`)
    : null

  const todayMessages = await db
    .select({ count: sql<number>`count(*)` })
    .from(messages)
    .where(and(eq(messages.serviceProviderId, serviceProviderId), gte(messages.submittedAt, today)))

  const totalMessages = Number(messageStats.totalMessages) || 0
  const deliveredMessages = Number(messageStats.deliveredMessages) || 0
  const deliveryRate = totalMessages > 0 ? (deliveredMessages / totalMessages) * 100 : 0

  return {
    walletBalance: wallet ? Number(wallet.balance) : 0,
    walletCurrency: wallet?.currencyCode || 'NGN',
    walletType: wallet?.walletType || 'prepaid',
    messagesToday: Number(todayMessages[0]?.count) || 0,
    deliveryRate: Math.round(deliveryRate * 10) / 10,
    monthToDateSpent: Number(monthSpend?.rows[0]?.spent) || 0,
  }
}

async function get7DayStats(serviceProviderId: number) {
  // The last 7 Lagos days, today included.
  const since = reportDayStart(6)
  const lagosDay = sql`to_char(${messages.submittedAt} AT TIME ZONE ${REPORT_TIME_ZONE_SQL}, 'YYYY-MM-DD')`

  const dailyStats = await db
    .select({
      date: sql<string>`${lagosDay}`,
      processed: sql<number>`count(*)`,
      delivered: sql<number>`SUM(CASE WHEN ${messages.finalStatus} = 'delivered' THEN 1 ELSE 0 END)`,
      failed: sql<number>`SUM(CASE WHEN ${messages.finalStatus} = 'failed' THEN 1 ELSE 0 END)`,
    })
    .from(messages)
    .where(and(eq(messages.serviceProviderId, serviceProviderId), gte(messages.submittedAt, since)))
    .groupBy(lagosDay)
    .orderBy(lagosDay)

  return dailyStats.map(stat => ({
    date: stat.date,
    processed: Number(stat.processed),
    delivered: Number(stat.delivered),
    failed: Number(stat.failed),
  }))
}

async function getMessagesBySender(serviceProviderId: number) {
  const senderStats = await db
    .select({
      senderId: messages.originalSourceAddr,
      count: sql<number>`count(*)`,
    })
    .from(messages)
    .where(eq(messages.serviceProviderId, serviceProviderId))
    .groupBy(messages.originalSourceAddr)
    .orderBy(sql`count(*) DESC`)
    .limit(5)

  const total = senderStats.reduce((sum, s) => sum + Number(s.count), 0)

  return senderStats.map(stat => ({
    senderId: stat.senderId || 'Unknown',
    count: Number(stat.count),
    percentage: total > 0 ? Math.round((Number(stat.count) / total) * 100) : 0,
  }))
}

async function getRecentMessages(serviceProviderId: number) {
  const recentMessages = await db
    .select({
      id: messages.id,
      sender: messages.originalSourceAddr,
      recipient: messages.originalDestinationAddr,
      status: messages.finalStatus,
      cost: messages.cost,
      currency: messages.currencyCode,
      submittedAt: messages.submittedAt,
    })
    .from(messages)
    .where(eq(messages.serviceProviderId, serviceProviderId))
    .orderBy(desc(messages.submittedAt))
    .limit(10)

  return recentMessages.map(msg => ({
    id: Number(msg.id),
    sender: msg.sender || 'Unknown',
    recipient: msg.recipient ? `${msg.recipient.slice(0, 6)}***${msg.recipient.slice(-4)}` : 'Unknown',
    status: msg.status || 'pending',
    cost: msg.cost ? Number(msg.cost) : 0,
    currency: msg.currency,
    time: msg.submittedAt ? getRelativeTime(msg.submittedAt) : 'Unknown',
  }))
}

function getRelativeTime(date: Date): string {
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins} min ago`
  if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
  return date.toLocaleDateString()
}

function ChartLegend() {
  return (
    <div className="flex items-center justify-center gap-6 mt-4">
      <div className="flex items-center gap-2">
        <div className="w-3 h-3 rounded bg-emerald-500" />
        <span className="text-sm text-zinc-400">Delivered</span>
      </div>
      <div className="flex items-center gap-2">
        <div className="w-3 h-3 rounded bg-red-500" />
        <span className="text-sm text-zinc-400">Failed</span>
      </div>
    </div>
  )
}

export default async function SPDashboardPage() {
  const serviceProviderId = await getCurrentServiceProviderId()

  if (!serviceProviderId) {
    return (
      <div className="p-8 text-center">
        <p className="text-zinc-400">{NOT_LINKED_MESSAGE}</p>
      </div>
    )
  }

  const [stats, sevenDayStats, messagesBySender, recentMessages] = await Promise.all([
    getSPStats(serviceProviderId),
    get7DayStats(serviceProviderId),
    getMessagesBySender(serviceProviderId),
    getRecentMessages(serviceProviderId),
  ])

  const maxMessages = Math.max(...sevenDayStats.map(d => d.processed), 1)

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Dashboard"
        description="Welcome back! Here's your account overview."
        actions={
          <Link
            href="/sp/compose"
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
          >
            <Send className="h-4 w-4" />
            Send SMS
          </Link>
        }
      />

      <div className="grid gap-4 md:grid-cols-4">
        <StatCard
          title="Wallet Balance"
          value={formatMoney(stats.walletBalance, stats.walletCurrency)}
          icon={Wallet}
          trend={null}
        />
        <StatCard
          title="Messages Today"
          value={stats.messagesToday}
          icon={MessageSquare}
          trend={null}
        />
        <StatCard
          title="Delivery Rate"
          value={stats.deliveryRate}
          icon={TrendingUp}
          suffix="%"
          trend={null}
        />
        <StatCard
          title="Spent This Month"
          value={formatMoney(stats.monthToDateSpent, stats.walletCurrency)}
          icon={TrendingDown}
          trend={null}
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <DataCard
          title="Message Volume (Last 7 Days)"
          subtitle="Daily message delivery status"
          className="lg:col-span-2"
        >
          {sevenDayStats.length > 0 ? (
            <div className="h-[200px] flex items-end gap-2">
              {sevenDayStats.map((day, i) => (
                <div key={i} className="flex-1 flex flex-col items-center gap-2">
                  <div className="w-full flex flex-col gap-0.5" style={{ height: '160px' }}>
                    <div
                      className="w-full bg-emerald-500 rounded-t transition-all duration-300 hover:bg-emerald-400"
                      style={{ height: `${(day.delivered / maxMessages) * 100}%` }}
                    />
                    <div
                      className="w-full bg-red-500 transition-all duration-300 hover:bg-red-400"
                      style={{ height: `${(day.failed / maxMessages) * 100}%` }}
                    />
                  </div>
                  <span className="text-xs text-zinc-500">
                    {new Date(`${day.date}T00:00:00Z`).toLocaleDateString('en-US', { weekday: 'short', timeZone: 'UTC' })}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <div className="h-[200px] flex items-center justify-center text-zinc-500">
              No message data available
            </div>
          )}
          <ChartLegend />
        </DataCard>

        <DataCard
          title="Messages by Sender ID"
          subtitle="Top senders by volume"
        >
          {messagesBySender.length > 0 ? (
            <div className="space-y-4">
              {messagesBySender.map((sender, i) => (
                <div key={i}>
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-sm font-medium text-white">{sender.senderId}</span>
                    <span className="text-sm text-zinc-400">{sender.percentage}%</span>
                  </div>
                  <div className="h-2 bg-zinc-800 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-white rounded-full transition-all duration-300"
                      style={{ width: `${sender.percentage}%` }}
                    />
                  </div>
                  <p className="text-xs text-zinc-500 mt-1">{sender.count.toLocaleString()} messages</p>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-zinc-500 text-center py-8">No sender data available</p>
          )}
        </DataCard>
      </div>

      <DataCard
        title="Recent Messages"
        subtitle="Latest message activity"
        action={{
          label: 'View All',
          href: '/sp/messages',
        }}
      >
        {recentMessages.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-zinc-800/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Sender</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Recipient</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Cost</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Time</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/10">
                {recentMessages.map((msg) => (
                  <tr key={msg.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap font-medium text-white">{msg.sender}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-zinc-400">{msg.recipient}</td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <StatusBadge status={msg.status as 'delivered' | 'pending' | 'failed'} />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap font-medium text-white">
                      {msg.cost > 0 ? formatMoney(msg.cost, msg.currency) : '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-zinc-500">{msg.time}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="p-8 text-center text-zinc-500">No messages found</div>
        )}
      </DataCard>
    </div>
  )
}
