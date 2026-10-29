'use client'

import { useState, useEffect, useCallback } from 'react'
import { reportDate, reportDayStart } from '@/lib/time'
import { Download, FileText, Filter, Calendar, Loader2 } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

interface ServiceProvider {
  id: number
  name: string
  email: string
}

interface DailyReport {
  reportDate: string
  serviceProviderId: number
  serviceProviderName: string
  serviceProviderEmail: string
  mnoId: number
  mnoName: string
  currencyCode: string
  totalReceived: number
  totalSegmentsReceived: number
  totalSegmentsSubmitted: number
  totalSegmentsDelivered: number
  totalSegmentsFailed: number
  totalSegmentsRejected: number
  netCost: number
}

const STATUS_OPTIONS = [
  { value: '', label: 'All Statuses' },
  { value: 'delivered', label: 'Delivered' },
  { value: 'failed', label: 'Failed' },
  { value: 'rejected', label: 'Rejected' },
  { value: 'pending', label: 'Pending' },
  { value: 'queued', label: 'Queued' },
  { value: 'processing', label: 'Processing' },
  { value: 'submitted', label: 'Submitted' },
  { value: 'accepted', label: 'Accepted' },
  { value: 'forwarded', label: 'Forwarded' },
  { value: 'undeliverable', label: 'Undeliverable' },
]

const EXPORT_FIELDS = [
  { key: 'reportDate', label: 'Date', enabled: true },
  { key: 'serviceProviderName', label: 'Service Provider', enabled: true },
  { key: 'serviceProviderEmail', label: 'Email', enabled: true },
  { key: 'mnoName', label: 'MNO', enabled: true },
  { key: 'currencyCode', label: 'Currency', enabled: true },
  { key: 'totalReceived', label: 'Total Received', enabled: true },
  { key: 'totalSegmentsReceived', label: 'Segments Received', enabled: true },
  { key: 'totalSegmentsSubmitted', label: 'Segments Submitted', enabled: true },
  { key: 'totalSegmentsDelivered', label: 'Segments Delivered', enabled: true },
  { key: 'totalSegmentsFailed', label: 'Segments Failed', enabled: true },
  { key: 'totalSegmentsRejected', label: 'Segments Rejected', enabled: true },
  { key: 'netCost', label: 'Net Cost', enabled: true },
]

export default function DailyReportPage() {
  const [loading, setLoading] = useState(true)
  const [initialLoad, setInitialLoad] = useState(true)
  const [reportData, setReportData] = useState<DailyReport[]>([])
  const [serviceProviders, setServiceProviders] = useState<ServiceProvider[]>([])
  const [filters, setFilters] = useState(() => {
    return {
      start_date: reportDate(reportDayStart(7)),
      end_date: reportDate(),
      sp_id: '',
      status: '',
    }
  })
  const [exportFields, setExportFields] = useState(EXPORT_FIELDS.map(f => ({ ...f })))
  const [showFieldSelector, setShowFieldSelector] = useState(false)
  const [error, setError] = useState('')

  const fetchServiceProviders = useCallback(async () => {
    try {
      console.log('Fetching service providers...')
      const res = await fetch('/api/admin/service-providers')
      console.log('Response status:', res.status)
      if (res.ok) {
        const data = await res.json()
        console.log('Service providers data:', data)
        const sps = data.providers || data.serviceProviders || []
        console.log('Setting service providers:', sps)
        setServiceProviders(sps)
      } else {
        console.error('Failed to fetch service providers:', res.statusText)
      }
    } catch (err) {
      console.error('Error fetching service providers:', err)
    }
  }, [])

  useEffect(() => {
    fetchServiceProviders()
  }, [fetchServiceProviders])

  const fetchReport = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const params = new URLSearchParams()
      params.set('start_date', filters.start_date)
      params.set('end_date', filters.end_date)
      if (filters.sp_id) params.set('sp_id', filters.sp_id)
      if (filters.status) params.set('status', filters.status)

      const res = await fetch(`/api/admin/messages/daily-report?${params}`)
      if (res.ok) {
        const data = await res.json()
        setReportData(data.data || [])
      } else {
        setError('Failed to fetch report')
      }
    } catch (err) {
      setError('An error occurred')
    } finally {
      setLoading(false)
      setInitialLoad(false)
    }
  }, [filters])

  useEffect(() => {
    fetchReport()
  }, [fetchReport])

  const exportToCSV = () => {
    const enabledFields = exportFields.filter(f => f.enabled)
    const headers = enabledFields.map(f => f.label).join(',')
    const rows = reportData.map(row => 
      enabledFields.map(f => {
        const value = row[f.key as keyof DailyReport]
        return typeof value === 'string' && value.includes(',') ? `"${value}"` : value
      }).join(',')
    )
    const csv = [headers, ...rows].join('\n')
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `daily-report-${filters.start_date}-${filters.end_date}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  const exportToJSON = () => {
    const enabledFields = exportFields.filter(f => f.enabled)
    const data = reportData.map(row => {
      const obj: Record<string, any> = {}
      enabledFields.forEach(f => {
        obj[f.label] = row[f.key as keyof DailyReport]
      })
      return obj
    })
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `daily-report-${filters.start_date}-${filters.end_date}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  const toggleField = (key: string) => {
    setExportFields(prev => prev.map(f => 
      f.key === key ? { ...f, enabled: !f.enabled } : f
    ))
  }

  const formatNumber = (num: number) => num.toLocaleString()
  const formatCurrency = (amount: number, currency: string) => 
    `${currency} ${amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`

  if (initialLoad) {
    return <PageLoadingSkeleton />
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Daily SMS Usage Report"
        description="Detailed SMS usage statistics by service provider"
        actions={
          <div className="flex items-center gap-2">
            <div className="relative">
              <Button
                onClick={() => setShowFieldSelector(!showFieldSelector)}
                variant="outline"
                className="border-white/10 bg-zinc-800 text-white hover:bg-zinc-700"
              >
                <FileText className="h-4 w-4 mr-2" />
                Select Fields
              </Button>
              {showFieldSelector && (
                <div className="absolute right-0 top-full mt-2 w-64 bg-zinc-900 border border-white/10 rounded-xl p-4 z-50 shadow-xl">
                  <h4 className="text-sm font-medium text-white mb-3">Export Fields</h4>
                  <div className="space-y-2 max-h-64 overflow-y-auto">
                    {exportFields.map(field => (
                      <label key={field.key} className="flex items-center gap-2 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={field.enabled}
                          onChange={() => toggleField(field.key)}
                          className="rounded border-white/20 bg-zinc-800"
                        />
                        <span className="text-sm text-zinc-300">{field.label}</span>
                      </label>
                    ))}
                  </div>
                  <div className="flex gap-2 mt-4 pt-3 border-t border-white/10">
                    <Button
                      size="sm"
                      onClick={() => setExportFields(prev => prev.map(f => ({ ...f, enabled: true })))}
                      variant="ghost"
                      className="text-xs text-zinc-400 hover:text-white"
                    >
                      Select All
                    </Button>
                    <Button
                      size="sm"
                      onClick={() => setExportFields(prev => prev.map(f => ({ ...f, enabled: false })))}
                      variant="ghost"
                      className="text-xs text-zinc-400 hover:text-white"
                    >
                      Deselect All
                    </Button>
                  </div>
                </div>
              )}
            </div>
            <Button
              onClick={exportToCSV}
              disabled={reportData.length === 0}
              className="bg-emerald-600 text-white hover:bg-emerald-700"
            >
              <Download className="h-4 w-4 mr-2" />
              CSV
            </Button>
            <Button
              onClick={exportToJSON}
              disabled={reportData.length === 0}
              className="bg-amber-600 text-white hover:bg-amber-700"
            >
              <FileText className="h-4 w-4 mr-2" />
              JSON
            </Button>
          </div>
        }
      />

      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400">
          {error}
        </div>
      )}

      <DataCard title="Filters" subtitle="Configure report filters">
        <div className="grid gap-4 md:grid-cols-5">
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Start Date</label>
            <input
              type="date"
              value={filters.start_date}
              onChange={(e) => setFilters(prev => ({ ...prev, start_date: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">End Date</label>
            <input
              type="date"
              value={filters.end_date}
              onChange={(e) => setFilters(prev => ({ ...prev, end_date: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Service Provider</label>
            <select
              value={filters.sp_id}
              onChange={(e) => setFilters(prev => ({ ...prev, sp_id: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
              disabled={serviceProviders.length === 0}
            >
              <option value="">All Service Providers</option>
              {serviceProviders.length === 0 ? (
                <option disabled>Loading...</option>
              ) : (
                serviceProviders.map(sp => (
                  <option key={sp.id} value={sp.id}>{sp.name}</option>
                ))
              )}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Status</label>
            <select
              value={filters.status}
              onChange={(e) => setFilters(prev => ({ ...prev, status: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            >
              {STATUS_OPTIONS.map(opt => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div className="flex items-end">
            <Button
              onClick={fetchReport}
              disabled={loading}
              className="w-full bg-white text-black hover:bg-zinc-200"
            >
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Filter className="h-4 w-4 mr-2" />}
              Apply Filters
            </Button>
          </div>
        </div>
      </DataCard>

      {loading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
        </div>
      ) : reportData.length === 0 ? (
        <DataCard title="Report Results">
          <div className="text-center py-12 text-zinc-500">
            No data available for the selected filters
          </div>
        </DataCard>
      ) : (
        <DataCard title="Report Results" subtitle={`${reportData.length} records found`}>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-xs font-semibold text-zinc-400 uppercase border-b border-white/10">
                  <th className="pb-3">Date</th>
                  <th className="pb-3">Service Provider</th>
                  <th className="pb-3">MNO</th>
                  <th className="pb-3 text-right">Received</th>
                  <th className="pb-3 text-right">Segments Recvd</th>
                  <th className="pb-3 text-right">Segments Sub</th>
                  <th className="pb-3 text-right">Delivered</th>
                  <th className="pb-3 text-right">Failed</th>
                  <th className="pb-3 text-right">Rejected</th>
                  <th className="pb-3 text-right">Net Cost</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/10">
                {reportData.map((row, idx) => (
                  <tr key={idx} className="hover:bg-white/5">
                    <td className="py-3 text-white">{row.reportDate}</td>
                    <td className="py-3">
                      <div className="text-white">{row.serviceProviderName}</div>
                      <div className="text-xs text-zinc-500">{row.serviceProviderEmail}</div>
                    </td>
                    <td className="py-3 text-zinc-300">{row.mnoName}</td>
                    <td className="py-3 text-right text-zinc-300">{formatNumber(row.totalReceived)}</td>
                    <td className="py-3 text-right text-zinc-300">{formatNumber(row.totalSegmentsReceived)}</td>
                    <td className="py-3 text-right text-blue-400">{formatNumber(row.totalSegmentsSubmitted)}</td>
                    <td className="py-3 text-right text-emerald-400">{formatNumber(row.totalSegmentsDelivered)}</td>
                    <td className="py-3 text-right text-red-400">{formatNumber(row.totalSegmentsFailed)}</td>
                    <td className="py-3 text-right text-orange-400">{formatNumber(row.totalSegmentsRejected)}</td>
                    <td className="py-3 text-right font-medium text-white">
                      {formatCurrency(row.netCost, row.currencyCode)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          
          <div className="mt-6 pt-4 border-t border-white/10">
            <h4 className="text-sm font-medium text-white mb-3">Summary</h4>
            <div className="grid gap-4 md:grid-cols-4">
              <div className="bg-zinc-800/50 rounded-lg p-3">
                <p className="text-xs text-zinc-400">Total Received</p>
                <p className="text-xl font-bold text-white">
                  {formatNumber(reportData.reduce((sum, r) => sum + Number(r.totalReceived), 0))}
                </p>
              </div>
              <div className="bg-zinc-800/50 rounded-lg p-3">
                <p className="text-xs text-zinc-400">Total Segments</p>
                <p className="text-xl font-bold text-blue-400">
                  {formatNumber(reportData.reduce((sum, r) => sum + Number(r.totalSegmentsReceived), 0))}
                </p>
              </div>
              <div className="bg-zinc-800/50 rounded-lg p-3">
                <p className="text-xs text-zinc-400">Total Delivered</p>
                <p className="text-xl font-bold text-emerald-400">
                  {formatNumber(reportData.reduce((sum, r) => sum + Number(r.totalSegmentsDelivered), 0))}
                </p>
              </div>
              <div className="bg-zinc-800/50 rounded-lg p-3">
                <p className="text-xs text-zinc-400">Total Cost</p>
                <p className="text-xl font-bold text-white">
                  {reportData[0]?.currencyCode} {reportData.reduce((sum, r) => sum + Number(r.netCost), 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                </p>
              </div>
            </div>
          </div>
        </DataCard>
      )}
    </div>
  )
}
