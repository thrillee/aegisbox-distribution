import { db } from '@/lib/db'
import { messages, messageSegments, serviceProviders, wallets, walletTransactions, msisdnPrefixEntries, msisdnPrefixGroups, routingAssignments, mnos } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, and, gte, lte, sql, like, inArray } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const spId = parseInt(resolvedParams.id)
    const { searchParams } = new URL(request.url)
    const days = parseInt(searchParams.get('days') || '30')

    if (isNaN(spId)) {
      return NextResponse.json({ error: 'Invalid service provider ID' }, { status: 400 })
    }

    const startDate = new Date()
    startDate.setDate(startDate.getDate() - days)

    const overviewResult = await db
      .select({
        totalMessages: sql<number>`count(*)`,
        totalSegments: sql<number>`sum(${messages.totalSegments})`,
        totalCost: sql<number>`coalesce(sum(${messages.cost}), 0)`,
        delivered: sql<number>`sum(case when ${messages.finalStatus} = 'delivered' then 1 else 0 end)`,
        failed: sql<number>`sum(case when ${messages.finalStatus} = 'failed' then 1 else 0 end)`,
        pending: sql<number>`sum(case when ${messages.finalStatus} in ('pending', 'queued', 'processing', 'received', 'submitted') then 1 else 0 end)`,
        rejected: sql<number>`sum(case when ${messages.finalStatus} = 'rejected' then 1 else 0 end)`,
      })
      .from(messages)
      .where(eq(messages.serviceProviderId, spId))

    const overview = overviewResult[0] || {}
    const totalMessages = Number(overview.totalMessages || 0)
    const deliveredCount = Number(overview.delivered || 0)
    const failedCount = Number(overview.failed || 0)

    const dailyStats = await db
      .select({
        date: sql<string>`date(${messages.submittedAt})`,
        total: sql<number>`count(*)`,
        delivered: sql<number>`sum(case when ${messages.finalStatus} = 'delivered' then 1 else 0 end)`,
        failed: sql<number>`sum(case when ${messages.finalStatus} = 'failed' then 1 else 0 end)`,
        cost: sql<number>`coalesce(sum(${messages.cost}), 0)`,
      })
      .from(messages)
      .where(and(
        eq(messages.serviceProviderId, spId),
        gte(messages.submittedAt, startDate)
      ))
      .groupBy(sql`date(${messages.submittedAt})`)
      .orderBy(sql`date(${messages.submittedAt})`)
      .limit(30)

    const topDestinationsRaw = await db
      .select({
        prefix: sql<string>`left(${messages.originalDestinationAddr}, 6)`,
        count: sql<number>`count(*)`,
        delivered: sql<number>`sum(case when ${messages.finalStatus} = 'delivered' then 1 else 0 end)`,
        failed: sql<number>`sum(case when ${messages.finalStatus} = 'failed' then 1 else 0 end)`,
      })
      .from(messages)
      .where(eq(messages.serviceProviderId, spId))
      .groupBy(sql`left(${messages.originalDestinationAddr}, 6)`)
      .orderBy(desc(sql`count(*)`))
      .limit(20)

    const prefixes = topDestinationsRaw.map(d => d.prefix)
    
    const prefixMnos = await db
      .select({
        prefix: sql<string>`left(${msisdnPrefixEntries.msisdnPrefix}, 6)`,
        mnoId: routingAssignments.mnoId,
        mnoName: mnos.name,
        networkName: msisdnPrefixGroups.name,
        mnoCountryCode: mnos.countryCode,
      })
      .from(msisdnPrefixEntries)
      .innerJoin(msisdnPrefixGroups, eq(msisdnPrefixEntries.msisdnPrefixGroupId, msisdnPrefixGroups.id))
      .innerJoin(routingAssignments, eq(msisdnPrefixGroups.id, routingAssignments.msisdnPrefixGroupId))
      .leftJoin(mnos, eq(routingAssignments.mnoId, mnos.id))
      .where(inArray(sql`left(${msisdnPrefixEntries.msisdnPrefix}, 6)`, prefixes))

    const prefixToMno = new Map<string, { mnoId: number; mnoName: string; networkName: string; mnoCountryCode: string }>()
    prefixMnos.forEach(row => {
      if (row.prefix) {
        prefixToMno.set(row.prefix, {
          mnoId: row.mnoId || 0,
          mnoName: row.mnoName || 'Unknown',
          networkName: row.networkName || 'Unknown',
          mnoCountryCode: row.mnoCountryCode || '',
        })
      }
    })

    const topDestinations = topDestinationsRaw.map(d => {
      const mnoInfo = prefixToMno.get(d.prefix || '') || { mnoId: 0, mnoName: 'Unknown', networkName: 'Unknown', mnoCountryCode: '' }
      return {
        prefix: d.prefix,
        count: Number(d.count),
        delivered: Number(d.delivered),
        failed: Number(d.failed),
        deliveryRate: Number(d.count) > 0 ? ((Number(d.delivered) / Number(d.count)) * 100).toFixed(2) : '0.00',
        mnoId: mnoInfo.mnoId,
        mnoName: mnoInfo.mnoName,
        networkName: mnoInfo.networkName,
        mnoCountryCode: mnoInfo.mnoCountryCode,
      }
    })

    const statusBreakdown = await db
      .select({
        status: messages.finalStatus,
        count: sql<number>`count(*)`,
        percentage: sql<number>`round((count(*)::decimal / nullif((select count(*) from messages where service_provider_id = ${spId}), 0)) * 100, 2)`,
      })
      .from(messages)
      .where(eq(messages.serviceProviderId, spId))
      .groupBy(messages.finalStatus)
      .orderBy(desc(sql`count(*)`))

    const walletStats = await db
      .select({
        balance: sql<number>`coalesce(sum(case when transaction_type = 'credit' then amount else 0 end) - sum(case when transaction_type in ('debit', 'reversal') then amount else 0 end), 0)`,
        totalCredits: sql<number>`coalesce(sum(case when transaction_type = 'credit' then amount else 0 end), 0)`,
        totalDebits: sql<number>`coalesce(sum(case when transaction_type in ('debit', 'reversal') then amount else 0 end), 0)`,
      })
      .from(walletTransactions)
      .innerJoin(wallets, eq(walletTransactions.walletId, wallets.id))
      .where(eq(wallets.serviceProviderId, spId))

    const hourlyDistribution = await db
      .select({
        hour: sql<number>`extract(hour from ${messages.submittedAt})`,
        count: sql<number>`count(*)`,
      })
      .from(messages)
      .where(and(
        eq(messages.serviceProviderId, spId),
        gte(messages.submittedAt, startDate)
      ))
      .groupBy(sql`extract(hour from ${messages.submittedAt})`)
      .orderBy(sql`extract(hour from ${messages.submittedAt})`)

    const costByDay = await db
      .select({
        date: sql<string>`date(${messages.submittedAt})`,
        cost: sql<number>`coalesce(sum(${messages.cost}), 0)`,
      })
      .from(messages)
      .where(and(
        eq(messages.serviceProviderId, spId),
        gte(messages.submittedAt, startDate)
      ))
      .groupBy(sql`date(${messages.submittedAt})`)
      .orderBy(sql`date(${messages.submittedAt})`)

    const providerInfo = await db
      .select({
        name: serviceProviders.name,
        email: serviceProviders.email,
        status: serviceProviders.status,
        currency: serviceProviders.defaultCurrencyCode,
        createdAt: serviceProviders.createdAt,
      })
      .from(serviceProviders)
      .where(eq(serviceProviders.id, spId))

    const walletBalance = await db
      .select({
        balance: wallets.balance,
      })
      .from(wallets)
      .where(eq(wallets.serviceProviderId, spId))

    return NextResponse.json({
      overview: {
        totalMessages,
        totalSegments: Number(overview.totalSegments || 0),
        totalCost: Number(overview.totalCost || 0),
        delivered: deliveredCount,
        failed: failedCount,
        pending: Number(overview.pending || 0),
        rejected: Number(overview.rejected || 0),
        deliveryRate: totalMessages > 0 ? ((deliveredCount / totalMessages) * 100).toFixed(2) : '0.00',
        failureRate: totalMessages > 0 ? ((failedCount / totalMessages) * 100).toFixed(2) : '0.00',
      },
      provider: providerInfo[0] || null,
      wallet: {
        balance: Number(walletBalance[0]?.balance || 0),
        totalCredits: Number(walletStats[0]?.totalCredits || 0),
        totalDebits: Number(walletStats[0]?.totalDebits || 0),
      },
      dailyStats: dailyStats.map(d => ({
        ...d,
        deliveryRate: d.total > 0 ? ((Number(d.delivered) / d.total) * 100).toFixed(2) : '0.00',
      })),
      topDestinations,
      statusBreakdown: statusBreakdown.map(s => ({
        status: s.status,
        count: Number(s.count),
        percentage: Number(s.percentage || 0),
      })),
      hourlyDistribution: hourlyDistribution.map(h => ({
        hour: Number(h.hour),
        count: Number(h.count),
      })),
      costByDay: costByDay.map(c => ({
        date: c.date,
        cost: Number(c.cost),
      })),
      period: {
        days,
        startDate: startDate.toISOString(),
        endDate: new Date().toISOString(),
      },
    })
  } catch (error) {
    console.error('[v0] Error fetching SP analytics:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
