import { NextResponse } from 'next/server'
import { db } from '@/lib/db'
import { wallets, walletTransactions } from '@/lib/schema'
import { requireServiceProvider } from '@/lib/service-provider'
import { eq, desc } from 'drizzle-orm'

export async function GET() {
  try {
    const sp = await requireServiceProvider()
    if (sp.response) return sp.response

    const spWallets = await db
      .select()
      .from(wallets)
      .where(eq(wallets.serviceProviderId, sp.serviceProviderId))
      .orderBy(desc(wallets.updatedAt))

    const walletData = await Promise.all(
      spWallets.map(async (wallet) => {
        const transactions = await db
          .select()
          .from(walletTransactions)
          .where(eq(walletTransactions.walletId, wallet.id))
          .orderBy(desc(walletTransactions.transactionDate))
          .limit(10)

        return {
          ...wallet,
          recentTransactions: transactions,
        }
      })
    )

    return NextResponse.json({ wallets: walletData })
  } catch (error) {
    console.error('Error fetching wallet:', error)
    return NextResponse.json({ error: 'Failed to fetch wallet' }, { status: 500 })
  }
}
