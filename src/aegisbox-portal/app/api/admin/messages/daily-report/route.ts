import { db } from '@/lib/db'
import { messages, serviceProviders, mnos } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, and, gte, lt, sql, or } from 'drizzle-orm'
import { REPORT_TIME_ZONE_SQL } from '@/lib/time-sql'
import { isReportDate, reportDayRange } from '@/lib/time'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    const startDate = searchParams.get('start_date')
    const endDate = searchParams.get('end_date')
    const spId = searchParams.get('sp_id')
    const status = searchParams.get('status')

    if (!startDate || !endDate) {
      return NextResponse.json({ error: 'Start date and end date are required' }, { status: 400 })
    }
    if (!isReportDate(startDate) || !isReportDate(endDate)) {
      return NextResponse.json({ error: 'Dates must be YYYY-MM-DD' }, { status: 400 })
    }

    const spCondition = spId ? eq(messages.serviceProviderId, parseInt(spId)) : undefined
    
    let statusCondition = undefined
    if (status) {
      const statusList = status.split(',')
      statusCondition = or(
        ...statusList.map(s => eq(messages.finalStatus, s))
      )
    }

    // Whole Lagos days, start_date through end_date.
    const { from, to } = reportDayRange(startDate, endDate)
    const dateCondition = and(gte(messages.submittedAt, from), lt(messages.submittedAt, to))
    const reportDay = sql`DATE(${messages.submittedAt} AT TIME ZONE ${REPORT_TIME_ZONE_SQL})`

    const conditions = [dateCondition]
    if (spCondition) conditions.push(spCondition)
    if (statusCondition) conditions.push(statusCondition)

    const whereClause = and(...conditions)

    const result = await db
      .select({
        reportDate: sql<string>`${reportDay}`,
        serviceProviderId: serviceProviders.id,
        serviceProviderName: serviceProviders.name,
        serviceProviderEmail: serviceProviders.email,
        mnoId: sql<number>`COALESCE(${mnos.id}, 0)`,
        mnoName: sql<string>`COALESCE(${mnos.name}, 'N/A')`,
        currencyCode: messages.currencyCode,
        totalReceived: sql<number>`COUNT(${messages.id})`,
        totalSegmentsReceived: sql<number>`COALESCE(SUM(${messages.totalSegments}), 0)`,
        totalSegmentsSubmitted: sql<number>`COALESCE(SUM(CASE WHEN ${messages.processingStatus} NOT IN ('received', 'failed_validation', 'failed_routing', 'failed_pricing') THEN ${messages.totalSegments} ELSE 0 END), 0)`,
        totalSegmentsDelivered: sql<number>`COALESCE(SUM(CASE WHEN ${messages.finalStatus} = 'delivered' THEN ${messages.totalSegments} ELSE 0 END), 0)`,
        totalSegmentsFailed: sql<number>`COALESCE(SUM(CASE WHEN ${messages.finalStatus} = 'failed' THEN ${messages.totalSegments} ELSE 0 END), 0)`,
        totalSegmentsRejected: sql<number>`COALESCE(SUM(CASE WHEN ${messages.finalStatus} = 'rejected' THEN ${messages.totalSegments} ELSE 0 END), 0)`,
        netCost: sql<number>`COALESCE(SUM(${messages.cost}), 0.0)`,
      })
      .from(messages)
      .innerJoin(serviceProviders, eq(messages.serviceProviderId, serviceProviders.id))
      .leftJoin(mnos, eq(messages.routedMnoId, mnos.id))
      .where(whereClause)
      .groupBy(reportDay, serviceProviders.id, mnos.id, mnos.name, messages.currencyCode)
      .orderBy(desc(reportDay), desc(serviceProviders.name))

    return NextResponse.json({ data: result })
  } catch (error) {
    console.error('[v0] Error fetching daily report:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
