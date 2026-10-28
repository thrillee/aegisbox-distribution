'use client'

import { useState, useEffect, useCallback } from 'react'
import { use } from 'react'
import { 
  TrendingUp, TrendingDown, MessageSquare, DollarSign, 
  CheckCircle, XCircle, Clock, BarChart3, Loader2,
  ArrowUp, ArrowDown, Activity, Globe, Calendar
} from 'lucide-react'
import { 
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  LineChart, Line, PieChart, Pie, Cell, AreaChart, Area
} from 'recharts'
import { DataCard } from '@/components/dashboard/data-card'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

interface AnalyticsData {
  overview: {
    totalMessages: number
    totalSegments: number
    totalCost: number
    delivered: number
    failed: number
    pending: number
    rejected: number
    deliveryRate: string
    failureRate: string
  }
  provider: {
    name: string
    email: string
    status: string
    currency: string
    createdAt: string
  } | null
  wallet: {
    balance: number
    totalCredits: number
    totalDebits: number
  }
  dailyStats: {
    date: string
    total: number
    delivered: number
    failed: number
    cost: number
    deliveryRate: string
  }[]
  topDestinations: {
    prefix: string
    count: number
    delivered: number
    failed: number
    deliveryRate: string
  }[]
  statusBreakdown: {
    status: string
    count: number
    percentage: number
  }[]
  hourlyDistribution: {
    hour: number
    count: number
  }[]
  costByDay: {
    date: string
    cost: number
  }[]
  period: {
    days: number
    startDate: string
    endDate: string
  }
}

const STATUS_COLORS: Record<string, string> = {
  delivered: '#10B981',
  failed: '#EF4444',
  pending: '#F59E0B',
  rejected: '#F97316',
  accepted: '#10B981',
  forwarded: '#10B981',
  undeliverable: '#EF4444',
  queued: '#F59E0B',
  processing: '#8B5CF6',
  submitted: '#6366F1',
  received: '#3B82F6',
}

export default function ServiceProviderAnalyticsPage({ params }: { params: Promise<{ id: string }> }) {
  const resolvedParams = use(params)
  const [loading, setLoading] = useState(true)
  const [initialLoad, setInitialLoad] = useState(true)
  const [analytics, setAnalytics] = useState<AnalyticsData | null>(null)
  const [period, setPeriod] = useState(30)

  const fetchAnalytics = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetch(`/api/admin/service-providers/${resolvedParams.id}/analytics?days=${period}`)
      if (res.ok) {
        const data = await res.json()
        setAnalytics(data)
      }
    } catch (error) {
      console.error('Error fetching analytics:', error)
    } finally {
      setLoading(false)
      setInitialLoad(false)
    }
  }, [resolvedParams.id, period])

  useEffect(() => {
    fetchAnalytics()
  }, [fetchAnalytics])

  const dailyChartData = analytics?.dailyStats.map(d => ({
    date: new Date(d.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    Total: d.total,
    Delivered: d.delivered,
    Failed: d.failed,
  })) || []

  const costChartData = analytics?.costByDay.map(c => ({
    date: new Date(c.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    Cost: c.cost,
  })) || []

  const hourlyData = analytics?.hourlyDistribution.map(h => ({
    hour: `${h.hour}:00`,
    messages: h.count,
  })) || []

  const pieData = analytics?.statusBreakdown.map(s => ({
    name: s.status.charAt(0).toUpperCase() + s.status.slice(1),
    value: s.count,
    color: STATUS_COLORS[s.status] || '#6B7280',
  })) || []

  const formatCurrency = (amount: number) => {
    if (!analytics?.provider) return `$${amount.toFixed(2)}`
    const symbol = analytics.provider.currency === 'NGN' ? '₦' : 
                   analytics.provider.currency === 'USD' ? '$' : 
                   analytics.provider.currency === 'EUR' ? '€' : 
                   analytics.provider.currency === 'GBP' ? '£' : ''
    return `${symbol}${amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
  }

  if (initialLoad) {
    return <PageLoadingSkeleton />
  }

  if (!analytics) {
    return (
      <div className="p-8 text-center">
        <BarChart3 className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
        <p className="text-zinc-500">Failed to load analytics</p>
      </div>
    )
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">
            {analytics.provider?.name || 'Service Provider'} Analytics
          </h1>
          <p className="text-zinc-400 mt-1">Performance metrics and usage statistics</p>
        </div>
        <div className="flex items-center gap-2">
          {[7, 14, 30, 90].map(d => (
            <button
              key={d}
              onClick={() => setPeriod(d)}
              className={`px-3 py-1.5 text-sm rounded-lg transition-colors ${
                period === d 
                  ? 'bg-white text-black' 
                  : 'bg-zinc-800 text-zinc-400 hover:text-white'
              }`}
            >
              {d}D
            </button>
          ))}
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-5">
        <DataCard title="Total Messages" subtitle="All messages sent">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-blue-500/20">
              <MessageSquare className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-2xl font-bold text-white">{analytics.overview.totalMessages.toLocaleString()}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="Delivery Rate" subtitle="Successful deliveries">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-emerald-500/20">
              <CheckCircle className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-2xl font-bold text-emerald-400">{analytics.overview.deliveryRate}%</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="Failure Rate" subtitle="Failed deliveries">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-red-500/20">
              <XCircle className="h-5 w-5 text-red-400" />
            </div>
            <div>
              <p className="text-2xl font-bold text-red-400">{analytics.overview.failureRate}%</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="Total Cost" subtitle="Revenue generated">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-amber-500/20">
              <DollarSign className="h-5 w-5 text-amber-400" />
            </div>
            <div>
              <p className="text-2xl font-bold text-white">{formatCurrency(analytics.overview.totalCost)}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="Wallet Balance" subtitle="Current balance">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-purple-500/20">
              <Wallet className="h-5 w-5 text-purple-400" />
            </div>
            <div>
              <p className="text-2xl font-bold text-white">{formatCurrency(analytics.wallet.balance)}</p>
            </div>
          </div>
        </DataCard>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <DataCard title="Daily Message Volume" subtitle="Messages over time" className="lg:col-span-2">
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={dailyChartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#3f3f46" />
              <XAxis dataKey="date" stroke="#a1a1aa" fontSize={12} />
              <YAxis stroke="#a1a1aa" fontSize={12} />
              <Tooltip 
                contentStyle={{ 
                  backgroundColor: '#18181b', 
                  border: '1px solid #3f3f46',
                  borderRadius: '8px',
                  color: '#fff'
                }}
              />
              <Area type="monotone" dataKey="Total" stroke="#6366f1" fill="#6366f1" fillOpacity={0.2} />
              <Area type="monotone" dataKey="Delivered" stroke="#10b981" fill="#10b981" fillOpacity={0.2} />
              <Area type="monotone" dataKey="Failed" stroke="#ef4444" fill="#ef4444" fillOpacity={0.2} />
            </AreaChart>
          </ResponsiveContainer>
        </DataCard>

        <DataCard title="Status Distribution" subtitle="Message status breakdown">
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={pieData}
                cx="50%"
                cy="50%"
                innerRadius={50}
                outerRadius={80}
                paddingAngle={2}
                dataKey="value"
              >
                {pieData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip 
                contentStyle={{ 
                  backgroundColor: '#18181b', 
                  border: '1px solid #3f3f46',
                  borderRadius: '8px',
                  color: '#fff'
                }}
              />
            </PieChart>
          </ResponsiveContainer>
          <div className="flex flex-wrap justify-center gap-3 mt-4">
            {pieData.slice(0, 5).map(item => (
              <div key={item.name} className="flex items-center gap-1.5">
                <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: item.color }} />
                <span className="text-xs text-zinc-400">{item.name}: {item.value.toLocaleString()}</span>
              </div>
            ))}
          </div>
        </DataCard>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <DataCard title="Daily Cost" subtitle="Cost over time">
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={costChartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#3f3f46" />
              <XAxis dataKey="date" stroke="#a1a1aa" fontSize={12} />
              <YAxis stroke="#a1a1aa" fontSize={12} tickFormatter={(v) => formatCurrency(v)} />
              <Tooltip 
                contentStyle={{ 
                  backgroundColor: '#18181b', 
                  border: '1px solid #3f3f46',
                  borderRadius: '8px',
                  color: '#fff'
                }}
                formatter={(value: number) => [formatCurrency(value), 'Cost']}
              />
              <Bar dataKey="Cost" fill="#f59e0b" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </DataCard>

        <DataCard title="Hourly Activity" subtitle="Messages by hour of day">
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={hourlyData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#3f3f46" />
              <XAxis dataKey="hour" stroke="#a1a1aa" fontSize={10} />
              <YAxis stroke="#a1a1aa" fontSize={12} />
              <Tooltip 
                contentStyle={{ 
                  backgroundColor: '#18181b', 
                  border: '1px solid #3f3f46',
                  borderRadius: '8px',
                  color: '#fff'
                }}
              />
              <Bar dataKey="messages" fill="#8b5cf6" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </DataCard>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <DataCard title="Top Destinations" subtitle="Most contacted numbers">
          <div className="space-y-3">
            {analytics.topDestinations.slice(0, 8).map((dest) => (
              <div key={dest.prefix} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Globe className="h-4 w-4 text-zinc-500" />
                  <span className="font-mono text-white">{dest.prefix}***</span>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-sm text-zinc-400">{dest.count.toLocaleString()}</span>
                  <span className={`text-sm font-medium ${
                    parseFloat(dest.deliveryRate) > 80 ? 'text-emerald-400' :
                    parseFloat(dest.deliveryRate) > 50 ? 'text-amber-400' : 'text-red-400'
                  }`}>
                    {dest.deliveryRate}%
                  </span>
                </div>
              </div>
            ))}
            {analytics.topDestinations.length === 0 && (
              <p className="text-center text-zinc-500 py-4">No destination data available</p>
            )}
          </div>
        </DataCard>

        <DataCard title="Additional Metrics" subtitle="More statistics">
          <div className="grid gap-4">
            <div className="flex items-center justify-between p-3 bg-zinc-800/50 rounded-lg">
              <div className="flex items-center gap-3">
                <Activity className="h-5 w-5 text-purple-400" />
                <span className="text-zinc-300">Total Segments</span>
              </div>
              <span className="text-xl font-bold text-white">{analytics.overview.totalSegments.toLocaleString()}</span>
            </div>
            <div className="flex items-center justify-between p-3 bg-zinc-800/50 rounded-lg">
              <div className="flex items-center gap-3">
                <Clock className="h-5 w-5 text-amber-400" />
                <span className="text-zinc-300">Pending Messages</span>
              </div>
              <span className="text-xl font-bold text-amber-400">{analytics.overview.pending.toLocaleString()}</span>
            </div>
            <div className="flex items-center justify-between p-3 bg-zinc-800/50 rounded-lg">
              <div className="flex items-center gap-3">
                <XCircle className="h-5 w-5 text-orange-400" />
                <span className="text-zinc-300">Rejected Messages</span>
              </div>
              <span className="text-xl font-bold text-orange-400">{analytics.overview.rejected.toLocaleString()}</span>
            </div>
            <div className="flex items-center justify-between p-3 bg-zinc-800/50 rounded-lg">
              <div className="flex items-center gap-3">
                <ArrowUp className="h-5 w-5 text-emerald-400" />
                <span className="text-zinc-300">Wallet Credits</span>
              </div>
              <span className="text-xl font-bold text-emerald-400">{formatCurrency(analytics.wallet.totalCredits)}</span>
            </div>
          </div>
        </DataCard>
      </div>

      <DataCard title="Status Breakdown" subtitle="Detailed status statistics">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="text-left text-xs font-semibold text-zinc-400 uppercase">
                <th className="pb-3">Status</th>
                <th className="pb-3 text-right">Count</th>
                <th className="pb-3 text-right">Percentage</th>
                <th className="pb-3 text-right">Visual</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/10">
              {analytics.statusBreakdown.map(row => (
                <tr key={row.status}>
                  <td className="py-3">
                    <span className="px-2 py-0.5 rounded text-xs font-medium capitalize" style={{
                      backgroundColor: `${STATUS_COLORS[row.status]}20`,
                      color: STATUS_COLORS[row.status]
                    }}>
                      {row.status}
                    </span>
                  </td>
                  <td className="py-3 text-right text-zinc-300">{row.count.toLocaleString()}</td>
                  <td className="py-3 text-right text-zinc-300">{row.percentage}%</td>
                  <td className="py-3 text-right">
                    <div className="w-32 h-2 bg-zinc-800 rounded-full overflow-hidden ml-auto">
                      <div 
                        className="h-full rounded-full"
                        style={{ 
                          width: `${Math.min(row.percentage, 100)}%`,
                          backgroundColor: STATUS_COLORS[row.status] || '#6B7280'
                        }}
                      />
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </DataCard>
    </div>
  )
}

function Wallet({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
    </svg>
  )
}
