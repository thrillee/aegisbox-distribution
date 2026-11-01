import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'
import { db } from '@/lib/db'
import { serviceProviders } from '@/lib/schema'

// Marks a customer as a test customer or not. Body: { is_test: boolean }.
export async function PATCH(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
  const body = await request.json().catch(() => null)
  if (typeof body?.is_test !== 'boolean') {
    return NextResponse.json({ error: 'is_test must be true or false' }, { status: 400 })
  }
  try {
    const [sp] = await db.update(serviceProviders)
      .set({ isTest: body.is_test, updatedAt: new Date() })
      .where(eq(serviceProviders.id, Number(id)))
      .returning({ id: serviceProviders.id, isTest: serviceProviders.isTest })
    if (!sp) return NextResponse.json({ error: 'Service provider not found' }, { status: 404 })
    return NextResponse.json({ serviceProvider: sp })
  } catch (error) {
    console.error('Error updating test flag:', error)
    return NextResponse.json({ error: 'Failed to update the customer' }, { status: 500 })
  }
}
