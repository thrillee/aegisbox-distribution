import Link from 'next/link'
import { notFound } from 'next/navigation'
import { sql } from 'drizzle-orm'
import { db } from '@/lib/db'
import { reportDateTime } from '@/lib/time'
import { VERDICT_STYLES, actorName } from '@/lib/link-scans'
import { HoldButtons, OverrideButtons, QuickListButtons } from '../actions'

export const dynamic = 'force-dynamic'

interface JobRow {
  id: number
  url: string
  normalized_url: string
  host: string
  link_kind: string
  scan_mode: string
  mode_reason: string | null
  status: string
  attempts: number
  last_error: string | null
  scan_id: number | null
  scan_kind: string | null
  verdict: string | null
  override_verdict: string | null
  override_note: string | null
  override_by: string | null
  override_at: Date | null
  score: number | null
  reasons: { code: string; points: number; detail: string }[] | null
  final_url: string | null
  final_host: string | null
  http_status: number | null
  title: string | null
  redirects: { url: string; status?: number | null; type: string }[] | null
  forms: { action?: string; method?: string; fields?: { type?: string; name?: string; label?: string; placeholder?: string }[] }[] | null
  downloads: { filename?: string; url?: string; mime?: string }[] | null
  blocked: { host?: string; ip?: string; reason?: string }[] | null
  tls_issuer: string | null
  scan_error: string | null
  duration_ms: number | null
  has_screenshot: boolean
  scanned_at: Date | null
}

const MODE_TEXT: Record<string, string> = {
  visit: 'opened in the browser',
  reputation: 'checked without opening',
  none: 'recorded only',
}
const REASON_TEXT: Record<string, string> = {
  non_web: 'not a web link (WhatsApp, phone, app store)',
  action_link: 'an unsubscribe, confirm or verify link, which a visit could trigger',
  sampled: 'one of many personal links of a campaign; shares the result of a visited one',
  not_clicked: 'outside the click percentage (LINKSCAN_CLICK_PERCENT), so not opened',
}
const HOLD_TEXT: Record<string, string> = {
  waiting: 'Waiting for link verdicts before sending',
  review: 'Held for admin review',
  released: 'Sent after the link check',
  rejected: 'Rejected by the link check (not sent)',
}

function Badge({ verdict }: { verdict: string }) {
  return <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${VERDICT_STYLES[verdict] ?? VERDICT_STYLES.none}`}>{verdict}</span>
}

export default async function LinkScanMessagePage({ params }: Readonly<{ params: Promise<{ messageId: string }> }>) {
  const { messageId } = await params
  if (!/^\d+$/.test(messageId)) notFound()
  const id = Number(messageId)

  const [msgRes, jobsRes, holdRes] = await Promise.all([
    db.execute(sql`
      SELECT m.id, m.original_source_addr AS sender, m.original_destination_addr AS recipient, m.short_message AS text,
             m.processing_status, m.final_status, m.error_code, m.error_description, m.received_at,
             sp.id AS sp_id, sp.name AS customer, sp.is_test, mno.name AS mno
      FROM messages m
      JOIN service_providers sp ON sp.id = m.service_provider_id
      LEFT JOIN mnos mno ON mno.id = m.routed_mno_id
      WHERE m.id = ${id}`),
    db.execute(sql`
      SELECT j.id, j.url, j.normalized_url, j.host, j.link_kind, j.scan_mode, j.mode_reason, j.status, j.attempts, j.last_error,
             s.id AS scan_id, s.scan_kind, s.verdict, s.override_verdict, s.override_note, s.override_by, s.override_at,
             s.score, s.reasons, s.final_url, s.final_host, s.http_status, s.title, s.redirects, s.forms, s.downloads,
             s.blocked, s.tls_issuer, s.scan_error, s.duration_ms, (s.screenshot_path IS NOT NULL) AS has_screenshot, s.scanned_at
      FROM link_scan_jobs j
      LEFT JOIN link_scans s ON s.id = j.link_scan_id
      WHERE j.message_id = ${id}
      ORDER BY j.id`),
    db.execute(sql`SELECT * FROM message_link_holds WHERE message_id = ${id}`),
  ])
  const msg = msgRes.rows[0] as Record<string, unknown> | undefined
  if (!msg) notFound()
  const jobs = jobsRes.rows as unknown as JobRow[]
  const hold = holdRes.rows[0] as Record<string, unknown> | undefined

  // Message text, numbers and screenshots are personal data: record who looked.
  await db.execute(sql`INSERT INTO link_scan_view_log (viewer, message_id) VALUES (${await actorName()}, ${id})`)

  const customer = { id: Number(msg.sp_id), name: String(msg.customer) }

  return (
    <div className="p-4 md:p-8 space-y-6">
      <div>
        <Link href="/admin/link-scans" className="text-sm text-zinc-400 hover:text-white">← Link Scans</Link>
        <h1 className="text-3xl font-bold text-white mt-2">Message #{id}</h1>
        <p className="text-zinc-400 mt-1">
          {String(msg.customer)}{msg.is_test ? ' (test customer)' : ''} · from <span className="font-mono">{String(msg.sender)}</span> to{' '}
          <span className="font-mono">{String(msg.recipient)}</span> · {reportDateTime(new Date(msg.received_at as string))}
          {msg.mno ? ` · ${String(msg.mno)}` : ''}
        </p>
      </div>

      <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6 space-y-3">
        <h2 className="font-semibold text-white">Text</h2>
        <p className="whitespace-pre-wrap break-words text-zinc-200 text-sm">{String(msg.text)}</p>
        <p className="text-xs text-zinc-500">
          Status: {String(msg.processing_status)} / {String(msg.final_status)}
          {msg.error_code ? ` · ${String(msg.error_code)}: ${String(msg.error_description ?? '')}` : ''} ·{' '}
          <Link href={`/admin/messages/${id}`} className="underline-offset-2 hover:underline">Open in Messages</Link>
        </p>
      </div>

      {hold && (
        <div className={`rounded-2xl border p-6 space-y-2 ${hold.status === 'review' ? 'border-red-500/30 bg-red-500/5' : 'border-white/10 bg-zinc-900/50'}`}>
          <h2 className="font-semibold text-white">Link check before sending ({String(hold.mode)} mode)</h2>
          <p className="text-sm text-zinc-300">
            {HOLD_TEXT[String(hold.status)] ?? String(hold.status)}
            {hold.decision ? ` · ${String(hold.decision).replace(/_/g, ' ')}` : ''}
            {hold.decided_by ? ` by ${String(hold.decided_by)}` : ''}
            {hold.decided_at ? ` at ${reportDateTime(new Date(hold.decided_at as string))}` : ''}
          </p>
          {(hold.status === 'review' || hold.status === 'waiting') && <HoldButtons messageId={id} />}
        </div>
      )}

      {jobs.length === 0 && <p className="text-sm text-zinc-500">No links were found in this message.</p>}

      {jobs.map((j) => {
        const effective = j.override_verdict ?? j.verdict
        const domain = j.final_host || j.host
        return (
          <div key={j.id} className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6 space-y-4">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="font-mono text-sm text-white break-all">{j.url}</p>
                <p className="text-xs text-zinc-500 mt-1">
                  {MODE_TEXT[j.scan_mode] ?? j.scan_mode}
                  {j.mode_reason ? `: ${REASON_TEXT[j.mode_reason] ?? j.mode_reason}` : ''}
                  {j.scan_kind === 'list' ? ' · decided by an allow or block list' : ''}
                </p>
              </div>
              <div className="flex items-center gap-2">
                {effective ? <Badge verdict={effective} /> : <Badge verdict={j.status === 'failed' ? 'failed' : j.status === 'skipped' ? 'none' : 'pending'} />}
                {j.score !== null && <span className="text-xs text-zinc-400">score {j.score}</span>}
              </div>
            </div>

            {j.override_verdict && (
              <p className="text-xs text-amber-200">
                Corrected from {j.verdict} to {j.override_verdict}
                {j.override_by ? ` by ${j.override_by}` : ''}{j.override_at ? ` on ${reportDateTime(new Date(j.override_at))}` : ''}
                {j.override_note ? `: ${j.override_note}` : ''}
              </p>
            )}

            {j.status === 'failed' && <p className="text-sm text-zinc-400">Could not be scanned after {j.attempts} attempts: {j.last_error}</p>}
            {(j.status === 'queued' || j.status === 'processing') && <p className="text-sm text-zinc-400">Waiting to be scanned{j.last_error ? ` (last try: ${j.last_error})` : ''}.</p>}

            {j.scan_id && (
              <div className="grid gap-6 lg:grid-cols-[3fr_2fr]">
                <div className="space-y-4 text-sm min-w-0">
                  {j.reasons && j.reasons.length > 0 && (
                    <div>
                      <h3 className="text-zinc-400 text-xs uppercase mb-1">Why</h3>
                      <ul className="space-y-1">
                        {j.reasons.map((r, i) => (
                          <li key={i} className="flex gap-2">
                            <span className="text-zinc-500 w-10 text-right shrink-0">+{r.points}</span>
                            <span className="text-zinc-200">{r.detail}</span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                  {j.final_url && (
                    <div>
                      <h3 className="text-zinc-400 text-xs uppercase mb-1">Final page</h3>
                      <p className="font-mono text-zinc-200 break-all">{j.final_url}</p>
                      <p className="text-zinc-500 text-xs">
                        {j.http_status ? `HTTP ${j.http_status}` : ''}{j.title ? ` · ${j.title}` : ''}{j.tls_issuer ? ` · certificate from ${j.tls_issuer}` : ''}
                        {j.duration_ms ? ` · ${(j.duration_ms / 1000).toFixed(1)} s` : ''}
                      </p>
                    </div>
                  )}
                  {j.redirects && j.redirects.length > 1 && (
                    <div>
                      <h3 className="text-zinc-400 text-xs uppercase mb-1">Redirects</h3>
                      <ol className="space-y-1">
                        {j.redirects.map((r, i) => (
                          <li key={i} className="font-mono text-xs text-zinc-300 break-all">
                            <span className="text-zinc-500">{i + 1}. {r.type}{r.status ? ` ${r.status}` : ''}</span> {r.url}
                          </li>
                        ))}
                      </ol>
                    </div>
                  )}
                  {j.forms && j.forms.length > 0 && (
                    <div>
                      <h3 className="text-zinc-400 text-xs uppercase mb-1">Forms</h3>
                      {j.forms.map((f, i) => (
                        <p key={i} className="text-xs text-zinc-300 break-all">
                          {f.method?.toUpperCase() ?? 'GET'} {f.action || '(same page)'}: {(f.fields ?? []).map((fd) => fd.label || fd.name || fd.placeholder || fd.type).filter(Boolean).join(', ')}
                        </p>
                      ))}
                    </div>
                  )}
                  {j.downloads && j.downloads.length > 0 && (
                    <div>
                      <h3 className="text-zinc-400 text-xs uppercase mb-1">Downloads (blocked, not saved)</h3>
                      {j.downloads.map((d, i) => <p key={i} className="font-mono text-xs text-zinc-300 break-all">{d.filename} {d.mime ? `(${d.mime})` : ''}</p>)}
                    </div>
                  )}
                  {j.blocked && j.blocked.length > 0 && (
                    <div>
                      <h3 className="text-zinc-400 text-xs uppercase mb-1">Blocked requests</h3>
                      {j.blocked.map((b, i) => <p key={i} className="font-mono text-xs text-zinc-300 break-all">{b.host} {b.ip ? `(${b.ip})` : ''}: {b.reason}</p>)}
                    </div>
                  )}
                  {j.scan_error && <p className="text-xs text-zinc-500">Error: {j.scan_error}</p>}
                  <p className="text-xs text-zinc-500">Scanned {j.scanned_at ? reportDateTime(new Date(j.scanned_at)) : ''} (result {j.scan_id})</p>
                  <OverrideButtons scanId={j.scan_id} current={j.override_verdict} />
                  {j.link_kind === 'web' && domain && <QuickListButtons domain={domain} customer={customer} />}
                </div>
                <div>
                  {j.has_screenshot ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={`/api/admin/link-scans/${j.scan_id}/screenshot`} alt={`Screenshot of ${j.final_url ?? j.url}`} className="rounded-lg border border-white/10 w-full" />
                  ) : (
                    <p className="text-xs text-zinc-500">No screenshot{j.scan_kind === 'visit' ? ' (deleted after 7 days, or the page did not load)' : ' (the link was not opened)'}.</p>
                  )}
                </div>
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
