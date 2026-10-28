import Link from 'next/link'
import { Plus, Search, Filter, Download, MoreHorizontal } from 'lucide-react'
import { db } from '@/lib/db'
import { serviceProviders, wallets, messages, spCredentials, routingGroups } from '@/lib/schema'
import { desc, eq, sql, count, ilike, or } from 'drizzle-orm'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

async function getServiceProviders(search: string | undefined, status: string | undefined) {
  try {
    let query = db.select({
      id: serviceProviders.id,
      name: serviceProviders.name,
      email: serviceProviders.email,
      status: serviceProviders.status,
      defaultCurrencyCode: serviceProviders.defaultCurrencyCode,
      createdAt: serviceProviders.createdAt,
    })
      .from(serviceProviders)

    if (search) {
      query = query.where(or(
        ilike(serviceProviders.name, `%${search}%`),
        ilike(serviceProviders.email, `%${search}%`)
      )) as any
    }

    if (status) {
      query = query.where(eq(serviceProviders.status, status as any)) as any
    }

    const providers = await query.orderBy(desc(serviceProviders.createdAt))

    const results = []
    const today = new Date()
    today.setHours(0, 0, 0, 0)

    for (const sp of providers) {
      const [wallet] = await db.select({
        balance: wallets.balance,
      })
        .from(wallets)
        .where(eq(wallets.serviceProviderId, sp.id))
        .limit(1)

      const [messageCount] = await db.select({ count: count() })
        .from(messages)
        .where(eq(messages.serviceProviderId, sp.id))

      const [credential] = await db.select({
        routingGroupId: spCredentials.routingGroupId,
      })
        .from(spCredentials)
        .where(eq(spCredentials.serviceProviderId, sp.id))
        .limit(1)

      let routingGroupName = 'Not Assigned'
      if (credential?.routingGroupId) {
        const [rg] = await db.select({ name: routingGroups.name })
          .from(routingGroups)
          .where(eq(routingGroups.id, credential.routingGroupId))
        routingGroupName = rg?.name || 'Not Assigned'
      }

      results.push({
        id: sp.id,
        name: sp.name,
        email: sp.email,
        status: sp.status,
        currency: sp.defaultCurrencyCode,
        walletBalance: parseFloat(wallet?.balance || '0'),
        messagesToday: parseInt(String(messageCount?.count || '0')),
        routingGroup: routingGroupName,
        createdAt: sp.createdAt,
      })
    }

    return results
  } catch (error) {
    console.error('Error fetching service providers:', error)
    return []
  }
}

function getStatusBadge(status: string) {
  const styles = {
    active: 'bg-emerald-500/20 text-emerald-400',
    pending: 'bg-amber-500/20 text-amber-400',
    suspended: 'bg-red-500/20 text-red-400',
    inactive: 'bg-zinc-500/20 text-zinc-400',
  }
  
  return (
    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status as keyof typeof styles] || styles.inactive}`}>
      {status}
    </span>
  )
}

export default async function ServiceProvidersPage({
  searchParams,
}: {
  searchParams: Promise<{ search?: string; status?: string }>
}) {
  const resolvedParams = await searchParams
  const providers = await getServiceProviders(resolvedParams.search, resolvedParams.status)

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Service Providers"
        description="Manage your service provider accounts"
        actions={
          <Link
            href="/admin/service-providers/new"
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
          >
            <Plus className="h-4 w-4" />
            Add Service Provider
          </Link>
        }
      />

      <DataCard title="" className="!p-0">
        <div className="p-4 border-b border-white/10">
          <div className="flex items-center gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
              <input
                type="text"
                placeholder="Search by name or email..."
                className="w-full pl-10 pr-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white placeholder:text-zinc-500 focus:ring-2 focus:ring-white/20 focus:border-transparent"
              />
            </div>
            <select className="px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20">
              <option value="">All Status</option>
              <option value="active">Active</option>
              <option value="pending">Pending</option>
              <option value="suspended">Suspended</option>
              <option value="inactive">Inactive</option>
            </select>
            <Button variant="outline" className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5">
              <Download className="h-4 w-4" />
              Export
            </Button>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-zinc-800/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                  Service Provider
                </th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                  Wallet
                </th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                  Messages
                </th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                  Routing Group
                </th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                  Created
                </th>
                <th className="px-6 py-3 text-right text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/10">
              {providers.length > 0 ? providers.map((provider) => (
                <tr key={provider.id} className="hover:bg-white/5 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div>
                      <Link 
                        href={`/admin/service-providers/${provider.id}`}
                        className="font-medium text-white hover:text-zinc-300 transition-colors"
                      >
                        {provider.name}
                      </Link>
                      <p className="text-sm text-zinc-400">{provider.email}</p>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {getStatusBadge(provider.status)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div>
                      <p className="font-medium text-white">
                        {provider.currency === 'NGN' ? '₦' : '$'}{provider.walletBalance.toLocaleString()}
                      </p>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <p className="font-medium text-white">
                      {provider.messagesToday.toLocaleString()}
                    </p>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <p className="text-sm text-zinc-400">
                      {provider.routingGroup}
                    </p>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <p className="text-sm text-zinc-400">
                      {new Date(provider.createdAt).toLocaleDateString()}
                    </p>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="flex items-center justify-end gap-2">
                      <Link
                        href={`/admin/service-providers/${provider.id}`}
                        className="px-3 py-1.5 text-sm font-medium text-white hover:bg-white/10 rounded-lg transition-colors"
                      >
                        View
                      </Link>
                      <Link
                        href={`/admin/service-providers/${provider.id}/edit`}
                        className="px-3 py-1.5 text-sm font-medium text-zinc-400 hover:bg-white/10 rounded-lg transition-colors"
                      >
                        Edit
                      </Link>
                    </div>
                  </td>
                </tr>
              )) : (
                <tr>
                  <td colSpan={7} className="px-6 py-12 text-center text-zinc-500">
                    No service providers found
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <div className="p-4 border-t border-white/10">
          <div className="flex items-center justify-between">
            <p className="text-sm text-zinc-400">
              Showing {providers.length} service providers
            </p>
            <div className="flex items-center gap-2">
              <Button variant="outline" disabled className="border-white/10 text-zinc-500">Previous</Button>
              <Button variant="outline" className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5">Next</Button>
            </div>
          </div>
        </div>
      </DataCard>
    </div>
  )
}
