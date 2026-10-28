import { DollarSign, RefreshCw } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

export default function ExchangeRatesPage() {
  const rates = [
    { from: 'USD', to: 'NGN', rate: 1500.00, updatedAt: new Date().toISOString() },
    { from: 'EUR', to: 'NGN', rate: 1625.00, updatedAt: new Date().toISOString() },
    { from: 'GBP', to: 'NGN', rate: 1900.00, updatedAt: new Date().toISOString() },
    { from: 'GHS', to: 'NGN', rate: 95.50, updatedAt: new Date().toISOString() },
    { from: 'KES', to: 'NGN', rate: 9.75, updatedAt: new Date().toISOString() },
    { from: 'ZAR', to: 'NGN', rate: 78.00, updatedAt: new Date().toISOString() },
  ]

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Exchange Rates"
        description="Manage currency exchange rates"
        actions={
          <Button className="bg-white text-black hover:bg-zinc-200">
            <RefreshCw className="h-4 w-4" />
            Refresh Rates
          </Button>
        }
      />

      <div className="grid gap-4 md:grid-cols-3">
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-blue-500/20">
              <DollarSign className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Base Currency</p>
              <p className="text-2xl font-bold text-white">NGN</p>
            </div>
          </div>
        </DataCard>
      </div>

      <DataCard title="Current Rates" subtitle="Exchange rates against NGN">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-zinc-800/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">From</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">To</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Rate</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Last Updated</th>
                <th className="px-6 py-3 text-right text-xs font-semibold text-zinc-400 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/10">
              {rates.map((rate, i) => (
                <tr key={i} className="hover:bg-white/5 transition-colors">
                  <td className="px-6 py-4 font-medium text-white">{rate.from}</td>
                  <td className="px-6 py-4 text-white">{rate.to}</td>
                  <td className="px-6 py-4 font-medium text-emerald-400">{rate.rate.toFixed(2)}</td>
                  <td className="px-6 py-4 text-zinc-400">
                    {new Date(rate.updatedAt).toLocaleString()}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button className="text-sm text-white hover:underline">Edit</button>
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
