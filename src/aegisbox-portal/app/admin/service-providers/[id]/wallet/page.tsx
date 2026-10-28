import Link from 'next/link'
import { ArrowLeft, Wallet, TrendingUp, TrendingDown, Plus, Minus, Send, Download } from 'lucide-react'

async function getWalletData(spId: string) {
  return {
    serviceProviderId: parseInt(spId),
    wallets: [
      { id: 1, currency: 'NGN', balance: 500000, walletType: 'prepaid', creditLimit: 0 },
      { id: 2, currency: 'USD', balance: 2500, walletType: 'prepaid', creditLimit: 0 },
    ],
    recentTransactions: [
      { id: 1, type: 'credit', amount: 100000, balanceAfter: 500000, description: 'Manual funding', date: '2024-06-15 10:30' },
      { id: 2, type: 'debit', amount: 2500, balanceAfter: 400000, description: 'SMS cost - 45000 messages', date: '2024-06-14 23:59' },
      { id: 3, type: 'credit', amount: 50000, balanceAfter: 425000, description: 'Manual funding', date: '2024-06-13 14:20' },
      { id: 4, type: 'debit', amount: 1800, balanceAfter: 375000, description: 'SMS cost - 32000 messages', date: '2024-06-13 08:15' },
      { id: 5, type: 'reversal', amount: 500, balanceAfter: 376800, description: 'Refund - undelivered message', date: '2024-06-12 16:45' },
    ],
  }
}

function getTransactionBadge(type: string) {
  const styles: Record<string, { bg: string; text: string; icon: any }> = {
    credit: { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-600 dark:text-green-400', icon: TrendingUp },
    debit: { bg: 'bg-red-100 dark:bg-red-900/30', text: 'text-red-600 dark:text-red-400', icon: TrendingDown },
    reversal: { bg: 'bg-yellow-100 dark:bg-yellow-900/30', text: 'text-yellow-600 dark:text-yellow-400', icon: Minus },
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

export default async function SPWalletPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const walletData = await getWalletData(id)

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center gap-4">
        <Link
          href={`/admin/service-providers/${id}`}
          className="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
        >
          <ArrowLeft className="h-5 w-5 text-slate-600 dark:text-slate-400" />
        </Link>
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white">Wallet Management</h1>
          <p className="text-slate-600 dark:text-slate-400">Manage wallet balance and transactions</p>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {walletData.wallets.map((wallet) => (
          <div key={wallet.id} className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <Wallet className="h-5 w-5 text-primary" />
                <span className="font-medium text-slate-900 dark:text-white">{wallet.currency} Wallet</span>
              </div>
              <span className="px-2 py-0.5 bg-slate-100 dark:bg-slate-700 rounded text-xs text-slate-600 dark:text-slate-400 capitalize">
                {wallet.walletType}
              </span>
            </div>
            <p className="text-3xl font-bold text-slate-900 dark:text-white mb-4">
              {wallet.currency === 'NGN' ? '₦' : '$'}{wallet.balance.toLocaleString()}
            </p>
            {wallet.walletType === 'postpaid' && (
              <p className="text-sm text-slate-500">
                Credit Limit: {wallet.currency === 'NGN' ? '₦' : '$'}{wallet.creditLimit.toLocaleString()}
              </p>
            )}
            <div className="flex gap-2 mt-4">
              <button className="flex-1 flex items-center justify-center gap-2 px-3 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors text-sm">
                <Plus className="h-4 w-4" />
                Fund
              </button>
              <button className="flex-1 flex items-center justify-center gap-2 px-3 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors text-sm">
                <Minus className="h-4 w-4" />
                Debit
              </button>
            </div>
          </div>
        ))}
      </div>

      <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
        <div className="p-6 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-900 dark:text-white">Recent Transactions</h2>
          <button className="flex items-center gap-2 px-3 py-1.5 border border-slate-200 dark:border-slate-600 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors text-sm">
            <Download className="h-4 w-4" />
            Export
          </button>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-slate-50 dark:bg-slate-700/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  Amount
                </th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  Balance After
                </th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  Description
                </th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  Date
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200 dark:divide-slate-700">
              {walletData.recentTransactions.map((tx) => (
                <tr key={tx.id} className="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    {getTransactionBadge(tx.type)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`font-medium ${
                      tx.type === 'credit' ? 'text-green-600 dark:text-green-400' : 
                      tx.type === 'reversal' ? 'text-yellow-600 dark:text-yellow-400' :
                      'text-red-600 dark:text-red-400'
                    }`}>
                      {tx.type === 'credit' || tx.type === 'reversal' ? '+' : '-'}₦{tx.amount.toLocaleString()}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-slate-900 dark:text-white">
                    ₦{tx.balanceAfter.toLocaleString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-slate-600 dark:text-slate-400">
                    {tx.description}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-slate-600 dark:text-slate-400">
                    {new Date(tx.date).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
