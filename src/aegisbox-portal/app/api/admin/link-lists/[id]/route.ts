import { NextResponse } from 'next/server'
import { sql } from 'drizzle-orm'
import { db } from '@/lib/db'

export async function DELETE(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
  try {
    const res = await db.execute(sql`DELETE FROM link_domain_lists WHERE id = ${Number(id)} RETURNING id`)
    if (res.rows.length === 0) return NextResponse.json({ error: 'Not found' }, { status: 404 })
    return NextResponse.json({ deleted: Number(id) })
  } catch (error) {
    console.error('Error removing list entry:', error)
    return NextResponse.json({ error: 'Failed to remove' }, { status: 500 })
  }
}
