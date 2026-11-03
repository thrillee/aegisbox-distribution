import { NextResponse } from 'next/server'
import { sql } from 'drizzle-orm'
import { db } from '@/lib/db'
import { actorName, isValidListDomain, normalizeListDomain } from '@/lib/link-scans'

// Adds or replaces an allow or block entry. Body: { domain, list_type: 'allow' | 'block',
// service_provider_id?: number | null, note?: string }. One entry per domain and scope: adding
// the other list type for the same domain replaces it.
export async function POST(request: Request) {
  const body = await request.json().catch(() => null)
  const domain = normalizeListDomain(String(body?.domain ?? ''))
  if (!isValidListDomain(domain)) {
    return NextResponse.json({ error: 'Enter a domain such as example.com' }, { status: 400 })
  }
  const listType = body?.list_type
  if (listType !== 'allow' && listType !== 'block') {
    return NextResponse.json({ error: 'Choose allow or block' }, { status: 400 })
  }
  const spRaw = body?.service_provider_id
  const spId = spRaw === null || spRaw === undefined || spRaw === '' ? null : Number(spRaw)
  if (spId !== null && !Number.isInteger(spId)) return NextResponse.json({ error: 'Invalid customer' }, { status: 400 })
  const note = typeof body?.note === 'string' ? body.note.trim().slice(0, 300) || null : null
  const by = await actorName()
  try {
    if (spId !== null) {
      const sp = await db.execute(sql`SELECT id FROM service_providers WHERE id = ${spId}`)
      if (sp.rows.length === 0) return NextResponse.json({ error: 'Customer not found' }, { status: 404 })
    }
    const res = await db.execute(sql`
      INSERT INTO link_domain_lists (domain, list_type, service_provider_id, note, created_by)
      VALUES (${domain}, ${listType}, ${spId}, ${note}, ${by})
      ON CONFLICT (domain, COALESCE(service_provider_id, 0))
      DO UPDATE SET list_type = EXCLUDED.list_type, note = EXCLUDED.note, created_by = EXCLUDED.created_by, updated_at = NOW()
      RETURNING id, domain, list_type`)
    return NextResponse.json({ entry: res.rows[0] }, { status: 201 })
  } catch (error) {
    console.error('Error saving list entry:', error)
    return NextResponse.json({ error: 'Failed to save' }, { status: 500 })
  }
}
