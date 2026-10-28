import { db } from '@/lib/db'
import { otpMessageTemplates } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const templateId = parseInt(id)

    const result = await db.select({
      id: otpMessageTemplates.id,
      name: otpMessageTemplates.name,
      contentTemplate: otpMessageTemplates.contentTemplate,
      defaultBrandName: otpMessageTemplates.defaultBrandName,
      status: otpMessageTemplates.status,
      createdAt: otpMessageTemplates.createdAt,
      updatedAt: otpMessageTemplates.updatedAt,
    })
      .from(otpMessageTemplates)
      .where(eq(otpMessageTemplates.id, templateId))

    if (result.length === 0) {
      return NextResponse.json({ error: 'OTP message template not found' }, { status: 404 })
    }

    return NextResponse.json({ template: result[0] })
  } catch (error) {
    console.error('[v0] Error fetching OTP message template:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function PUT(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const templateId = parseInt(id)
    const body = await request.json()
    const { name, contentTemplate, defaultBrandName, status } = body

    if (!name || !contentTemplate) {
      return NextResponse.json({ error: 'Name and content template are required' }, { status: 400 })
    }

    const result = await db.update(otpMessageTemplates)
      .set({
        name,
        contentTemplate,
        defaultBrandName: defaultBrandName || null,
        status: status || 'active',
        updatedAt: new Date(),
      })
      .where(eq(otpMessageTemplates.id, templateId))
      .returning()

    if (result.length === 0) {
      return NextResponse.json({ error: 'OTP message template not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, template: result[0] })
  } catch (error: any) {
    console.error('[v0] Error updating OTP message template:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'An OTP message template with this name already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to update OTP message template' }, { status: 500 })
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const templateId = parseInt(id)

    const result = await db.delete(otpMessageTemplates)
      .where(eq(otpMessageTemplates.id, templateId))
      .returning()

    if (result.length === 0) {
      return NextResponse.json({ error: 'OTP message template not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('[v0] Error deleting OTP message template:', error)
    return NextResponse.json({ error: 'Failed to delete OTP message template' }, { status: 500 })
  }
}
