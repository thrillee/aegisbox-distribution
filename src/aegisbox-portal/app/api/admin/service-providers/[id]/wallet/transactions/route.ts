import { db } from '@/lib/db'
import { walletTransactions, wallets } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq, desc, asc, and, gte, lte } from 'drizzle-orm'
import { transactionTypeEnum } from '@/lib/schema'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const spId = parseInt(resolvedParams.id)

    if (isNaN(spId)) {
      return NextResponse.json({ error: 'Invalid service provider ID' }, { status: 400 })
    }

    const { searchParams } = new URL(request.url)
    const page = parseInt(searchParams.get('page') || '1')
    const limit = parseInt(searchParams.get('limit') || '20')
    const type = searchParams.get('type')
    const status = searchParams.get('status')
    const startDate = searchParams.get('startDate')
    const endDate = searchParams.get('endDate')
    const sortOrder = searchParams.get('sort') === 'asc' ? asc : desc

    const [wallet] = await db
      .select()
      .from(wallets)
      .where(eq(wallets.serviceProviderId, spId))
      .limit(1)

    if (!wallet) {
      return NextResponse.json({ error: 'Wallet not found' }, { status: 404 })
    }

    const conditions = [eq(walletTransactions.walletId, wallet.id)]

    if (type && ['credit', 'debit', 'reversal'].includes(type)) {
      conditions.push(eq(walletTransactions.transactionType, type as 'credit' | 'debit' | 'reversal'))
    }

    if (startDate) {
      conditions.push(gte(walletTransactions.transactionDate, new Date(startDate)))
    }

    if (endDate) {
      conditions.push(lte(walletTransactions.transactionDate, new Date(endDate + 'T23:59:59')))
    }

    const transactions = await db
      .select({
        id: walletTransactions.id,
        transactionType: walletTransactions.transactionType,
        amount: walletTransactions.amount,
        balanceBefore: walletTransactions.balanceBefore,
        balanceAfter: walletTransactions.balanceAfter,
        description: walletTransactions.description,
        transactionDate: walletTransactions.transactionDate,
      })
      .from(walletTransactions)
      .where(and(...conditions))
      .orderBy(sortOrder(walletTransactions.transactionDate))
      .limit(limit)
      .offset((page - 1) * limit)

    const [totalCountResult] = await db
      .select({ count: walletTransactions.id })
      .from(walletTransactions)
      .where(and(...conditions))

    const [totalStats] = await db
      .select({
        totalCredits: walletTransactions.amount,
      })
      .from(walletTransactions)
      .where(and(eq(walletTransactions.walletId, wallet.id)))

    const creditsResult = await db
      .select({
        sum: walletTransactions.amount,
      })
      .from(walletTransactions)
      .where(and(
        eq(walletTransactions.walletId, wallet.id),
        eq(walletTransactions.transactionType, 'credit' as any)
      ))

    const debitsResult = await db
      .select({
        sum: walletTransactions.amount,
      })
      .from(walletTransactions)
      .where(and(
        eq(walletTransactions.walletId, wallet.id),
        eq(walletTransactions.transactionType, 'debit' as any)
      ))

    const totalCredits = Number(creditsResult[0]?.sum || 0)
    const totalDebits = Number(debitsResult[0]?.sum || 0)

    return NextResponse.json({
      wallet: {
        id: wallet.id,
        balance: wallet.balance,
        currencyCode: wallet.currencyCode,
      },
      transactions: transactions.map(t => ({
        ...t,
        amount: Number(t.amount),
        balanceBefore: Number(t.balanceBefore),
        balanceAfter: Number(t.balanceAfter),
      })),
      pagination: {
        page,
        limit,
        total: totalCountResult?.count || 0,
        totalPages: Math.ceil((totalCountResult?.count || 0) / limit),
      },
      summary: {
        totalCredits,
        totalDebits,
        netChange: totalCredits - totalDebits,
      },
    })
  } catch (error) {
    console.error('[v0] Error fetching wallet transactions:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
