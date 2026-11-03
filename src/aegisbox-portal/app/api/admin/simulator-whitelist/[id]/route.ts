import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'
import { db } from '@/lib/db'
import { simulatorWhitelistNumbers } from '@/lib/schema'

export async function DELETE(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
  try {
    const [row] = await db.delete(simulatorWhitelistNumbers)
      .where(eq(simulatorWhitelistNumbers.id, Number(id)))
      .returning({ id: simulatorWhitelistNumbers.id })
    if (!row) return NextResponse.json({ error: 'Not found' }, { status: 404 })
    return NextResponse.json({ deleted: row.id })
  } catch (error) {
    console.error('Error removing whitelisted number:', error)
    return NextResponse.json({ error: 'Failed to remove the number' }, { status: 500 })
  }
}
