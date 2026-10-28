'use client'

import { useState, useEffect } from 'react'
import { TrendingUp, TrendingDown, MessageSquare, Users, Phone, Clock, Calendar } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

const mockAnalytics = {
  summary: {
    totalMessages: 125000,
    deliveredMessages: 118750,
    failedMessages: 3125,
    pendingMessages: 3125,
    deliveryRate: 95.0,
    avgCostPerMessage: 2.5,
    totalSpent: 312500,
  },
  dailyStats: [
    { date: '2024-06-01', sent: 15000, delivered: 14250, failed: 375, cost: 37500 },
    { date: '2024-06-02', sent: 18000, delivered: 17100, failed: 450, cost: 45000 },
    { date: '2024-06-03', sent: 12000, delivered: 11400, failed: 300, cost: 30000 },
    { date: '2024-06-04', sent: 20000, delivered: 19000, failed: 500, cost: 50000 },
    { date: '2024-06-05', sent: 16500, delivered: 15675, failed: 412, cost: 41250 },
    { date: '2024-06-06', sent: 22000, delivered: 20900, failed: 550, cost: 55000 },
    { date: '2024-06-07', sent: 20500, delivered: 19475, failed: 512, cost: 51250 },
  ],
  topSenders: [
    { senderId: 'ALERT', messages: 45000, percentage: 36.0 },
    { senderId: 'PROMO', messages: 35000, percentage: 28.0 },
    { senderId: 'VERIFY', messages: 25000, percentage: 20.0 },
    { senderId: 'NOTIFY', messages: 12500, percentage: 10.0 },
    { senderId: 'INFO', messages: 7500, percentage: 6.0 },
  ],
  messagesByMNO: [
    { mno: 'MTN', messages: 62500, percentage: 50.0, delivered: 59375 },
    { mno: 'Airtel', messages: 31250, percentage: 25.0, delivered: 29687 },
    { mno: 'Glo', messages: 18750, percentage: 15.0, delivered: 17812 },
    { mno: '9mobile', messages: 12500, percentage: 10.0, delivered: 11875 },
  ],
}

function formatCurrency(amount: number, currency: string = 'NGN') {
  return currency === 'NGN' ? `₦${amount.toLocaleString()}` : `$${amount.toLocaleString()}`
}

function calculatePercentageChange(current: number, previous: number) {
  if (previous === 0) return 100
  return ((current - previous) / previous) * 100
}

export default function AnalyticsPage() {
  const [period, setPeriod] = useState<'7d' | '30d' | '90d'>('7d')
  const [analytics] = useState(mockAnalytics)
  const [initialLoad, setInitialLoad] = useState(true)

  useEffect(() => {
    // Simulate initial data load
    const timer = setTimeout(() => {
      setInitialLoad(false)
    }, 100)
    return () => clearTimeout(timer)
  }, [])

  const maxDailyMessages = Math.max(...analytics.dailyStats.map(d => d.sent))

  if (initialLoad) {
    return <PageLoadingSkeleton />
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Analytics</h1>
          <p className="text-zinc-400 mt-1">
            Monitor your messaging performance and statistics
          </p>
        </div>
        <div className="flex items-center gap-2 bg-zinc-900/50 border border-white/10 rounded-xl p-1">
          <button
            onClick={() => setPeriod('7d')}
            className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
              period === '7d' ? 'bg-white text-black' : 'text-zinc-400 hover:text-white'
            }`}
          >
            7 Days
          </button>
          <button
            onClick={() => setPeriod('30d')}
            className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
              period === '30d' ? 'bg-white text-black' : 'text-zinc-400 hover:text-white'
            }`}
          >
            30 Days
          </button>
          <button
            onClick={() => setPeriod('90d')}
            className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
              period === '90d' ? 'bg-white text-black' : 'text-zinc-400 hover:text-white'
            }`}
          >
            90 Days
          </button>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-4">
          <div className="flex flex-row items-center justify-between pb-2">
            <p className="text-sm font-medium text-zinc-400">
              Total Messages
            </p>
            <MessageSquare className="h-4 w-4 text-zinc-500" />
          </div>
          <div className="text-2xl font-bold text-white">
            {analytics.summary.totalMessages.toLocaleString()}
          </div>
          <div className="flex items-center gap-1 mt-1">
            <TrendingUp className="h-3 w-3 text-emerald-400" />
            <span className="text-xs text-emerald-400 font-medium">+12.5%</span>
            <span className="text-xs text-zinc-500">vs last period</span>
          </div>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-4">
          <div className="flex flex-row items-center justify-between pb-2">
            <p className="text-sm font-medium text-zinc-400">
              Delivery Rate
            </p>
            <TrendingUp className="h-4 w-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-white">
            {analytics.summary.deliveryRate}%
          </div>
          <div className="flex items-center gap-1 mt-1">
            <TrendingUp className="h-3 w-3 text-emerald-400" />
            <span className="text-xs text-emerald-400 font-medium">+2.3%</span>
            <span className="text-xs text-zinc-500">vs last period</span>
          </div>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-4">
          <div className="flex flex-row items-center justify-between pb-2">
            <p className="text-sm font-medium text-zinc-400">
              Total Spent
            </p>
            <TrendingDown className="h-4 w-4 text-red-400" />
          </div>
          <div className="text-2xl font-bold text-white">
            {formatCurrency(analytics.summary.totalSpent)}
          </div>
          <div className="flex items-center gap-1 mt-1">
            <TrendingUp className="h-3 w-3 text-emerald-400" />
            <span className="text-xs text-emerald-400 font-medium">+8.2%</span>
            <span className="text-xs text-zinc-500">vs last period</span>
          </div>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-4">
          <div className="flex flex-row items-center justify-between pb-2">
            <p className="text-sm font-medium text-zinc-400">
              Failed Messages
            </p>
            <Users className="h-4 w-4 text-red-400" />
          </div>
          <div className="text-2xl font-bold text-white">
            {analytics.summary.failedMessages.toLocaleString()}
          </div>
          <div className="flex items-center gap-1 mt-1">
            <TrendingDown className="h-3 w-3 text-emerald-400" />
            <span className="text-xs text-emerald-400 font-medium">-15.3%</span>
            <span className="text-xs text-zinc-500">vs last period</span>
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
          <div className="p-6 border-b border-white/10">
            <p className="text-lg font-semibold text-white">
              Messages Over Time
            </p>
          </div>
          <div className="p-6">
            <div className="h-64 flex items-end gap-2">
              {analytics.dailyStats.map((day, index) => (
                <div key={index} className="flex-1 flex flex-col items-center gap-2">
                  <div className="w-full flex flex-col gap-0.5">
                    <div
                      className="w-full bg-emerald-500 rounded-t"
                      style={{ height: `${(day.sent / maxDailyMessages) * 160}px` }}
                    />
                    <div
                      className="w-full bg-red-500 rounded-b"
                      style={{ height: `${(day.failed / maxDailyMessages) * 160}px` }}
                    />
                  </div>
                  <span className="text-xs text-zinc-500">
                    {new Date(day.date).toLocaleDateString('en-US', { weekday: 'short' })}
                  </span>
                </div>
              ))}
            </div>
            <div className="flex items-center justify-center gap-6 mt-4">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 bg-emerald-500 rounded" />
                <span className="text-sm text-zinc-400">Delivered</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 bg-red-500 rounded" />
                <span className="text-sm text-zinc-400">Failed</span>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
          <div className="p-6 border-b border-white/10">
            <p className="text-lg font-semibold text-white">
              Messages by MNO
            </p>
          </div>
          <div className="p-6">
            <div className="space-y-4">
              {analytics.messagesByMNO.map((mno) => (
                <div key={mno.mno} className="space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Phone className="h-4 w-4 text-zinc-500" />
                      <span className="font-medium text-white">{mno.mno}</span>
                    </div>
                    <span className="text-sm text-zinc-400">
                      {mno.messages.toLocaleString()} ({mno.percentage}%)
                    </span>
                  </div>
                  <div className="w-full bg-zinc-800 rounded-full h-2">
                    <div
                      className="bg-white h-2 rounded-full"
                      style={{ width: `${mno.percentage}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
          <div className="p-6 border-b border-white/10">
            <p className="text-lg font-semibold text-white">
              Top Sender IDs
            </p>
          </div>
          <div className="p-6">
            <div className="space-y-4">
              {analytics.topSenders.map((sender, index) => (
                <div key={sender.senderId} className="flex items-center justify-between p-3 bg-black/50 rounded-xl">
                  <div className="flex items-center gap-3">
                    <span className="flex items-center justify-center w-6 h-6 bg-white/10 text-white rounded-full text-xs font-bold">
                      {index + 1}
                    </span>
                    <span className="font-mono font-medium text-white">{sender.senderId}</span>
                  </div>
                  <div className="text-right">
                    <div className="font-medium text-white">
                      {sender.messages.toLocaleString()} messages
                    </div>
                    <div className="text-xs text-zinc-500">{sender.percentage}% of total</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
          <div className="p-6 border-b border-white/10">
            <p className="text-lg font-semibold text-white">
              Daily Cost Breakdown
            </p>
          </div>
          <div className="p-6">
            <div className="space-y-3">
              {analytics.dailyStats.map((day) => (
                <div key={day.date} className="flex items-center justify-between p-3 bg-black/50 rounded-xl">
                  <div className="flex items-center gap-3">
                    <Calendar className="h-4 w-4 text-zinc-500" />
                    <span className="text-sm text-white">
                      {new Date(day.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                    </span>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-sm text-zinc-500">
                      {day.sent.toLocaleString()} msgs
                    </span>
                    <span className="font-semibold text-white">
                      {formatCurrency(day.cost)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
