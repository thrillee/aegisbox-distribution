import { NextRequest, NextResponse } from 'next/server'
import { db } from '@/lib/db'
import { messages } from '@/lib/schema'
import { requireServiceProvider } from '@/lib/service-provider'
import { desc, gte, lte, eq, and, or, ilike, sql } from 'drizzle-orm'

const MAX_PAGE_SIZE = 200

// Lists the provider's messages. Sending goes through the HTTP or SMPP API so that
// validation, pricing and wallet debits always apply.
export async function GET(request: NextRequest) {
  try {
    const sp = await requireServiceProvider()
    if (sp.response) return sp.response

    const { searchParams } = new URL(request.url)
    const page = Math.max(1, parseInt(searchParams.get('page') || '1') || 1)
    const limit = Math.min(MAX_PAGE_SIZE, Math.max(1, parseInt(searchParams.get('limit') || '50') || 50))
    const status = searchParams.get('status')
    const senderId = searchParams.get('senderId')
    const startDate = searchParams.get('startDate')
    const endDate = searchParams.get('endDate')
    const search = searchParams.get('search')

    const offset = (page - 1) * limit

    const conditions = [eq(messages.serviceProviderId, sp.serviceProviderId)]
    if (status) {
      conditions.push(eq(messages.finalStatus, status))
    }
    if (senderId) {
      conditions.push(eq(messages.originalSourceAddr, senderId))
    }
    if (startDate && !Number.isNaN(Date.parse(startDate))) {
      conditions.push(gte(messages.submittedAt, new Date(startDate)))
    }
    if (endDate && !Number.isNaN(Date.parse(endDate))) {
      conditions.push(lte(messages.submittedAt, new Date(endDate)))
    }
    if (search) {
      const pattern = `%${search}%`
      conditions.push(
        or(
          sql`${messages.id}::text LIKE ${pattern}`,
          ilike(messages.originalDestinationAddr, pattern),
          ilike(messages.clientMessageId, pattern),
          ilike(messages.clientRef, pattern)
        )!
      )
    }

    const whereClause = and(...conditions)

    const [messagesList, totalCount] = await Promise.all([
      db
        .select()
        .from(messages)
        .where(whereClause)
        .orderBy(desc(messages.submittedAt))
        .limit(limit)
        .offset(offset),
      db
        .select({ count: sql<number>`count(*)::int` })
        .from(messages)
        .where(whereClause),
    ])

    const total = totalCount[0]?.count || 0
    return NextResponse.json({
      messages: messagesList,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    })
  } catch (error) {
    console.error('Error fetching messages:', error)
    return NextResponse.json({ error: 'Failed to fetch messages' }, { status: 500 })
  }
}
