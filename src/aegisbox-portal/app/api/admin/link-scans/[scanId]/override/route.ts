import { NextResponse } from 'next/server'
import { sql } from 'drizzle-orm'
import { db } from '@/lib/db'
import { VERDICTS, actorName } from '@/lib/link-scans'

// Sets or clears an admin's verdict on a scan result. Body: { verdict: 'safe' | 'suspicious' |
// 'malicious' | null, note?: string }. The override applies wherever the result is used,
// including messages held for a verdict.
export async function POST(request: Request, { params }: { params: Promise<{ scanId: string }> }) {
  const { scanId } = await params
  if (!/^\d+$/.test(scanId)) return NextResponse.json({ error: 'Invalid scan ID' }, { status: 400 })
  const body = await request.json().catch(() => null)
  const verdict = body?.verdict ?? null
  if (verdict !== null && !VERDICTS.includes(verdict)) {
    return NextResponse.json({ error: 'Invalid verdict' }, { status: 400 })
  }
  const note = typeof body?.note === 'string' ? body.note.trim().slice(0, 500) || null : null
  const by = await actorName()
  try {
    const res = await db.execute(sql`
      UPDATE link_scans
      SET override_verdict = ${verdict}, override_note = ${verdict === null ? null : note},
          override_by = ${verdict === null ? null : by}, override_at = ${verdict === null ? null : sql`NOW()`}
      WHERE id = ${Number(scanId)}
      RETURNING id`)
    if (res.rows.length === 0) return NextResponse.json({ error: 'Scan not found' }, { status: 404 })
    return NextResponse.json({ id: Number(scanId), override_verdict: verdict })
  } catch (error) {
    console.error('Error saving verdict override:', error)
    return NextResponse.json({ error: 'Failed to save' }, { status: 500 })
  }
}
