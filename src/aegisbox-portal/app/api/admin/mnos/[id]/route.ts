import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'
import { db } from '@/lib/db'
import { mnos } from '@/lib/schema'

// Sets an MNO's flags. Body: { is_simulator: boolean }.
export async function PATCH(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
  const body = await request.json().catch(() => null)
  if (typeof body?.is_simulator !== 'boolean') {
    return NextResponse.json({ error: 'is_simulator must be true or false' }, { status: 400 })
  }
  try {
    const [mno] = await db.update(mnos)
      .set({ isSimulator: body.is_simulator, updatedAt: new Date() })
      .where(eq(mnos.id, Number(id)))
      .returning({ id: mnos.id, isSimulator: mnos.isSimulator })
    if (!mno) return NextResponse.json({ error: 'MNO not found' }, { status: 404 })
    return NextResponse.json({ mno })
  } catch (error) {
    console.error('Error updating MNO flags:', error)
    return NextResponse.json({ error: 'Failed to update MNO' }, { status: 500 })
  }
}
