import { db } from '@/lib/db'
import { messages, messageSegments, serviceProviders, spCredentials, mnos } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, and, gte, lt, like, sql } from 'drizzle-orm'
import { isReportDate, reportDayRange } from '@/lib/time'
import { waitingForDlr } from '@/lib/pending-dlr'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    const page = Math.max(1, parseInt(searchParams.get('page') || '1') || 1)
    // At most 200 per page: the bulk manual DLR limit, so "select all" can be sent in one go.
    const limit = Math.min(200, Math.max(1, parseInt(searchParams.get('limit') || '50') || 50))
    const offset = (page - 1) * limit
    
    const status = searchParams.get('status')
    const spId = searchParams.get('sp_id')
    const mnoId = searchParams.get('mno_id')
    const clientRef = searchParams.get('client_ref')
    const sender = searchParams.get('sender')
    const destination = searchParams.get('destination')
    const dateFrom = searchParams.get('date_from')
    const dateTo = searchParams.get('date_to')

    const conditions = []

    if (status === 'awaiting_dlr') {
      conditions.push(waitingForDlr(sql`${messages.id}`, sql`${messages.finalStatus}`))
    } else if (status && status !== 'all') {
      conditions.push(eq(messages.processingStatus, status))
    }
    
    if (spId) {
      conditions.push(eq(messages.serviceProviderId, parseInt(spId)))
    }
    
    if (mnoId) {
      conditions.push(eq(messages.routedMnoId, parseInt(mnoId)))
    }
    
    if (clientRef) {
      conditions.push(like(messages.clientRef, `%${clientRef}%`))
    }

    if (sender) {
      conditions.push(like(messages.originalSourceAddr, `%${sender}%`))
    }

    if (destination) {
      conditions.push(like(messages.originalDestinationAddr, `%${destination}%`))
    }

    // Whole Lagos days, both dates included.
    if (dateFrom && isReportDate(dateFrom)) {
      conditions.push(gte(messages.receivedAt, reportDayRange(dateFrom, dateFrom).from))
    }
    if (dateTo && isReportDate(dateTo)) {
      conditions.push(lt(messages.receivedAt, reportDayRange(dateTo, dateTo).to))
    }

    const whereClause = conditions.length > 0 ? and(...conditions) : undefined

    const [messagesList, totalResult] = await Promise.all([
      db
        .select({
          id: messages.id,
          clientMessageId: messages.clientMessageId,
          clientRef: messages.clientRef,
          originalSourceAddr: messages.originalSourceAddr,
          sentSourceAddr: messages.sentSourceAddr,
          originalDestinationAddr: messages.originalDestinationAddr,
          shortMessage: sql<string>`substring(${messages.shortMessage} for 100)`,
          totalSegments: messages.totalSegments,
          receivedAt: messages.receivedAt,
          completedAt: messages.completedAt,
          processingStatus: messages.processingStatus,
          finalStatus: messages.finalStatus,
          cost: messages.cost,
          currencyCode: messages.currencyCode,
          serviceProviderId: messages.serviceProviderId,
          spName: serviceProviders.name,
          spCredentialId: messages.spCredentialId,
          spSystemId: spCredentials.systemId,
          routedMnoId: messages.routedMnoId,
          mnoName: mnos.name,
          // One row per message: a multipart message has one segment row each, so joining
          // segments listed it once per part. Report the segments' DLR status when they all
          // agree, MIXED when they differ, and null before any DLR arrives.
          dlrStatus: sql<string | null>`(
            SELECT CASE
              WHEN count(DISTINCT ${messageSegments.dlrStatus}) = 0 THEN NULL
              WHEN count(DISTINCT ${messageSegments.dlrStatus}) = 1 AND count(${messageSegments.dlrStatus}) = count(*) THEN max(${messageSegments.dlrStatus})
              ELSE 'MIXED'
            END
            FROM ${messageSegments}
            WHERE ${messageSegments.messageId} = ${messages.id}
          )`,
        })
        .from(messages)
        .leftJoin(serviceProviders, eq(messages.serviceProviderId, serviceProviders.id))
        .leftJoin(spCredentials, eq(messages.spCredentialId, spCredentials.id))
        .leftJoin(mnos, eq(messages.routedMnoId, mnos.id))
        .where(whereClause)
        .orderBy(desc(messages.receivedAt))
        .limit(limit)
        .offset(offset),
      db
        .select({ count: sql<number>`count(*)::int` })
        .from(messages)
        .where(whereClause),
    ])

    const total = totalResult[0]?.count || 0

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
    console.error('[v0] Error fetching messages:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
