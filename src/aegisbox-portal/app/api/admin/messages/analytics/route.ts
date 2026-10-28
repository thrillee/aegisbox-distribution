import { db } from '@/lib/db'
import { messages, messageSegments, serviceProviders, spCredentials, mnos } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, and, gte, lte, sql, isNull, not, inArray } from 'drizzle-orm'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    const spId = searchParams.get('sp_id')
    const mnoId = searchParams.get('mno_id')
    const dateFrom = searchParams.get('date_from')
    const dateTo = searchParams.get('date_to')

    let dateCondition = undefined
    if (dateFrom && dateTo) {
      dateCondition = and(gte(messages.receivedAt, new Date(dateFrom)), lte(messages.receivedAt, new Date(dateTo)))
    } else if (dateFrom) {
      dateCondition = gte(messages.receivedAt, new Date(dateFrom))
    } else if (dateTo) {
      dateCondition = lte(messages.receivedAt, new Date(dateTo))
    }

    const spCondition = spId ? eq(messages.serviceProviderId, parseInt(spId)) : undefined
    const mnoCondition = mnoId ? eq(messages.routedMnoId, parseInt(mnoId)) : undefined
    
    const conditions = []
    if (dateCondition) conditions.push(dateCondition)
    if (spCondition) conditions.push(spCondition)
    if (mnoCondition) conditions.push(mnoCondition)
    const whereClause = conditions.length > 0 ? and(...conditions) : undefined

    const statusCountsResult = await db
      .select({
        status: messages.finalStatus,
        count: sql<number>`count(*)`,
        totalCost: sql<number>`sum(${messages.cost})`,
      })
      .from(messages)
      .where(whereClause)
      .groupBy(messages.finalStatus)

    const statusCounts: Record<string, { count: number; totalCost: number }> = {}
    let totalMessages = 0
    let totalCost = 0

    for (const row of statusCountsResult) {
      if (row.status) {
        statusCounts[row.status] = {
          count: Number(row.count),
          totalCost: Number(row.totalCost || 0),
        }
        totalMessages += Number(row.count)
        totalCost += Number(row.totalCost || 0)
      }
    }

    const deliveredCount = (statusCounts['delivered']?.count || 0) + (statusCounts['accepted']?.count || 0) + (statusCounts['forwarded']?.count || 0)
    const failedCount = (statusCounts['failed']?.count || 0) + (statusCounts['rejected']?.count || 0) + (statusCounts['undeliverable']?.count || 0)
    const pendingCount = (statusCounts['pending']?.count || 0) + (statusCounts['queued']?.count || 0) + (statusCounts['processing']?.count || 0) + (statusCounts['submitted']?.count || 0) + (statusCounts['received']?.count || 0)

    const spStats = spId ? null : await db
      .select({
        spId: messages.serviceProviderId,
        spName: serviceProviders.name,
        totalMessages: sql<number>`count(*)`,
        totalCost: sql<number>`sum(${messages.cost})`,
      })
      .from(messages)
      .leftJoin(serviceProviders, eq(messages.serviceProviderId, serviceProviders.id))
      .where(whereClause)
      .groupBy(messages.serviceProviderId, serviceProviders.name)
      .orderBy(desc(sql<number>`count(*)`))
      .limit(10)

    const mnoConditions = whereClause ? [whereClause, not(isNull(messages.routedMnoId))] : [not(isNull(messages.routedMnoId))]
    
    const mnoStatsResult = await db
      .select({
        mnoId: messages.routedMnoId,
        mnoName: mnos.name,
        totalMessages: sql<number>`count(*)`,
        delivered: sql<number>`sum(case when ${messages.finalStatus} in ('delivered', 'accepted', 'forwarded') then 1 else 0 end)`,
        failed: sql<number>`sum(case when ${messages.finalStatus} in ('failed', 'rejected', 'undeliverable') then 1 else 0 end)`,
      })
      .from(messages)
      .leftJoin(mnos, eq(messages.routedMnoId, mnos.id))
      .where(and(...mnoConditions))
      .groupBy(messages.routedMnoId, mnos.name)
      .orderBy(desc(sql<number>`count(*)`))
      .limit(10)

    const mnoStats = mnoId ? null : mnoStatsResult.map(s => ({
      id: s.mnoId || 0,
      name: s.mnoName || `MNO ${s.mnoId}`,
      totalMessages: Number(s.totalMessages),
      delivered: Number(s.delivered || 0),
      failed: Number(s.failed || 0),
      deliveryRate: s.totalMessages > 0 ? ((Number(s.delivered || 0) / Number(s.totalMessages)) * 100).toFixed(2) : '0.00',
    }))

    const dailyStats = await db
      .select({
        date: sql<string>`date(${messages.receivedAt})`,
        received: sql<number>`count(*)`,
        delivered: sql<number>`sum(case when ${messages.finalStatus} in ('delivered', 'accepted', 'forwarded') then 1 else 0 end)`,
        failed: sql<number>`sum(case when ${messages.finalStatus} in ('failed', 'rejected', 'undeliverable') then 1 else 0 end)`,
      })
      .from(messages)
      .where(whereClause)
      .groupBy(sql`date(${messages.receivedAt})`)
      .orderBy(sql`date(${messages.receivedAt})`)
      .limit(30)

    return NextResponse.json({
      overview: {
        totalMessages,
        delivered: deliveredCount,
        deliveredRate: totalMessages > 0 ? ((deliveredCount / totalMessages) * 100).toFixed(2) : '0.00',
        failed: failedCount,
        failedRate: totalMessages > 0 ? ((failedCount / totalMessages) * 100).toFixed(2) : '0.00',
        pending: pendingCount,
        totalCost: totalCost.toFixed(2),
      },
      byStatus: statusCounts,
      byDay: dailyStats,
      byServiceProvider: spStats?.map(s => ({
        id: s.spId || 0,
        name: s.spName || 'Unknown',
        totalMessages: Number(s.totalMessages),
        totalCost: Number(s.totalCost || 0),
      })) || [],
      byMno: mnoStats || [],
    })
  } catch (error) {
    console.error('[v0] Error fetching message analytics:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
