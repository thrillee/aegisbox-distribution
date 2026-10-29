import { db } from '@/lib/db'
import { NextResponse } from 'next/server'
import { sql, type SQL } from 'drizzle-orm'
import { isReportDate, reportDayRange } from '@/lib/time'
import { REPORT_TIME_ZONE_SQL } from '@/lib/time-sql'

// The gateway sets final_status to pending, delivered, failed or rejected. Rejected counts as
// failed; anything not final yet counts as pending.
const statusCounts = sql`
  count(*) AS total,
  count(*) FILTER (WHERE m.final_status = 'delivered') AS delivered,
  count(*) FILTER (WHERE m.final_status IN ('failed', 'rejected')) AS failed`

// Cost charged per message, less any refund made for it.
const refunds = sql`
  LEFT JOIN (
    SELECT message_id, SUM(amount) AS refunded
    FROM wallet_transactions
    WHERE transaction_type = 'reversal' AND message_id IS NOT NULL
    GROUP BY message_id
  ) r ON r.message_id = m.id`
const netCost = sql`COALESCE(SUM(m.cost - COALESCE(r.refunded, 0)), 0)`

const num = (v: unknown) => Number(v ?? 0) || 0
const rate = (part: number, total: number) => (total > 0 ? ((part / total) * 100).toFixed(2) : '0.00')
const intParam = (v: string | null) => (v && /^\d+$/.test(v) ? Number(v) : null)

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    const spId = intParam(searchParams.get('sp_id'))
    const mnoId = intParam(searchParams.get('mno_id'))
    const dateFrom = searchParams.get('date_from') || ''
    const dateTo = searchParams.get('date_to') || ''
    if ((dateFrom && !isReportDate(dateFrom)) || (dateTo && !isReportDate(dateTo))) {
      return NextResponse.json({ error: 'Dates must be YYYY-MM-DD' }, { status: 400 })
    }

    // Whole Lagos days, both ends included, by when the customer submitted the message.
    const conditions: SQL[] = []
    if (dateFrom) conditions.push(sql`m.submitted_at >= ${reportDayRange(dateFrom, dateFrom).from}`)
    if (dateTo) conditions.push(sql`m.submitted_at < ${reportDayRange(dateTo, dateTo).to}`)
    if (spId) conditions.push(sql`m.service_provider_id = ${spId}`)
    if (mnoId) conditions.push(sql`m.routed_mno_id = ${mnoId}`)
    const where = conditions.length > 0 ? sql`WHERE ${sql.join(conditions, sql` AND `)}` : sql``
    const day = sql`to_char(m.submitted_at AT TIME ZONE ${REPORT_TIME_ZONE_SQL}, 'YYYY-MM-DD')`

    const [overviewRes, chargedRes, dailyRes, spRes, mnoRes] = await Promise.all([
      db.execute(sql`SELECT ${statusCounts} FROM messages m ${where}`),
      db.execute(sql`
        SELECT m.currency_code AS currency, ${netCost} AS charged
        FROM messages m ${refunds} ${where}
        GROUP BY m.currency_code ORDER BY m.currency_code`),
      // Latest 60 days in the range, oldest first for the chart.
      db.execute(sql`
        SELECT * FROM (
          SELECT ${day} AS date, ${statusCounts}
          FROM messages m ${where}
          GROUP BY 1 ORDER BY 1 DESC LIMIT 60
        ) d ORDER BY date`),
      db.execute(sql`
        SELECT m.service_provider_id AS id, COALESCE(sp.name, 'Unknown') AS name, m.currency_code AS currency,
               ${statusCounts}, ${netCost} AS charged
        FROM messages m
        LEFT JOIN service_providers sp ON sp.id = m.service_provider_id
        ${refunds} ${where}
        GROUP BY m.service_provider_id, sp.name, m.currency_code
        ORDER BY total DESC`),
      db.execute(sql`
        SELECT m.routed_mno_id AS id, COALESCE(mn.name, 'Not routed') AS name, ${statusCounts}
        FROM messages m
        LEFT JOIN mnos mn ON mn.id = m.routed_mno_id
        ${where}
        GROUP BY m.routed_mno_id, mn.name
        ORDER BY total DESC`),
    ])

    const o = overviewRes.rows[0] ?? {}
    const total = num(o.total)
    const delivered = num(o.delivered)
    const failed = num(o.failed)

    return NextResponse.json({
      overview: {
        totalMessages: total,
        delivered,
        deliveredRate: rate(delivered, total),
        failed,
        failedRate: rate(failed, total),
        pending: total - delivered - failed,
      },
      charged: chargedRes.rows.map((r) => ({ currency: String(r.currency), amount: num(r.charged) })),
      byDay: dailyRes.rows.map((r) => ({
        date: String(r.date),
        received: num(r.total),
        delivered: num(r.delivered),
        failed: num(r.failed),
      })),
      byServiceProvider: spRes.rows.map((r) => ({
        id: num(r.id),
        name: String(r.name),
        currency: String(r.currency),
        totalMessages: num(r.total),
        delivered: num(r.delivered),
        failed: num(r.failed),
        charged: num(r.charged),
      })),
      byMno: mnoRes.rows.map((r) => ({
        id: num(r.id),
        name: String(r.name),
        totalMessages: num(r.total),
        delivered: num(r.delivered),
        failed: num(r.failed),
        deliveryRate: rate(num(r.delivered), num(r.total)),
      })),
    })
  } catch (error) {
    console.error('Error fetching message analytics:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
