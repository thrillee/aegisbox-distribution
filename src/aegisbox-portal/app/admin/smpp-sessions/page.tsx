import Link from 'next/link'
import { sql, type SQL } from 'drizzle-orm'
import { db } from '@/lib/db'
import { isReportDate, reportDate, reportDateTime, reportDayRange, reportDayStart } from '@/lib/time'

// SMPP bind history recorded by the gateway (smpp_sessions): one row per bind attempt.

const PAGE_SIZE = 100
const STATUSES = ['connected', 'closed', 'refused'] as const

interface SessionRow {
  id: number
  system_id: string
  customer: string | null
  service_provider_id: number | null
  bind_type: string
  remote_addr: string
  status: string
  started_at: Date
  ended_at: Date | null
  end_reason: string | null
}

const BIND_TYPES: Record<string, string> = { trx: 'Transceiver', tx: 'Transmitter', rx: 'Receiver' }
const STATUS_STYLES: Record<string, string> = {
  connected: 'bg-emerald-500/20 text-emerald-400',
  closed: 'bg-zinc-500/20 text-zinc-300',
  refused: 'bg-red-500/20 text-red-400',
}

function duration(from: Date, to: Date) {
  let secs = Math.max(0, Math.round((to.getTime() - from.getTime()) / 1000))
  const d = Math.floor(secs / 86400)
  secs -= d * 86400
  const h = Math.floor(secs / 3600)
  secs -= h * 3600
  const m = Math.floor(secs / 60)
  const s = secs - m * 60
  if (d > 0) return `${d}d ${h}h`
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

// The client's IP without the port.
const clientIp = (addr: string) => addr.replace(/:\d+$/, '')

export default async function SmppSessionsPage({
  searchParams,
}: Readonly<{ searchParams: Promise<{ sp?: string; status?: string; from?: string; to?: string; page?: string }> }>) {
  const params = await searchParams
  const spId = params.sp && /^\d+$/.test(params.sp) ? Number(params.sp) : null
  const status = STATUSES.find((s) => s === params.status) ?? null
  const from = params.from && isReportDate(params.from) ? params.from : reportDate(reportDayStart(6))
  const to = params.to && isReportDate(params.to) ? params.to : reportDate()
  const page = params.page && /^\d+$/.test(params.page) ? Math.max(1, Number(params.page)) : 1
  const range = reportDayRange(from, to)

  const conditions: SQL[] = [sql`s.started_at >= ${range.from}`, sql`s.started_at < ${range.to}`]
  if (spId) conditions.push(sql`s.service_provider_id = ${spId}`)
  if (status) conditions.push(sql`s.status = ${status}`)
  const where = sql.join(conditions, sql` AND `)

  const [rowsRes, countRes, liveRes, customersRes] = await Promise.all([
    db.execute(sql`
      SELECT s.id, s.system_id, sp.name AS customer, s.service_provider_id, s.bind_type, s.remote_addr,
             s.status, s.started_at, s.ended_at, s.end_reason
      FROM smpp_sessions s
      LEFT JOIN service_providers sp ON sp.id = s.service_provider_id
      WHERE ${where}
      ORDER BY s.started_at DESC, s.id DESC
      LIMIT ${PAGE_SIZE} OFFSET ${(page - 1) * PAGE_SIZE}`),
    db.execute(sql`
      SELECT count(*) AS total,
             count(*) FILTER (WHERE s.status = 'refused') AS refused
      FROM smpp_sessions s WHERE ${where}`),
    // Connected right now, whatever the date filter.
    db.execute(sql`
      SELECT s.system_id, sp.name AS customer, count(*) AS sessions, min(s.started_at) AS since,
             string_agg(DISTINCT s.bind_type, ', ') AS bind_types
      FROM smpp_sessions s
      LEFT JOIN service_providers sp ON sp.id = s.service_provider_id
      WHERE s.status = 'connected' ${spId ? sql`AND s.service_provider_id = ${spId}` : sql``}
      GROUP BY s.system_id, sp.name
      ORDER BY sp.name, s.system_id`),
    db.execute(sql`SELECT id, name FROM service_providers ORDER BY name`),
  ])

  const rows = rowsRes.rows as unknown as SessionRow[]
  const total = Number(countRes.rows[0]?.total ?? 0)
  const refused = Number(countRes.rows[0]?.refused ?? 0)
  const live = liveRes.rows as unknown as { system_id: string; customer: string | null; sessions: number; since: Date; bind_types: string }[]
  const customers = customersRes.rows as unknown as { id: number; name: string }[]
  const now = new Date()
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const pageHref = (p: number) => {
    const q = new URLSearchParams({ from, to, page: String(p) })
    if (spId) q.set('sp', String(spId))
    if (status) q.set('status', status)
    return `/admin/smpp-sessions?${q}`
  }

  const input = 'px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white text-sm'

  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-white">SMPP Sessions</h1>
        <p className="text-zinc-400 mt-1">When customers connect and disconnect, and refused binds. Times are Lagos time.</p>
      </div>

      <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
        <h2 className="font-semibold text-white mb-4">Connected now</h2>
        {live.length === 0 ? (
          <p className="text-sm text-zinc-500">No customer is connected.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-zinc-400 border-b border-white/10">
                  <th className="py-2 pr-4 font-medium">Customer</th>
                  <th className="py-2 pr-4 font-medium">System ID</th>
                  <th className="py-2 pr-4 font-medium text-right">Sessions</th>
                  <th className="py-2 pr-4 font-medium">Bind types</th>
                  <th className="py-2 font-medium">Connected since</th>
                </tr>
              </thead>
              <tbody>
                {live.map((l) => (
                  <tr key={l.system_id} className="border-b border-white/5">
                    <td className="py-2 pr-4 text-white">{l.customer ?? '-'}</td>
                    <td className="py-2 pr-4 font-mono text-white">{l.system_id}</td>
                    <td className="py-2 pr-4 text-right text-emerald-400">{Number(l.sessions)}</td>
                    <td className="py-2 pr-4 text-zinc-300">{l.bind_types}</td>
                    <td className="py-2 text-zinc-300">
                      {reportDateTime(new Date(l.since))}
                      <span className="text-zinc-500"> ({duration(new Date(l.since), now)})</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6 space-y-4">
        <form method="get" className="flex flex-wrap items-end gap-3">
          <label className="text-sm text-zinc-400">
            <span className="block mb-1">Customer</span>
            <select name="sp" defaultValue={spId ?? ''} className={input}>
              <option value="">All customers</option>
              {customers.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </label>
          <label className="text-sm text-zinc-400">
            <span className="block mb-1">Status</span>
            <select name="status" defaultValue={status ?? ''} className={input}>
              <option value="">All</option>
              <option value="connected">Connected</option>
              <option value="closed">Closed</option>
              <option value="refused">Refused</option>
            </select>
          </label>
          <label className="text-sm text-zinc-400">
            <span className="block mb-1">From</span>
            <input type="date" name="from" defaultValue={from} className={input} />
          </label>
          <label className="text-sm text-zinc-400">
            <span className="block mb-1">To</span>
            <input type="date" name="to" defaultValue={to} className={input} />
          </label>
          <button type="submit" className="px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 text-sm">
            Show
          </button>
        </form>

        <p className="text-sm text-zinc-400">
          {total.toLocaleString()} bind attempts, {refused.toLocaleString()} refused
        </p>

        {rows.length === 0 ? (
          <p className="text-sm text-zinc-500">No sessions in this period.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-zinc-400 border-b border-white/10">
                  <th className="py-2 pr-4 font-medium">Connected</th>
                  <th className="py-2 pr-4 font-medium">Disconnected</th>
                  <th className="py-2 pr-4 font-medium">Duration</th>
                  <th className="py-2 pr-4 font-medium">Customer</th>
                  <th className="py-2 pr-4 font-medium">System ID</th>
                  <th className="py-2 pr-4 font-medium">Bind type</th>
                  <th className="py-2 pr-4 font-medium">Client IP</th>
                  <th className="py-2 pr-4 font-medium">Status</th>
                  <th className="py-2 font-medium">Reason</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((r) => {
                  const started = new Date(r.started_at)
                  const ended = r.ended_at ? new Date(r.ended_at) : null
                  return (
                    <tr key={r.id} className="border-b border-white/5 align-top">
                      <td className="py-2 pr-4 text-zinc-300 whitespace-nowrap">{reportDateTime(started)}</td>
                      <td className="py-2 pr-4 text-zinc-300 whitespace-nowrap">
                        {r.status === 'refused' ? '-' : ended ? reportDateTime(ended) : <span className="text-emerald-400">still connected</span>}
                      </td>
                      <td className="py-2 pr-4 text-zinc-300 whitespace-nowrap">
                        {r.status === 'refused' ? '-' : duration(started, ended ?? now)}
                      </td>
                      <td className="py-2 pr-4 text-white">
                        {r.service_provider_id ? (
                          <Link href={`/admin/service-providers/${r.service_provider_id}`} className="hover:underline">
                            {r.customer ?? `#${r.service_provider_id}`}
                          </Link>
                        ) : (
                          <span className="text-zinc-500">unknown</span>
                        )}
                      </td>
                      <td className="py-2 pr-4 font-mono text-white">{r.system_id}</td>
                      <td className="py-2 pr-4 text-zinc-300">{BIND_TYPES[r.bind_type] ?? r.bind_type}</td>
                      <td className="py-2 pr-4 font-mono text-zinc-400">{clientIp(r.remote_addr)}</td>
                      <td className="py-2 pr-4">
                        <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium capitalize ${STATUS_STYLES[r.status] ?? STATUS_STYLES.closed}`}>
                          {r.status}
                        </span>
                      </td>
                      <td className="py-2 text-zinc-400">{r.end_reason ?? ''}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}

        {pages > 1 && (
          <div className="flex items-center gap-3 text-sm">
            {page > 1 && <Link href={pageHref(page - 1)} className="text-white hover:underline">Previous</Link>}
            <span className="text-zinc-400">Page {page} of {pages}</span>
            {page < pages && <Link href={pageHref(page + 1)} className="text-white hover:underline">Next</Link>}
          </div>
        )}
      </div>
    </div>
  )
}
