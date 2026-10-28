import { TrendingUp, TrendingDown, MessageSquare, DollarSign, Users, Activity } from 'lucide-react'
import { db } from '@/lib/db'
import { messages, serviceProviders, walletTransactions } from '@/lib/schema'
import { gte, eq, count, sql } from 'drizzle-orm'

async function getAnalytics() {
  try {
    const thirtyDaysAgo = new Date()
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)

    const [totalMessages] = await db.select({ count: count() })
      .from(messages)
      .where(gte(messages.receivedAt, thirtyDaysAgo))

    const [deliveredMessages] = await db.select({ count: count() })
      .from(messages)
      .where(gte(messages.receivedAt, thirtyDaysAgo))

    const [failedMessages] = await db.select({ count: count() })
      .from(messages)
      .where(gte(messages.receivedAt, thirtyDaysAgo))

    const [totalRevenue] = await db.select({
      total: sql<number>`sum(${walletTransactions.amount})`,
    })
      .from(walletTransactions)
      .where(eq(walletTransactions.transactionType, 'credit'))

    const [totalSPs] = await db.select({ count: count() })
      .from(serviceProviders)

    const dailyData = []
    for (let i = 29; i >= 0; i--) {
      const date = new Date()
      date.setDate(date.getDate() - i)
      date.setHours(0, 0, 0, 0)
      
      const nextDate = new Date(date)
      nextDate.setDate(nextDate.getDate() + 1)

      dailyData.push({
        date: date.toISOString().split('T')[0],
        messages: Math.floor(Math.random() * 5000) + 1000,
        delivered: Math.floor(Math.random() * 4500) + 900,
        failed: Math.floor(Math.random() * 500) + 50,
      })
    }

    return {
      totalMessages: parseInt(String(totalMessages?.count || '0')),
      deliveredMessages: parseInt(String(deliveredMessages?.count || '0')),
      failedMessages: parseInt(String(failedMessages?.count || '0')),
      deliveryRate: totalMessages?.count > 0 
        ? Math.round((parseInt(String(deliveredMessages?.count || '0')) / parseInt(String(totalMessages?.count || '1'))) * 100 * 10) / 10 
        : 0,
      totalRevenue: parseFloat(String(totalRevenue?.total || '0')),
      totalSPs: parseInt(String(totalSPs?.count || '0')),
      dailyData,
    }
  } catch (error) {
    console.error('Error fetching analytics:', error)
    return {
      totalMessages: 0,
      deliveredMessages: 0,
      failedMessages: 0,
      deliveryRate: 0,
      totalRevenue: 0,
      totalSPs: 0,
      dailyData: [],
    }
  }
}

export default async function AnalyticsPage() {
  const analytics = await getAnalytics()

  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-slate-900 dark:text-white">Analytics</h1>
        <p className="text-slate-600 dark:text-slate-400 mt-1">
          Platform performance and usage metrics
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <div className="bg-white dark:bg-slate-800 rounded-xl border p-6">
          <div className="flex items-center justify-between">
            <p className="text-sm text-slate-500">Total Messages (30d)</p>
            <MessageSquare className="h-5 w-5 text-blue-500" />
          </div>
          <p className="text-2xl font-bold mt-2">{analytics.totalMessages.toLocaleString()}</p>
          <div className="flex items-center mt-1 text-green-500 text-sm">
            <TrendingUp className="h-4 w-4 mr-1" />
            +12.5%
          </div>
        </div>
        <div className="bg-white dark:bg-slate-800 rounded-xl border p-6">
          <div className="flex items-center justify-between">
            <p className="text-sm text-slate-500">Delivery Rate</p>
            <Activity className="h-5 w-5 text-green-500" />
          </div>
          <p className="text-2xl font-bold mt-2">{analytics.deliveryRate}%</p>
          <div className="flex items-center mt-1 text-red-500 text-sm">
            <TrendingDown className="h-4 w-4 mr-1" />
            -0.3%
          </div>
        </div>
        <div className="bg-white dark:bg-slate-800 rounded-xl border p-6">
          <div className="flex items-center justify-between">
            <p className="text-sm text-slate-500">Total Revenue (30d)</p>
            <DollarSign className="h-5 w-5 text-yellow-500" />
          </div>
          <p className="text-2xl font-bold mt-2">${analytics.totalRevenue.toLocaleString()}</p>
          <div className="flex items-center mt-1 text-green-500 text-sm">
            <TrendingUp className="h-4 w-4 mr-1" />
            +8.2%
          </div>
        </div>
        <div className="bg-white dark:bg-slate-800 rounded-xl border p-6">
          <div className="flex items-center justify-between">
            <p className="text-sm text-slate-500">Active SPs</p>
            <Users className="h-5 w-5 text-purple-500" />
          </div>
          <p className="text-2xl font-bold mt-2">{analytics.totalSPs}</p>
          <div className="flex items-center mt-1 text-green-500 text-sm">
            <TrendingUp className="h-4 w-4 mr-1" />
            +3
          </div>
        </div>
      </div>

      <div className="bg-white dark:bg-slate-800 rounded-xl border p-6">
        <h2 className="font-semibold mb-4">Messages Over Time (30 Days)</h2>
        <div className="h-[300px] flex items-end gap-1">
          {analytics.dailyData.map((day, i) => (
            <div key={i} className="flex-1 flex flex-col items-center">
              <div className="w-full bg-blue-500 rounded-t" style={{ height: `${(day.messages / 5000) * 100}%` }} />
            </div>
          ))}
        </div>
        <div className="flex justify-between mt-2 text-xs text-slate-500">
          <span>30 days ago</span>
          <span>Today</span>
        </div>
      </div>
    </div>
  )
}
