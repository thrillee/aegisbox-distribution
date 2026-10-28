import { db } from '@/lib/db'
import { templates } from '@/lib/schema'
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
      id: templates.id,
      name: templates.name,
      content: templates.content,
      status: templates.status,
      serviceProviderId: templates.serviceProviderId,
      createdAt: templates.createdAt,
      updatedAt: templates.updatedAt,
    })
      .from(templates)
      .where(eq(templates.id, templateId))

    if (result.length === 0) {
      return NextResponse.json({ error: 'Template not found' }, { status: 404 })
    }

    return NextResponse.json({ template: result[0] })
  } catch (error) {
    console.error('[v0] Error fetching template:', error)
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
    const { name, content, status } = body

    if (!name || !content) {
      return NextResponse.json({ error: 'Name and content are required' }, { status: 400 })
    }

    const result = await db.update(templates)
      .set({
        name,
        content,
        status: status || 'active',
        updatedAt: new Date(),
      })
      .where(eq(templates.id, templateId))
      .returning()

    if (result.length === 0) {
      return NextResponse.json({ error: 'Template not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, template: result[0] })
  } catch (error: any) {
    console.error('[v0] Error updating template:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'A template with this name already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to update template' }, { status: 500 })
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params
    const templateId = parseInt(id)

    const result = await db.delete(templates)
      .where(eq(templates.id, templateId))
      .returning()

    if (result.length === 0) {
      return NextResponse.json({ error: 'Template not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('[v0] Error deleting template:', error)
    return NextResponse.json({ error: 'Failed to delete template' }, { status: 500 })
  }
}
