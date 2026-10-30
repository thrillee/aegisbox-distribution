import { sql, type SQL } from 'drizzle-orm'
import { db } from '@/lib/db'
import { reportDayRange } from '@/lib/time'

// Messages still waiting for a final DLR, by the same rules as the gateway's manual DLR
// (ApplyManualDLRAndEnqueue): a segment counts when it was sent, has no final DLR, and either
// the MNO accepted it (has an MNO message ID) or, for "no answer", no submit failure was recorded
// (error code empty or the DLR timeout marker TIMEOUT, and no error description). Messages with
// any failed segment, failed or refunded messages are left out.

export interface PendingDlrFilter {
  serviceProviderId: number
  from: string // YYYY-MM-DD, Lagos day
  to: string
  olderThanMinutes: number
}

const FINAL = sql`('DELIVRD','UNDELIV','EXPIRED','REJECTD','DELETED')`
const accepted = sql`ms.mno_message_id IS NOT NULL`
// Never answered: no error recorded (TIMEOUT only marks "no DLR in time") or the submit timed out
// (MNO_TIMEOUT). Same rule as the gateway's include_unconfirmed; rejections are never included.
const noAnswer = sql`(ms.mno_message_id IS NULL AND ms.sent_to_mno_at IS NOT NULL AND (((ms.error_code IS NULL OR ms.error_code = 'TIMEOUT') AND ms.error_description IS NULL) OR ms.error_code = 'MNO_TIMEOUT'))`
const awaiting = sql`(ms.dlr_status IS NULL OR upper(ms.dlr_status) NOT IN ${FINAL})`

// For a message list: messages a manual DLR can complete (waiting, accepted or no answer, no
// failed or unsent part, not closed or refunded). messageId is the outer query's messages.id.
export function waitingForDlr(messageId: SQL, finalStatus: SQL) {
  return sql`(
    ${finalStatus} NOT IN ('delivered', 'failed', 'rejected', 'expired')
    AND EXISTS (SELECT 1 FROM message_segments ms WHERE ms.message_id = ${messageId} AND ${awaiting} AND (${accepted} OR ${noAnswer}))
    AND NOT EXISTS (SELECT 1 FROM message_segments ms WHERE ms.message_id = ${messageId} AND ${awaiting} AND NOT ${accepted} AND NOT ${noAnswer})
    AND NOT EXISTS (SELECT 1 FROM wallet_transactions wt WHERE wt.message_id = ${messageId} AND wt.transaction_type = 'reversal')
  )`
}

// Per message: whether it has segments of each kind still waiting.
function classified(f: PendingDlrFilter) {
  const { from, to } = reportDayRange(f.from, f.to)
  return sql`
    SELECT m.id,
           bool_or(${awaiting} AND ${accepted}) AS has_accepted,
           bool_or(${awaiting} AND ${noAnswer}) AS has_no_answer,
           bool_or(${awaiting} AND NOT ${accepted} AND NOT ${noAnswer}) AS has_failed,
           bool_and(NOT ${awaiting} OR coalesce(ms.sent_to_mno_at < now() - make_interval(mins => ${f.olderThanMinutes}), false)) AS old_enough
    FROM messages m
    JOIN message_segments ms ON ms.message_id = m.id
    WHERE m.service_provider_id = ${f.serviceProviderId}
      AND m.submitted_at >= ${from} AND m.submitted_at < ${to}
      AND m.final_status NOT IN ('delivered', 'failed', 'rejected', 'expired')
      AND NOT EXISTS (SELECT 1 FROM wallet_transactions wt WHERE wt.message_id = m.id AND wt.transaction_type = 'reversal')
    GROUP BY m.id
    HAVING bool_or(${awaiting})`
}

export interface PendingDlrSummary {
  accepted: number // MNO accepted, no DLR yet
  noAnswer: number // no MNO answer recorded, no failure recorded
  failed: number // has a failed segment: never completed here
  tooRecent: number // waiting but sent within the last olderThanMinutes
}

export async function pendingDlrSummary(f: PendingDlrFilter): Promise<PendingDlrSummary> {
  const res = await db.execute(sql`
    SELECT count(*) FILTER (WHERE NOT has_failed AND old_enough AND has_accepted AND NOT has_no_answer) AS accepted,
           count(*) FILTER (WHERE NOT has_failed AND old_enough AND has_no_answer) AS no_answer,
           count(*) FILTER (WHERE has_failed) AS failed,
           count(*) FILTER (WHERE NOT has_failed AND NOT old_enough) AS too_recent
    FROM (${classified(f)}) c`)
  const r = res.rows[0] ?? {}
  return {
    accepted: Number(r.accepted ?? 0),
    noAnswer: Number(r.no_answer ?? 0),
    failed: Number(r.failed ?? 0),
    tooRecent: Number(r.too_recent ?? 0),
  }
}

// Message IDs to send a DLR for. Without includeNoAnswer only messages whose waiting segments
// were all accepted by the MNO are returned.
export async function pendingDlrMessageIds(f: PendingDlrFilter, includeNoAnswer: boolean): Promise<number[]> {
  const res = await db.execute(sql`
    SELECT id FROM (${classified(f)}) c
    WHERE NOT has_failed AND old_enough ${includeNoAnswer ? sql`` : sql`AND NOT has_no_answer`}
    ORDER BY id`)
  return res.rows.map((r) => Number(r.id))
}
