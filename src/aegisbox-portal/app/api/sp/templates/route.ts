import { NextRequest, NextResponse } from 'next/server'
import { db } from '@/lib/db'
import { templates } from '@/lib/schema'
import { requireServiceProvider } from '@/lib/service-provider'
import { desc, eq, and, like } from 'drizzle-orm'

export async function GET(request: NextRequest) {
  try {
    const sp = await requireServiceProvider()
    if (sp.response) return sp.response

    const { searchParams } = new URL(request.url)
    const search = searchParams.get('search')

    const conditions = [eq(templates.serviceProviderId, sp.serviceProviderId)]
    if (search) {
      conditions.push(like(templates.name, `%${search}%`))
    }

    const templateList = await db
      .select()
      .from(templates)
      .where(and(...conditions))
      .orderBy(desc(templates.createdAt))

    return NextResponse.json({ templates: templateList })
  } catch (error) {
    console.error('Error fetching templates:', error)
    return NextResponse.json({ error: 'Failed to fetch templates' }, { status: 500 })
  }
}

export async function POST(request: NextRequest) {
  try {
    const sp = await requireServiceProvider()
    if (sp.response) return sp.response

    const body = await request.json()
    const { name, content } = body

    if (!name || !content) {
      return NextResponse.json({ error: 'Name and content are required' }, { status: 400 })
    }

    const newTemplate = await db.insert(templates).values({
      serviceProviderId: sp.serviceProviderId,
      name,
      content,
      status: 'active',
    }).returning()

    return NextResponse.json({ success: true, template: newTemplate[0] })
  } catch (error) {
    console.error('Error creating template:', error)
    return NextResponse.json({ error: 'Failed to create template' }, { status: 500 })
  }
}
