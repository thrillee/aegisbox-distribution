import Link from 'next/link'
import { sql, type SQL } from 'drizzle-orm'
import { db } from '@/lib/db'
import { isReportDate, reportDate, reportDateTime, reportDayRange, reportDayStart } from '@/lib/time'
import { VERDICTS, VERDICT_STYLES } from '@/lib/link-scans'
import { HoldButtons } from './actions'
import { ReloadButton } from '@/components/dashboard/reload-button'

// Link scanning results (step 5): one row per message with links, with the worst verdict of its
// links. Simulator traffic of test customers, and real traffic of customers with a scan mode.

export const dynamic = 'force-dynamic'

const PAGE_SIZE = 50
const WORST = ['', 'safe', 'suspicious', 'malicious']

interface MessageRow {
  message_id: number
  links: number
  pending: number
  worst: number | null
  queued_at: Date
  sender: string
  recipient: string
  text: string
  customer: string
  is_test: boolean
  hold_status: string | null
}

interface ReviewRow {
  message_id: number
  customer: string
  sender: string
  recipient: string
  text: string
  worst_verdict: string | null
  decision: string | null
  created_at: Date
}

function VerdictBadge({ verdict }: { verdict: string }) {
  return (
    <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${VERDICT_STYLES[verdict] ?? VERDICT_STYLES.none}`}>
      {verdict}
    </span>
  )
}

function age(seconds: number) {
  if (seconds < 90) return `${seconds}s`
  if (seconds < 5400) return `${Math.round(seconds / 60)} min`
  return `${Math.round(seconds / 3600)} h`
}

export default async function LinkScansPage({
  searchParams,
}: Readonly<{ searchParams: Promise<{ sp?: string; sender?: string; verdict?: string; traffic?: string; from?: string; to?: string; page?: string }> }>) {
  const params = await searchParams
  const spId = params.sp && /^\d+$/.test(params.sp) ? Number(params.sp) : null
  const sender = (params.sender ?? '').trim().slice(0, 21)
  const verdict = VERDICTS.find((v) => v === params.verdict) ?? (params.verdict === 'pending' ? 'pending' : null)
  const traffic = params.traffic === 'test' || params.traffic === 'real' ? params.traffic : null
  const from = params.from && isReportDate(params.from) ? params.from : reportDate(reportDayStart(6))
  const to = params.to && isReportDate(params.to) ? params.to : reportDate()
  const page = params.page && /^\d+$/.test(params.page) ? Math.max(1, Number(params.page)) : 1
  const range = reportDayRange(from, to)

  const jobConds: SQL[] = [sql`j.created_at >= ${range.from}`, sql`j.created_at < ${range.to}`]
  if (spId) jobConds.push(sql`j.service_provider_id = ${spId}`)
  const outerConds: SQL[] = [sql`TRUE`]
  if (sender) outerConds.push(sql`m.original_source_addr ILIKE ${'%' + sender.replace(/[%_\\]/g, '\\$&') + '%'}`)
  if (verdict === 'pending') outerConds.push(sql`msg.pending > 0`)
  else if (verdict) outerConds.push(sql`msg.worst = ${WORST.indexOf(verdict)}`)
  if (traffic) outerConds.push(sql`sp.is_test = ${traffic === 'test'}`)

  const perMessage = sql`
    WITH msg AS (
      SELECT j.message_id,
             COUNT(*) FILTER (WHERE j.scan_mode <> 'none')::INT AS links,
             COUNT(*) FILTER (WHERE j.status IN ('queued', 'processing'))::INT AS pending,
             MAX(CASE COALESCE(s.override_verdict, s.verdict)
                   WHEN 'malicious' THEN 3 WHEN 'suspicious' THEN 2 WHEN 'safe' THEN 1 END) AS worst,
             MIN(j.created_at) AS queued_at
      FROM link_scan_jobs j
      LEFT JOIN link_scans s ON s.id = j.link_scan_id
      WHERE ${sql.join(jobConds, sql` AND `)}
      GROUP BY j.message_id
    )`
  const outerWhere = sql.join(outerConds, sql` AND `)

  const [rowsRes, countRes, summaryRes, queueRes, topRes, reviewRes, customersRes] = await Promise.all([
    db.execute(sql`${perMessage}
      SELECT msg.*, m.original_source_addr AS sender, m.original_destination_addr AS recipient,
             LEFT(m.short_message, 160) AS text, sp.name AS customer, sp.is_test, h.status AS hold_status
      FROM msg
      JOIN messages m ON m.id = msg.message_id
      JOIN service_providers sp ON sp.id = m.service_provider_id
      LEFT JOIN message_link_holds h ON h.message_id = msg.message_id
      WHERE ${outerWhere}
      ORDER BY msg.queued_at DESC
      LIMIT ${PAGE_SIZE} OFFSET ${(page - 1) * PAGE_SIZE}`),
    db.execute(sql`${perMessage}
      SELECT COUNT(*)::INT AS total FROM msg
      JOIN messages m ON m.id = msg.message_id
      JOIN service_providers sp ON sp.id = m.service_provider_id
      WHERE ${outerWhere}`),
    db.execute(sql`${perMessage}
      SELECT COUNT(*)::INT AS messages,
             COUNT(*) FILTER (WHERE worst = 3)::INT AS malicious,
             COUNT(*) FILTER (WHERE worst = 2)::INT AS suspicious,
             COUNT(*) FILTER (WHERE worst = 1)::INT AS safe,
             COUNT(*) FILTER (WHERE pending > 0)::INT AS pending
      FROM msg`),
    // Queue health, whatever the filters.
    db.execute(sql`
      SELECT COUNT(*) FILTER (WHERE status IN ('queued', 'processing'))::INT AS waiting,
             COALESCE(EXTRACT(EPOCH FROM NOW() - MIN(created_at) FILTER (WHERE status IN ('queued', 'processing'))), 0)::INT AS oldest_seconds,
             COUNT(*) FILTER (WHERE status = 'failed' AND completed_at > NOW() - INTERVAL '24 hours')::INT AS failed_24h,
             (SELECT COUNT(*) FROM link_scans WHERE scanned_at > NOW() - INTERVAL '1 hour' AND scan_kind = 'visit')::INT AS visits_1h,
             (SELECT MAX(scanned_at) FROM link_scans WHERE scan_kind = 'visit') AS last_visit,
             (SELECT COUNT(*) FROM message_link_holds WHERE status = 'waiting')::INT AS holds_waiting
      FROM link_scan_jobs`),
    db.execute(sql`
      SELECT COALESCE(s.final_host, j.host) AS host,
             COUNT(DISTINCT j.message_id)::INT AS messages,
             MAX(CASE COALESCE(s.override_verdict, s.verdict) WHEN 'malicious' THEN 3 ELSE 2 END) AS worst
      FROM link_scan_jobs j
      JOIN link_scans s ON s.id = j.link_scan_id
      WHERE ${sql.join(jobConds, sql` AND `)}
        AND COALESCE(s.override_verdict, s.verdict) IN ('malicious', 'suspicious')
      GROUP BY 1
      ORDER BY messages DESC, host
      LIMIT 10`),
    db.execute(sql`
      SELECT h.message_id, sp.name AS customer, m.original_source_addr AS sender,
             m.original_destination_addr AS recipient, LEFT(m.short_message, 160) AS text,
             h.worst_verdict, h.decision, h.created_at
      FROM message_link_holds h
      JOIN messages m ON m.id = h.message_id
      JOIN service_providers sp ON sp.id = h.service_provider_id
      WHERE h.status = 'review'
      ORDER BY h.created_at
      LIMIT 100`),
    db.execute(sql`SELECT id, name FROM service_providers ORDER BY name`),
  ])

  const rows = rowsRes.rows as unknown as MessageRow[]
  const total = Number(countRes.rows[0]?.total ?? 0)
  const summary = summaryRes.rows[0] as unknown as { messages: number; malicious: number; suspicious: number; safe: number; pending: number }
  const queue = queueRes.rows[0] as unknown as { waiting: number; oldest_seconds: number; failed_24h: number; visits_1h: number; last_visit: Date | null; holds_waiting: number }
  const top = topRes.rows as unknown as { host: string; messages: number; worst: number }[]
  const reviews = reviewRes.rows as unknown as ReviewRow[]
  const customers = customersRes.rows as unknown as { id: number; name: string }[]
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const warnings: string[] = []
  if (queue.waiting > 0 && queue.oldest_seconds > 15 * 60) {
    warnings.push(`${queue.waiting} links are waiting, the oldest for ${age(queue.oldest_seconds)}. Is the linkscanner service running and LINKSCAN_ENABLED on?`)
  } else if (queue.waiting > 500) {
    warnings.push(`${queue.waiting} links are waiting to be scanned. The queue is growing faster than the scanner keeps up.`)
  }
  if (queue.failed_24h > 0) {
    warnings.push(`${queue.failed_24h} links could not be scanned in the last 24 hours (see Failed in the queue).`)
  }

  const href = (p: number, overrides: Record<string, string | null> = {}) => {
    const q = new URLSearchParams({ from, to, page: String(p) })
    if (spId) q.set('sp', String(spId))
    if (sender) q.set('sender', sender)
    if (verdict) q.set('verdict', verdict)
    if (traffic) q.set('traffic', traffic)
    for (const [k, v] of Object.entries(overrides)) {
      if (v === null) q.delete(k)
      else q.set(k, v)
    }
    return `/admin/link-scans?${q}`
  }
  const input = 'px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white text-sm'

  return (
    <div className="p-4 md:p-8 space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Link Scans</h1>
          <p className="text-zinc-400 mt-1">
            Links in messages, opened in a locked-down browser and judged safe, suspicious or malicious. Times are Lagos time.
            Opening a message's details is logged.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <ReloadButton />
          <Link href="/admin/link-scans/lists" className="text-sm font-medium text-zinc-300 hover:text-white border border-white/10 rounded-lg px-3 py-2">
            Allow and block lists
          </Link>
        </div>
      </div>

      {warnings.length > 0 && (
        <div className="rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 space-y-1">
          {warnings.map((w) => <p key={w} className="text-sm text-amber-200">{w}</p>)}
        </div>
      )}

      {reviews.length > 0 && (
        <div className="rounded-2xl border border-red-500/30 bg-red-500/5 p-6">
          <h2 className="font-semibold text-white">Held for review ({reviews.length})</h2>
          <p className="text-sm text-zinc-400 mb-4">
            Real messages stopped by the link check for customers in hold mode. Send or reject each one; messages not decided within 24 hours are rejected.
          </p>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-zinc-400 border-b border-white/10">
                  <th className="py-2 pr-4 font-medium">Message</th>
                  <th className="py-2 pr-4 font-medium">Customer</th>
                  <th className="py-2 pr-4 font-medium">Sender</th>
                  <th className="py-2 pr-4 font-medium">Why</th>
                  <th className="py-2 pr-4 font-medium">Held since</th>
                  <th className="py-2 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {reviews.map((r) => (
                  <tr key={r.message_id} className="border-b border-white/5 align-top">
                    <td className="py-2 pr-4">
                      <Link href={`/admin/link-scans/${r.message_id}`} className="text-white underline-offset-2 hover:underline">#{r.message_id}</Link>
                      <p className="text-xs text-zinc-500 max-w-xs truncate">{r.text}</p>
                    </td>
                    <td className="py-2 pr-4 text-zinc-300">{r.customer}</td>
                    <td className="py-2 pr-4 font-mono text-zinc-300">{r.sender}</td>
                    <td className="py-2 pr-4">{r.worst_verdict ? <VerdictBadge verdict={r.worst_verdict} /> : <span className="text-zinc-400">{r.decision === 'timeout' ? 'no verdict in time' : r.decision}</span>}</td>
                    <td className="py-2 pr-4 text-zinc-400 whitespace-nowrap">{reportDateTime(new Date(r.created_at))}</td>
                    <td className="py-2"><HoldButtons messageId={r.message_id} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <div className="grid gap-4 grid-cols-2 lg:grid-cols-5">
        {[
          { label: 'Messages with links', value: summary.messages, cls: 'text-white', v: null },
          { label: 'Malicious', value: summary.malicious, cls: 'text-red-400', v: 'malicious' },
          { label: 'Suspicious', value: summary.suspicious, cls: 'text-amber-300', v: 'suspicious' },
          { label: 'Safe', value: summary.safe, cls: 'text-emerald-400', v: 'safe' },
          { label: 'Waiting for a verdict', value: summary.pending, cls: 'text-zinc-300', v: 'pending' },
        ].map((c) => (
          <Link key={c.label} href={href(1, { verdict: c.v })} className="rounded-2xl border border-white/10 bg-zinc-900/50 p-4 hover:border-white/20">
            <p className="text-xs text-zinc-400">{c.label}</p>
            <p className={`text-2xl font-bold mt-1 ${c.cls}`}>{Number(c.value).toLocaleString()}</p>
          </Link>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
        <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
          <h2 className="font-semibold text-white mb-3">Scanner</h2>
          <dl className="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm">
            <div><dt className="text-zinc-500">Links waiting</dt><dd className="text-white text-lg">{queue.waiting}</dd></div>
            <div><dt className="text-zinc-500">Oldest waiting</dt><dd className="text-white text-lg">{queue.waiting ? age(queue.oldest_seconds) : '-'}</dd></div>
            <div><dt className="text-zinc-500">Pages opened, last hour</dt><dd className="text-white text-lg">{queue.visits_1h}</dd></div>
            <div><dt className="text-zinc-500">Failed, last 24 h</dt><dd className="text-white text-lg">{queue.failed_24h}</dd></div>
            <div><dt className="text-zinc-500">Real messages waiting</dt><dd className="text-white text-lg">{queue.holds_waiting}</dd></div>
            <div><dt className="text-zinc-500">Last page opened</dt><dd className="text-white">{queue.last_visit ? reportDateTime(new Date(queue.last_visit)) : 'never'}</dd></div>
          </dl>
        </div>
        <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
          <h2 className="font-semibold text-white mb-3">Top bad domains</h2>
          {top.length === 0 ? (
            <p className="text-sm text-zinc-500">None in this period.</p>
          ) : (
            <ul className="space-y-1 text-sm">
              {top.map((t) => (
                <li key={t.host} className="flex items-center justify-between gap-2">
                  <span className="font-mono text-zinc-200 truncate">{t.host}</span>
                  <span className="flex items-center gap-2">
                    <VerdictBadge verdict={t.worst === 3 ? 'malicious' : 'suspicious'} />
                    <span className="text-zinc-400 w-10 text-right">{t.messages}</span>
                  </span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      <form className="flex flex-wrap gap-3 items-end rounded-2xl border border-white/10 bg-zinc-900/50 p-4" method="get">
        <label className="text-sm text-zinc-400 space-y-1">
          <span className="block">Customer</span>
          <select name="sp" defaultValue={spId ?? ''} className={input}>
            <option value="">All</option>
            {customers.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
        </label>
        <label className="text-sm text-zinc-400 space-y-1">
          <span className="block">Sender</span>
          <input name="sender" defaultValue={sender} className={input} placeholder="Any" />
        </label>
        <label className="text-sm text-zinc-400 space-y-1">
          <span className="block">Verdict</span>
          <select name="verdict" defaultValue={verdict ?? ''} className={input}>
            <option value="">Any</option>
            <option value="malicious">Malicious</option>
            <option value="suspicious">Suspicious</option>
            <option value="safe">Safe</option>
            <option value="pending">Waiting</option>
          </select>
        </label>
        <label className="text-sm text-zinc-400 space-y-1">
          <span className="block">Traffic</span>
          <select name="traffic" defaultValue={traffic ?? ''} className={input}>
            <option value="">All</option>
            <option value="test">Test (simulator)</option>
            <option value="real">Real customers</option>
          </select>
        </label>
        <label className="text-sm text-zinc-400 space-y-1">
          <span className="block">From</span>
          <input type="date" name="from" defaultValue={from} className={input} />
        </label>
        <label className="text-sm text-zinc-400 space-y-1">
          <span className="block">To</span>
          <input type="date" name="to" defaultValue={to} className={input} />
        </label>
        <button type="submit" className="px-4 py-2 rounded-lg bg-white text-black text-sm font-medium">Filter</button>
      </form>

      <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-semibold text-white">Messages</h2>
          <span className="text-sm text-zinc-500">{total.toLocaleString()} message{total === 1 ? '' : 's'}</span>
        </div>
        {rows.length === 0 ? (
          <p className="text-sm text-zinc-500 py-6 text-center">No messages with links in this period.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-zinc-400 border-b border-white/10">
                  <th className="py-2 pr-4 font-medium">Time</th>
                  <th className="py-2 pr-4 font-medium">Customer</th>
                  <th className="py-2 pr-4 font-medium">Sender</th>
                  <th className="py-2 pr-4 font-medium">Recipient</th>
                  <th className="py-2 pr-4 font-medium">Message</th>
                  <th className="py-2 pr-4 font-medium text-right">Links</th>
                  <th className="py-2 font-medium">Verdict</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((r) => {
                  const worst = r.pending > 0 && !r.worst ? 'pending' : WORST[Number(r.worst ?? 0)] || (r.links === 0 ? 'none' : 'failed')
                  return (
                    <tr key={r.message_id} className="border-b border-white/5 hover:bg-white/5 align-top">
                      <td className="py-2 pr-4 text-zinc-400 whitespace-nowrap">
                        <Link href={`/admin/link-scans/${r.message_id}`} className="hover:text-white">{reportDateTime(new Date(r.queued_at))}</Link>
                      </td>
                      <td className="py-2 pr-4 text-zinc-300">
                        {r.customer}
                        {r.is_test && <span className="ml-1 text-xs text-zinc-500">(test)</span>}
                      </td>
                      <td className="py-2 pr-4 font-mono text-zinc-300">{r.sender}</td>
                      <td className="py-2 pr-4 font-mono text-zinc-400">{r.recipient}</td>
                      <td className="py-2 pr-4 text-zinc-300 max-w-md">
                        <Link href={`/admin/link-scans/${r.message_id}`} className="hover:text-white line-clamp-2 break-words">{r.text}</Link>
                      </td>
                      <td className="py-2 pr-4 text-right text-zinc-300">{r.links}</td>
                      <td className="py-2 whitespace-nowrap">
                        <VerdictBadge verdict={worst} />
                        {r.hold_status && <span className="ml-1 text-xs text-zinc-500">{r.hold_status}</span>}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
        {pages > 1 && (
          <div className="flex items-center justify-between mt-4 text-sm">
            {page > 1 ? <Link href={href(page - 1)} className="text-zinc-300 hover:text-white">Previous</Link> : <span />}
            <span className="text-zinc-500">Page {page} of {pages}</span>
            {page < pages ? <Link href={href(page + 1)} className="text-zinc-300 hover:text-white">Next</Link> : <span />}
          </div>
        )}
      </div>
    </div>
  )
}
