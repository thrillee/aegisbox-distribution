import Link from 'next/link'
import { sql } from 'drizzle-orm'
import { db } from '@/lib/db'
import { REPORT_TIME_ZONE_SQL } from '@/lib/time-sql'
import { reportDate, reportDayStart } from '@/lib/time'
import { getCurrentServiceProviderId, NOT_LINKED_MESSAGE } from '@/lib/service-provider'
import { formatMoney } from '@/lib/money'

const PERIODS = { '7d': 7, '30d': 30, '90d': 90 } as const
type Period = keyof typeof PERIODS

interface Summary {
  total: number
  delivered: number
  failed: number
  pending: number
}

interface DailyRow extends Summary {
  day: string
  cost: number
}

interface GroupRow {
  label: string
  total: number
  delivered: number
  failed: number
}

const num = (v: unknown) => Number(v ?? 0) || 0

// Final status buckets: delivered, failed, and everything not final yet.
const statusCounts = sql`
  count(*) AS total,
  count(*) FILTER (WHERE m.final_status = 'delivered') AS delivered,
  count(*) FILTER (WHERE m.final_status = 'failed') AS failed,
  count(*) FILTER (WHERE m.final_status NOT IN ('delivered', 'failed')) AS pending`

async function getAnalytics(serviceProviderId: number, since: Date) {
  const [summaryRes, dailyRes, senderRes, networkRes, spendRes, currencyRes] = await Promise.all([
    db.execute(sql`
      SELECT ${statusCounts}
      FROM messages m
      WHERE m.service_provider_id = ${serviceProviderId} AND m.submitted_at >= ${since}`),
    db.execute(sql`
      SELECT to_char(m.submitted_at AT TIME ZONE ${REPORT_TIME_ZONE_SQL}, 'YYYY-MM-DD') AS day, ${statusCounts},
             COALESCE(SUM(m.cost - COALESCE(r.refunded, 0)), 0) AS cost
      FROM messages m
      -- Net of refunds, so the days add up to the Spent total.
      LEFT JOIN (
        SELECT message_id, SUM(amount) AS refunded
        FROM wallet_transactions
        WHERE transaction_type = 'reversal' AND message_id IS NOT NULL
        GROUP BY message_id
      ) r ON r.message_id = m.id
      WHERE m.service_provider_id = ${serviceProviderId} AND m.submitted_at >= ${since}
      GROUP BY 1 ORDER BY 1 DESC`),
    db.execute(sql`
      SELECT m.original_source_addr AS label, ${statusCounts}
      FROM messages m
      WHERE m.service_provider_id = ${serviceProviderId} AND m.submitted_at >= ${since}
      GROUP BY 1 ORDER BY total DESC LIMIT 20`),
    // Network from the destination's longest matching prefix, never the upstream route used.
    db.execute(sql`
      SELECT COALESCE(regexp_replace(net.name, '\s*Prefixes$', ''), 'Other') AS label, ${statusCounts}
      FROM messages m
      LEFT JOIN LATERAL (
        SELECT pg.name
        FROM msisdn_prefix_entries e
        JOIN msisdn_prefix_groups pg ON pg.id = e.msisdn_prefix_group_id
        WHERE ltrim(m.original_destination_addr, '+') LIKE e.msisdn_prefix || '%'
        ORDER BY length(e.msisdn_prefix) DESC
        LIMIT 1
      ) net ON true
      WHERE m.service_provider_id = ${serviceProviderId} AND m.submitted_at >= ${since}
      GROUP BY 1 ORDER BY total DESC`),
    // Spend is what the wallet was actually charged: debits less refunds.
    db.execute(sql`
      SELECT w.currency_code AS currency,
             COALESCE(SUM(CASE wt.transaction_type WHEN 'debit' THEN wt.amount
                                                   WHEN 'reversal' THEN -wt.amount ELSE 0 END), 0) AS spent
      FROM wallet_transactions wt
      JOIN wallets w ON w.id = wt.wallet_id
      WHERE w.service_provider_id = ${serviceProviderId} AND wt.transaction_date >= ${since}
      GROUP BY w.currency_code`),
    db.execute(sql`SELECT default_currency_code AS currency FROM service_providers WHERE id = ${serviceProviderId}`),
  ])

  const s = summaryRes.rows[0] ?? {}
  const toGroup = (r: Record<string, unknown>): GroupRow => ({
    label: String(r.label ?? 'Unknown'),
    total: num(r.total),
    delivered: num(r.delivered),
    failed: num(r.failed),
  })

  return {
    summary: { total: num(s.total), delivered: num(s.delivered), failed: num(s.failed), pending: num(s.pending) } as Summary,
    daily: dailyRes.rows.map((r) => ({ ...toGroup(r), day: String(r.day), pending: num(r.pending), cost: num(r.cost) })) as DailyRow[],
    bySender: senderRes.rows.map(toGroup),
    byNetwork: networkRes.rows.map(toGroup),
    spend: spendRes.rows.map((r) => ({ currency: String(r.currency), spent: num(r.spent) })),
    currency: String(currencyRes.rows[0]?.currency ?? 'NGN'),
  }
}

function rate(delivered: number, total: number) {
  return total > 0 ? `${((delivered / total) * 100).toFixed(1)}%` : '-'
}

function SummaryCard({ title, value, detail }: Readonly<{ title: string; value: string; detail?: string }>) {
  return (
    <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-4">
      <p className="text-sm font-medium text-zinc-400 pb-2">{title}</p>
      <div className="text-2xl font-bold text-white">{value}</div>
      {detail && <p className="text-xs text-zinc-500 mt-1">{detail}</p>}
    </div>
  )
}

function GroupTable({ title, rows, labelHeader }: Readonly<{ title: string; rows: GroupRow[]; labelHeader: string }>) {
  return (
    <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
      <h2 className="text-lg font-semibold text-white mb-4">{title}</h2>
      {rows.length === 0 ? (
        <p className="text-sm text-zinc-500">No messages in this period.</p>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-zinc-400 border-b border-white/10">
              <th className="py-2 pr-4 font-medium">{labelHeader}</th>
              <th className="py-2 pr-4 font-medium text-right">Messages</th>
              <th className="py-2 pr-4 font-medium text-right">Delivered</th>
              <th className="py-2 pr-4 font-medium text-right">Failed</th>
              <th className="py-2 font-medium text-right">Delivery rate</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((r) => (
              <tr key={r.label} className="border-b border-white/5">
                <td className="py-2 pr-4 text-white">{r.label}</td>
                <td className="py-2 pr-4 text-right text-zinc-300">{r.total.toLocaleString()}</td>
                <td className="py-2 pr-4 text-right text-emerald-400">{r.delivered.toLocaleString()}</td>
                <td className="py-2 pr-4 text-right text-red-400">{r.failed.toLocaleString()}</td>
                <td className="py-2 text-right text-zinc-300">{rate(r.delivered, r.total)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}

export default async function AnalyticsPage({
  searchParams,
}: Readonly<{ searchParams: Promise<{ period?: string }> }>) {
  const serviceProviderId = await getCurrentServiceProviderId()
  if (!serviceProviderId) {
    return (
      <div className="p-8 text-center">
        <p className="text-zinc-400">{NOT_LINKED_MESSAGE}</p>
      </div>
    )
  }

  const params = await searchParams
  const period: Period = params.period && params.period in PERIODS ? (params.period as Period) : '7d'
  const since = reportDayStart(PERIODS[period] - 1)

  const data = await getAnalytics(serviceProviderId, since)
  const { summary } = data
  const spent = data.spend.length > 0
    ? data.spend.map((s) => formatMoney(s.spent, s.currency)).join(' · ')
    : formatMoney(0, data.currency)

  return (
    <div className="p-8 space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Analytics</h1>
          <p className="text-zinc-400 mt-1">Your messages since {reportDate(since)} (Lagos time)</p>
        </div>
        <div className="flex items-center gap-2 bg-zinc-900/50 border border-white/10 rounded-xl p-1">
          {(Object.keys(PERIODS) as Period[]).map((p) => (
            <Link
              key={p}
              href={`/sp/analytics?period=${p}`}
              className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                period === p ? 'bg-white text-black' : 'text-zinc-400 hover:text-white'
              }`}
            >
              {PERIODS[p]} Days
            </Link>
          ))}
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <SummaryCard title="Messages" value={summary.total.toLocaleString()} detail={`${summary.pending.toLocaleString()} awaiting a final status`} />
        <SummaryCard title="Delivered" value={summary.delivered.toLocaleString()} detail={`Delivery rate ${rate(summary.delivered, summary.total)}`} />
        <SummaryCard title="Failed" value={summary.failed.toLocaleString()} />
        <SummaryCard title="Spent" value={spent} detail="Charged to your wallet, less refunds" />
      </div>

      <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
        <h2 className="text-lg font-semibold text-white mb-4">Daily</h2>
        {data.daily.length === 0 ? (
          <p className="text-sm text-zinc-500">No messages in this period.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-zinc-400 border-b border-white/10">
                <th className="py-2 pr-4 font-medium">Date</th>
                <th className="py-2 pr-4 font-medium text-right">Messages</th>
                <th className="py-2 pr-4 font-medium text-right">Delivered</th>
                <th className="py-2 pr-4 font-medium text-right">Failed</th>
                <th className="py-2 pr-4 font-medium text-right">Pending</th>
                <th className="py-2 pr-4 font-medium text-right">Delivery rate</th>
                <th className="py-2 font-medium text-right">Cost</th>
              </tr>
            </thead>
            <tbody>
              {data.daily.map((d) => (
                <tr key={d.day} className="border-b border-white/5">
                  <td className="py-2 pr-4 text-white">{d.day}</td>
                  <td className="py-2 pr-4 text-right text-zinc-300">{d.total.toLocaleString()}</td>
                  <td className="py-2 pr-4 text-right text-emerald-400">{d.delivered.toLocaleString()}</td>
                  <td className="py-2 pr-4 text-right text-red-400">{d.failed.toLocaleString()}</td>
                  <td className="py-2 pr-4 text-right text-amber-400">{d.pending.toLocaleString()}</td>
                  <td className="py-2 pr-4 text-right text-zinc-300">{rate(d.delivered, d.total)}</td>
                  <td className="py-2 text-right text-zinc-300">{formatMoney(d.cost, data.currency)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <GroupTable title="By network" labelHeader="Network" rows={data.byNetwork} />
        <GroupTable title="By sender ID" labelHeader="Sender ID" rows={data.bySender} />
      </div>
    </div>
  )
}
