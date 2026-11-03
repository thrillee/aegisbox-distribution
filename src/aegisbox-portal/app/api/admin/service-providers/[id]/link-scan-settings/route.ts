import { NextResponse } from 'next/server'
import { sql } from 'drizzle-orm'
import { db } from '@/lib/db'
import { SCAN_MODES, actorName } from '@/lib/link-scans'

const DEFAULTS = { mode: 'off', hold_seconds: 60, fail_open: true, block_suspicious: false, charge_blocked: false }

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
  const res = await db.execute(sql`
    SELECT mode, hold_seconds, fail_open, block_suspicious, charge_blocked, updated_by, updated_at
    FROM link_scan_settings WHERE service_provider_id = ${Number(id)}`)
  return NextResponse.json(res.rows[0] ?? { ...DEFAULTS, updated_by: null, updated_at: null })
}

// Body: { mode, hold_seconds, fail_open, block_suspicious, charge_blocked }. Takes effect within
// 30 seconds (the gateway caches settings).
export async function PUT(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
  const body = await request.json().catch(() => null)
  const mode = body?.mode
  if (!SCAN_MODES.includes(mode)) return NextResponse.json({ error: 'Invalid mode' }, { status: 400 })
  const hold = Number(body?.hold_seconds ?? 60)
  if (!Number.isInteger(hold) || hold < 5 || hold > 600) {
    return NextResponse.json({ error: 'Wait time must be between 5 and 600 seconds' }, { status: 400 })
  }
  const flag = (v: unknown, d: boolean) => (typeof v === 'boolean' ? v : d)
  const failOpen = flag(body?.fail_open, true)
  const blockSuspicious = flag(body?.block_suspicious, false)
  const chargeBlocked = flag(body?.charge_blocked, false)
  const by = await actorName()
  try {
    const sp = await db.execute(sql`SELECT id FROM service_providers WHERE id = ${Number(id)}`)
    if (sp.rows.length === 0) return NextResponse.json({ error: 'Customer not found' }, { status: 404 })
    const res = await db.execute(sql`
      INSERT INTO link_scan_settings (service_provider_id, mode, hold_seconds, fail_open, block_suspicious, charge_blocked, updated_by, updated_at)
      VALUES (${Number(id)}, ${mode}, ${hold}, ${failOpen}, ${blockSuspicious}, ${chargeBlocked}, ${by}, NOW())
      ON CONFLICT (service_provider_id) DO UPDATE
      SET mode = EXCLUDED.mode, hold_seconds = EXCLUDED.hold_seconds, fail_open = EXCLUDED.fail_open,
          block_suspicious = EXCLUDED.block_suspicious, charge_blocked = EXCLUDED.charge_blocked,
          updated_by = EXCLUDED.updated_by, updated_at = NOW()
      RETURNING mode, hold_seconds, fail_open, block_suspicious, charge_blocked, updated_by, updated_at`)
    return NextResponse.json(res.rows[0])
  } catch (error) {
    console.error('Error saving link scan settings:', error)
    return NextResponse.json({ error: 'Failed to save' }, { status: 500 })
  }
}
