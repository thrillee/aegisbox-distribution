import { db } from '@/lib/db'
import { senderIds } from '@/lib/schema'
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
      id: senderIds.id,
      senderIdString: senderIds.senderIdString,
      serviceProviderId: senderIds.serviceProviderId,
      mnoConnectionId: senderIds.mnoConnectionId,
      status: senderIds.status,
      providerId: senderIds.providerId,
      rateLimit: senderIds.rateLimit,
      otpTemplate: senderIds.otpTemplate,
      networks: senderIds.networks,
      maxUsageCount: senderIds.maxUsageCount,
      currentUsageCount: senderIds.currentUsageCount,
      resetIntervalHours: senderIds.resetIntervalHours,
      createdAt: senderIds.createdAt,
      updatedAt: senderIds.updatedAt,
    })
      .from(senderIds)
      .where(eq(senderIds.id, senderId))

    if (result.length === 0) {
      return NextResponse.json({ error: 'Sender ID not found' }, { status: 404 })
    }

    return NextResponse.json({ sender: result[0] })
  } catch (error) {
    console.error('[v0] Error fetching sender ID:', error)
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
    const { senderIdString, serviceProviderId, mnoConnectionId, status, rateLimit, otpTemplate, networks, maxUsageCount, resetIntervalHours } = body

    const result = await db.update(senderIds)
      .set({
        senderIdString: senderIdString || undefined,
        serviceProviderId: serviceProviderId ?? null,
        mnoConnectionId: mnoConnectionId ?? null,
        status: status || 'pending',
        rateLimit: rateLimit ?? 100,
        otpTemplate: otpTemplate ?? null,
        networks: networks ?? null,
        maxUsageCount: maxUsageCount ?? 10000,
        resetIntervalHours: resetIntervalHours ?? null,
        updatedAt: new Date(),
      })
      .where(eq(senderIds.id, senderId))
      .returning()

    if (result.length === 0) {
      return NextResponse.json({ error: 'Sender ID not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, sender: result[0] })
  } catch (error: any) {
    console.error('[v0] Error updating sender ID:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'This sender ID string already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to update sender ID' }, { status: 500 })
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const senderId = parseInt(id)

    const result = await db.delete(senderIds)
      .where(eq(senderIds.id, senderId))
      .returning()

    if (result.length === 0) {
      return NextResponse.json({ error: 'Sender ID not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('[v0] Error deleting sender ID:', error)
    return NextResponse.json({ error: 'Failed to delete sender ID' }, { status: 500 })
  }
}
