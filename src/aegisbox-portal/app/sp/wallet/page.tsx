import { Wallet, TrendingUp, TrendingDown, Plus, Minus, Download } from 'lucide-react'
import { db } from '@/lib/db'
import { getCurrentServiceProviderId, NOT_LINKED_MESSAGE } from '@/lib/service-provider'
import { wallets, walletTransactions } from '@/lib/schema'
import { desc, eq, sql } from 'drizzle-orm'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'
import { formatMoney } from '@/lib/money'

async function getWalletData(serviceProviderId: number) {
  
  const walletList = await db
    .select({
      id: wallets.id,
      currencyCode: wallets.currencyCode,
      balance: wallets.balance,
      walletType: wallets.walletType,
    })
    .from(wallets)
    .where(eq(wallets.serviceProviderId, serviceProviderId))
    .orderBy(desc(wallets.updatedAt))

  const walletIds = walletList.map(w => w.id)

  let recentTransactions: any[] = []
  if (walletIds.length > 0) {
    recentTransactions = await db
      .select({
        id: walletTransactions.id,
        transactionType: walletTransactions.transactionType,
        amount: walletTransactions.amount,
        balanceAfter: walletTransactions.balanceAfter,
        description: walletTransactions.description,
        transactionDate: walletTransactions.transactionDate,
      })
      .from(walletTransactions)
      .where(eq(walletTransactions.walletId, walletIds[0]))
      .orderBy(desc(walletTransactions.transactionDate))
      .limit(20)
  }

  return {
    wallets: walletList.map(w => ({
      id: Number(w.id),
      currency: w.currencyCode || 'NGN',
      balance: w.balance ? Number(w.balance) : 0,
      walletType: w.walletType || 'prepaid',
    })),
    recentTransactions: recentTransactions.map(tx => ({
      id: Number(tx.id),
      type: tx.transactionType || 'debit',
      amount: tx.amount ? Number(tx.amount) : 0,
      balanceAfter: tx.balanceAfter ? Number(tx.balanceAfter) : 0,
      description: tx.description || 'Transaction',
      date: tx.transactionDate?.toISOString() || new Date().toISOString(),
    })),
  }
}

function getTransactionBadge(type: string) {
  const styles: Record<string, { bg: string; icon: any; text: string }> = {
    credit: { bg: 'bg-emerald-500/20 text-emerald-400', icon: TrendingUp, text: 'Credit' },
    debit: { bg: 'bg-red-500/20 text-red-400', icon: TrendingDown, text: 'Debit' },
    reversal: { bg: 'bg-amber-500/20 text-amber-400', icon: Minus, text: 'Reversal' },
  }

  const style = styles[type] || styles.debit
  const Icon = style.icon

  return (
    <span className={`flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${style.bg}`}>
      <Icon className="h-3 w-3" />
      <span>{style.text}</span>
    </span>
  )
}

function formatCurrency(amount: number, currency: string) {
  return formatMoney(amount, currency)
}

export default async function WalletPage() {
  const serviceProviderId = await getCurrentServiceProviderId()

  if (!serviceProviderId) {
    return (
      <div className="p-8 text-center">
        <p className="text-zinc-400">{NOT_LINKED_MESSAGE}</p>
      </div>
    )
  }

  const { wallets: walletList, recentTransactions } = await getWalletData(serviceProviderId)

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Wallet"
        description="Manage your wallet balance and transactions"
        actions={
          <Button className="bg-white text-black hover:bg-zinc-200">
            <Plus className="h-4 w-4" />
            Fund Wallet
          </Button>
        }
      />

      {walletList.length > 0 ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {walletList.map((wallet) => (
            <DataCard key={wallet.id} title="">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <Wallet className="h-5 w-5 text-white" />
                  <span className="font-medium text-white">{wallet.currency} Wallet</span>
                </div>
                <span className="px-2 py-0.5 bg-white/10 rounded text-xs text-zinc-400 capitalize">
                  {wallet.walletType}
                </span>
              </div>
              <p className="text-3xl font-bold text-white mb-4">
                {formatCurrency(wallet.balance, wallet.currency)}
              </p>
              <div className="flex gap-2 mt-4">
                <Button className="flex-1 bg-emerald-500 text-white hover:bg-emerald-600">
                  <Plus className="h-4 w-4" />
                  Add Funds
                </Button>
                <Button variant="outline" className="flex-1 border-white/10 text-zinc-400 hover:text-white hover:bg-white/5">
                  <Minus className="h-4 w-4" />
                  Debit
                </Button>
              </div>
            </DataCard>
          ))}
        </div>
      ) : (
        <DataCard title="">
          <div className="p-8 text-center">
            <Wallet className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
            <p className="text-zinc-400">No wallet found for your organization.</p>
            <p className="text-sm text-zinc-500 mt-1">Contact support to set up your wallet.</p>
          </div>
        </DataCard>
      )}

      <DataCard title="Recent Transactions" subtitle="Your recent wallet activity" action={{ label: 'Export', href: '#' }}>
        {recentTransactions.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-zinc-800/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Type</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Amount</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Balance After</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Description</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Date</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/10">
                {recentTransactions.map((tx) => (
                  <tr key={tx.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">{getTransactionBadge(tx.type)}</td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`font-medium ${
                        tx.type === 'credit' ? 'text-emerald-400' :
                        tx.type === 'reversal' ? 'text-amber-400' : 'text-red-400'
                      }`}>
                        {tx.type === 'credit' || tx.type === 'reversal' ? '+' : '-'}{formatMoney(tx.amount, walletList[0]?.currency)}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-white">{formatMoney(tx.balanceAfter, walletList[0]?.currency)}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-zinc-400">{tx.description}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-zinc-400">{new Date(tx.date).toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="p-8 text-center text-zinc-500">No transactions found</div>
        )}
      </DataCard>
    </div>
  )
}
