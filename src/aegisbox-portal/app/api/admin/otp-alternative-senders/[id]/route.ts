import { db } from '@/lib/db'
import { otpAlternativeSenders } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const senderId = parseInt(id)

    const result = await db.select({
      id: otpAlternativeSenders.id,
      senderIdString: otpAlternativeSenders.senderIdString,
      serviceProviderId: otpAlternativeSenders.serviceProviderId,
      mnoId: otpAlternativeSenders.mnoId,
      status: otpAlternativeSenders.status,
      currentUsageCount: otpAlternativeSenders.currentUsageCount,
      maxUsageCount: otpAlternativeSenders.maxUsageCount,
      resetIntervalHours: otpAlternativeSenders.resetIntervalHours,
      lastResetAt: otpAlternativeSenders.lastResetAt,
      lastUsedAt: otpAlternativeSenders.lastUsedAt,
      notes: otpAlternativeSenders.notes,
      createdAt: otpAlternativeSenders.createdAt,
      updatedAt: otpAlternativeSenders.updatedAt,
    })
      .from(otpAlternativeSenders)
      .where(eq(otpAlternativeSenders.id, senderId))

    if (result.length === 0) {
      return NextResponse.json({ error: 'OTP alternative sender not found' }, { status: 404 })
    }

    return NextResponse.json({ sender: result[0] })
  } catch (error) {
    console.error('[v0] Error fetching OTP alternative sender:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function PUT(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const senderId = parseInt(id)
    const body = await request.json()
    const { senderIdString, serviceProviderId, mnoId, status, maxUsageCount, resetIntervalHours, notes } = body

    const result = await db.update(otpAlternativeSenders)
      .set({
        senderIdString: senderIdString || undefined,
        serviceProviderId: serviceProviderId ?? null,
        mnoId: mnoId ?? null,
        status: status || 'active',
        maxUsageCount: maxUsageCount ?? 10000,
        resetIntervalHours: resetIntervalHours ?? null,
        notes: notes ?? null,
        updatedAt: new Date(),
      })
      .where(eq(otpAlternativeSenders.id, senderId))
      .returning()

    if (result.length === 0) {
      return NextResponse.json({ error: 'OTP alternative sender not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, sender: result[0] })
  } catch (error: any) {
    console.error('[v0] Error updating OTP alternative sender:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'This sender ID string already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to update OTP alternative sender' }, { status: 500 })
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const senderId = parseInt(id)

    const result = await db.delete(otpAlternativeSenders)
      .where(eq(otpAlternativeSenders.id, senderId))
      .returning()

    if (result.length === 0) {
      return NextResponse.json({ error: 'OTP alternative sender not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('[v0] Error deleting OTP alternative sender:', error)
    return NextResponse.json({ error: 'Failed to delete OTP alternative sender' }, { status: 500 })
  }
}
