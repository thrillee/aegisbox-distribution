import { db } from '@/lib/db'
import { otpSenderTemplateAssignments } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const assignmentId = parseInt(id)

    const result = await db.select({
      id: otpSenderTemplateAssignments.id,
      otpAlternativeSenderId: otpSenderTemplateAssignments.otpAlternativeSenderId,
      otpMessageTemplateId: otpSenderTemplateAssignments.otpMessageTemplateId,
      mnoId: otpSenderTemplateAssignments.mnoId,
      priority: otpSenderTemplateAssignments.priority,
      status: otpSenderTemplateAssignments.status,
      assignmentUsageCount: otpSenderTemplateAssignments.assignmentUsageCount,
      createdAt: otpSenderTemplateAssignments.createdAt,
      updatedAt: otpSenderTemplateAssignments.updatedAt,
    })
      .from(otpSenderTemplateAssignments)
      .where(eq(otpSenderTemplateAssignments.id, assignmentId))

    if (result.length === 0) {
      return NextResponse.json({ error: 'OTP sender template assignment not found' }, { status: 404 })
    }

    return NextResponse.json({ assignment: result[0] })
  } catch (error) {
    console.error('[v0] Error fetching OTP sender template assignment:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function PUT(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const assignmentId = parseInt(id)
    const body = await request.json()
    const { otpAlternativeSenderId, otpMessageTemplateId, mnoId, priority, status } = body

    const result = await db.update(otpSenderTemplateAssignments)
      .set({
        otpAlternativeSenderId: otpAlternativeSenderId || undefined,
        otpMessageTemplateId: otpMessageTemplateId || undefined,
        mnoId: mnoId ?? null,
        priority: priority ?? 0,
        status: status || 'active',
        updatedAt: new Date(),
      })
      .where(eq(otpSenderTemplateAssignments.id, assignmentId))
      .returning()

    if (result.length === 0) {
      return NextResponse.json({ error: 'OTP sender template assignment not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, assignment: result[0] })
  } catch (error: any) {
    console.error('[v0] Error updating OTP sender template assignment:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'This assignment already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to update OTP sender template assignment' }, { status: 500 })
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const assignmentId = parseInt(id)

    const result = await db.delete(otpSenderTemplateAssignments)
      .where(eq(otpSenderTemplateAssignments.id, assignmentId))
      .returning()

    if (result.length === 0) {
      return NextResponse.json({ error: 'OTP sender template assignment not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('[v0] Error deleting OTP sender template assignment:', error)
    return NextResponse.json({ error: 'Failed to delete OTP sender template assignment' }, { status: 500 })
  }
}
