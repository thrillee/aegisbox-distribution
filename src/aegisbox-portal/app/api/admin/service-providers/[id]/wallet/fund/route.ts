import { db } from '@/lib/db'
import { wallets, walletTransactions } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq, and } from 'drizzle-orm'

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const spId = parseInt(resolvedParams.id)

    if (isNaN(spId)) {
      return NextResponse.json({ error: 'Invalid service provider ID' }, { status: 400 })
    }

    const body = await request.json()
    const { amount, currency, description } = body

    if (!amount || amount <= 0) {
      return NextResponse.json({ error: 'Valid amount is required' }, { status: 400 })
    }

    const [wallet] = await db
      .select()
      .from(wallets)
      .where(
        and(
          eq(wallets.serviceProviderId, spId),
          eq(wallets.currencyCode, currency || 'NGN')
        )
      )
      .limit(1)

    if (!wallet) {
      return NextResponse.json({ error: 'Wallet not found' }, { status: 404 })
    }

    const currentBalance = Number(wallet.balance) || 0
    const newBalance = currentBalance + amount

    await db
      .update(wallets)
      .set({
        balance: newBalance.toString(),
      })
      .where(eq(wallets.id, wallet.id))

    await db
      .insert(walletTransactions)
      .values({
        walletId: wallet.id,
        transactionType: 'credit',
        amount: amount.toString(),
        balanceBefore: currentBalance.toString(),
        balanceAfter: newBalance.toString(),
        description: description || 'Manual funding',
      })

    return NextResponse.json({
      success: true,
      newBalance,
      transactionId: wallet.id,
    })
  } catch (error) {
    console.error('[v0] Error funding wallet:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
