import Link from 'next/link'
import { Wallet, Search, Plus, ArrowUpRight, Filter, Download, TrendingUp, TrendingDown, Minus } from 'lucide-react'
import { db } from '@/lib/db'
import { wallets, serviceProviders, walletTransactions } from '@/lib/schema'
import { desc, eq } from 'drizzle-orm'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

function getTransactionBadge(type: string) {
  const styles: Record<string, { bg: string; text: string; icon: any }> = {
    credit: { bg: 'bg-emerald-500/20', text: 'text-emerald-400', icon: TrendingUp },
    debit: { bg: 'bg-red-500/20', text: 'text-red-400', icon: TrendingDown },
    reversal: { bg: 'bg-amber-500/20', text: 'text-amber-400', icon: Minus },
  }
  
  const style = styles[type] || styles.debit
  const Icon = style.icon
  
  return (
    <span className={`flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${style.bg} ${style.text}`}>
      <Icon className="h-3 w-3" />
      <span className="capitalize">{type}</span>
    </span>
  )
}

async function getWallets() {
  try {
    const walletList = await db.select({
      id: wallets.id,
      serviceProviderId: wallets.serviceProviderId,
      currencyCode: wallets.currencyCode,
      balance: wallets.balance,
      walletType: wallets.walletType,
      createdAt: wallets.createdAt,
    })
      .from(wallets)
      .orderBy(desc(wallets.createdAt))

    const results = []
    for (const wallet of walletList) {
      const [sp] = await db.select({
        name: serviceProviders.name,
        email: serviceProviders.email,
      })
        .from(serviceProviders)
        .where(eq(serviceProviders.id, wallet.serviceProviderId))

      const balance = parseFloat(String(wallet.balance || '0'))

      const transactions = await db.select({
        id: walletTransactions.id,
        transactionType: walletTransactions.transactionType,
        amount: walletTransactions.amount,
        balanceAfter: walletTransactions.balanceAfter,
        description: walletTransactions.description,
        transactionDate: walletTransactions.transactionDate,
      })
        .from(walletTransactions)
        .where(eq(walletTransactions.walletId, wallet.id))
        .orderBy(desc(walletTransactions.transactionDate))
        .limit(5)

      results.push({
        id: wallet.id,
        serviceProviderId: wallet.serviceProviderId,
        spName: sp?.name || 'Unknown',
        spEmail: sp?.email || 'Unknown',
        currency: wallet.currencyCode,
        balance,
        walletType: wallet.walletType,
        createdAt: wallet.createdAt,
        transactions: transactions.map(t => ({
          ...t,
          amount: Number(t.amount),
          balanceAfter: Number(t.balanceAfter),
        })),
      })
    }

    return results
  } catch (error) {
    console.error('Error fetching wallets:', error)
    return []
  }
}

export default async function WalletsPage() {
  const wallets = await getWallets()
  const totalBalance = wallets.reduce((acc, w) => acc + w.balance, 0)

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Wallets"
        description="Manage service provider wallets and transactions"
      />

      <div className="grid gap-4 md:grid-cols-4">
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-blue-500/20">
              <Wallet className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Total Wallets</p>
              <p className="text-2xl font-bold text-white">{wallets.length}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-emerald-500/20">
              <ArrowUpRight className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Total Balance</p>
              <p className="text-2xl font-bold text-white">₦{totalBalance.toLocaleString()}</p>
            </div>
          </div>
        </DataCard>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
        <input
          type="text"
          placeholder="Search by service provider..."
          className="w-full pl-10 pr-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:ring-2 focus:ring-white/20"
        />
      </div>

      {wallets.length > 0 ? wallets.map((wallet) => (
        <DataCard key={wallet.id} title={wallet.spName} subtitle={`${wallet.spEmail} • ${wallet.currency} Wallet`}>
          <div className="grid gap-6 md:grid-cols-2">
            <div>
              <p className="text-sm text-zinc-400">Current Balance</p>
              <p className="text-2xl font-bold text-white">
                {wallet.currency === 'NGN' ? '₦' : '$'}{wallet.balance.toLocaleString()}
              </p>
              <p className="text-xs text-zinc-500 mt-1">
                Wallet Type: {wallet.walletType || 'prepaid'}
              </p>
            </div>
            <div className="flex items-center justify-end gap-2">
              <Link
                href={`/admin/service-providers/${wallet.serviceProviderId}`}
                className="text-sm text-zinc-400 hover:text-white transition-colors"
              >
                View SP
              </Link>
            </div>
          </div>

          {wallet.transactions && wallet.transactions.length > 0 && (
            <div className="mt-6">
              <div className="flex items-center justify-between mb-3">
                <h4 className="text-sm font-medium text-zinc-400">Recent Transactions</h4>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-zinc-800/50">
                    <tr>
                      <th className="px-4 py-2 text-left text-xs font-semibold text-zinc-400 uppercase">Type</th>
                      <th className="px-4 py-2 text-left text-xs font-semibold text-zinc-400 uppercase">Amount</th>
                      <th className="px-4 py-2 text-left text-xs font-semibold text-zinc-400 uppercase">Balance After</th>
                      <th className="px-4 py-2 text-left text-xs font-semibold text-zinc-400 uppercase">Description</th>
                      <th className="px-4 py-2 text-left text-xs font-semibold text-zinc-400 uppercase">Date</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/5">
                    {wallet.transactions.map((tx) => (
                      <tr key={tx.id} className="hover:bg-white/5 transition-colors">
                        <td className="px-4 py-3 whitespace-nowrap">
                          {getTransactionBadge(tx.transactionType)}
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap">
                          <span className={`font-medium ${
                            tx.transactionType === 'credit' || tx.transactionType === 'reversal' 
                              ? 'text-emerald-400' 
                              : 'text-red-400'
                          }`}>
                            {tx.transactionType === 'credit' || tx.transactionType === 'reversal' ? '+' : '-'}₦{tx.amount.toLocaleString()}
                          </span>
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap text-white">
                          ₦{tx.balanceAfter.toLocaleString()}
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap text-zinc-400">
                          {tx.description || '-'}
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap text-zinc-500 text-sm">
                          {new Date(tx.transactionDate).toLocaleString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </DataCard>
      )) : (
        <DataCard title="" className="!p-0">
          <div className="p-12 text-center text-zinc-500">
            No wallets found
          </div>
        </DataCard>
      )}
    </div>
  )
}
