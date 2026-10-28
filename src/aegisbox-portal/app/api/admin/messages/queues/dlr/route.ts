import { db } from '@/lib/db'
import { dlrForwardingQueue, messages } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, and, gte, lte, like, sql } from 'drizzle-orm'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    const page = parseInt(searchParams.get('page') || '1')
    const limit = parseInt(searchParams.get('limit') || '50')
    const offset = (page - 1) * limit
    const status = searchParams.get('status')
    const dateFrom = searchParams.get('date_from')
    const dateTo = searchParams.get('date_to')

    const conditions = []

    if (status && status !== 'all') {
      conditions.push(eq(dlrForwardingQueue.status, status))
    }

    let dateCondition = null
    if (dateFrom && dateTo) {
      dateCondition = and(
        gte(dlrForwardingQueue.createdAt, new Date(dateFrom)),
        lte(dlrForwardingQueue.createdAt, new Date(dateTo))
      )
    } else if (dateFrom) {
      dateCondition = gte(dlrForwardingQueue.createdAt, new Date(dateFrom))
    } else if (dateTo) {
      dateCondition = lte(dlrForwardingQueue.createdAt, new Date(dateTo))
    }

    if (dateCondition) {
      conditions.push(dateCondition)
    }

    const whereClause = conditions.length > 0 ? and(...conditions) : undefined

    const [queueItems, totalResult] = await Promise.all([
      db
        .select({
          id: dlrForwardingQueue.id,
          messageId: dlrForwardingQueue.messageId,
          status: dlrForwardingQueue.status,
          attempts: dlrForwardingQueue.attempts,
          maxAttempts: dlrForwardingQueue.maxAttempts,
          lastAttemptAt: dlrForwardingQueue.lastAttemptAt,
          nextRetryAt: dlrForwardingQueue.nextRetryAt,
          errorMessage: dlrForwardingQueue.errorMessage,
          createdAt: dlrForwardingQueue.createdAt,
          lockedAt: dlrForwardingQueue.lockedAt,
          clientMessageId: messages.clientMessageId,
          clientRef: messages.clientRef,
        })
        .from(dlrForwardingQueue)
        .leftJoin(messages, eq(dlrForwardingQueue.messageId, messages.id))
        .where(whereClause)
        .orderBy(desc(dlrForwardingQueue.createdAt))
        .limit(limit)
        .offset(offset),
      db
        .select({ count: sql<number>`count(*)` })
        .from(dlrForwardingQueue)
        .where(whereClause),
    ])

    const total = totalResult[0]?.count || 0

    return NextResponse.json({
      queue: queueItems,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    })
  } catch (error) {
    console.error('[v0] Error fetching DLR queue:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
