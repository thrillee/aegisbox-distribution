import Link from 'next/link'
import { Plus, Search, DollarSign, Filter } from 'lucide-react'
import { db } from '@/lib/db'
import { pricingRules, serviceProviders, mnos } from '@/lib/schema'
import { desc, eq } from 'drizzle-orm'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

async function getPricingRules() {
  try {
    const rules = await db.select({
      id: pricingRules.id,
      serviceProviderId: pricingRules.serviceProviderId,
      mnoId: pricingRules.mnoId,
      currencyCode: pricingRules.currencyCode,
      pricePerSms: pricingRules.pricePerSms,
      createdAt: pricingRules.createdAt,
    })
      .from(pricingRules)
      .orderBy(desc(pricingRules.createdAt))

    const results = []
    for (const rule of rules) {
      let spName = 'Default'
      let mnoName = 'All MNOs'

      if (rule.serviceProviderId) {
        const [sp] = await db.select({ name: serviceProviders.name })
          .from(serviceProviders)
          .where(eq(serviceProviders.id, rule.serviceProviderId))
        spName = sp?.name || 'Unknown'
      }

      if (rule.mnoId) {
        const [mno] = await db.select({ name: mnos.name })
          .from(mnos)
          .where(eq(mnos.id, rule.mnoId))
        mnoName = mno?.name || 'Unknown'
      }

      results.push({
        id: rule.id,
        serviceProviderId: rule.serviceProviderId,
        serviceProvider: spName,
        mnoId: rule.mnoId,
        mno: mnoName,
        currency: rule.currencyCode,
        pricePerSms: parseFloat(String(rule.pricePerSms || '0')),
        createdAt: rule.createdAt,
      })
    }

    return results
  } catch (error) {
    console.error('Error fetching pricing rules:', error)
    return []
  }
}

export default async function PricingPage() {
  const pricingRules = await getPricingRules()

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Pricing"
        description="Manage SMS pricing rules for service providers"
        actions={
          <Link href="/admin/pricing/new">
            <Button className="bg-white text-black hover:bg-zinc-200">
              <Plus className="h-4 w-4" />
              Add Pricing Rule
            </Button>
          </Link>
        }
      />

      <div className="grid gap-4 md:grid-cols-3">
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-blue-500/20">
              <DollarSign className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Total Rules</p>
              <p className="text-2xl font-bold text-white">{pricingRules.length}</p>
            </div>
          </div>
        </DataCard>
      </div>

      <DataCard title="" className="!p-0">
        <div className="p-4 border-b border-white/10">
          <div className="flex items-center gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
              <input
                type="text"
                placeholder="Search pricing rules..."
                className="w-full pl-10 pr-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:ring-2 focus:ring-white/20"
              />
            </div>
            <Button variant="outline" className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5">
              <Filter className="h-4 w-4" />
              Filters
            </Button>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-zinc-800/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Service Provider</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">MNO</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Currency</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Price/SMS</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Created</th>
                <th className="px-6 py-3 text-right text-xs font-semibold text-zinc-400 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/10">
              {pricingRules.length > 0 ? pricingRules.map((rule) => (
                <tr key={rule.id} className="hover:bg-white/5 transition-colors">
                  <td className="px-6 py-4">
                    <Link href={`/admin/service-providers/${rule.serviceProviderId}`} className="font-medium text-white hover:text-zinc-300">
                      {rule.serviceProvider}
                    </Link>
                  </td>
                  <td className="px-6 py-4 text-white">{rule.mno}</td>
                  <td className="px-6 py-4 text-white">{rule.currency}</td>
                  <td className="px-6 py-4 font-medium text-white">
                    {rule.currency === 'NGN' ? '₦' : '$'}{rule.pricePerSms.toFixed(4)}
                  </td>
                  <td className="px-6 py-4 text-zinc-400">
                    {new Date(rule.createdAt).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <Link href={`/admin/pricing/${rule.id}`} className="text-sm text-white hover:underline">
                      Edit
                    </Link>
                  </td>
                </tr>
              )) : (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center text-zinc-500">
                    No pricing rules configured
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </DataCard>
    </div>
  )
}
