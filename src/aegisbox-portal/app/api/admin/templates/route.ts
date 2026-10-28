import { db } from '@/lib/db'
import { templates, statusEnum } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, count, sql, and } from 'drizzle-orm'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    const page = parseInt(searchParams.get('page') || '1')
    const limit = parseInt(searchParams.get('limit') || '10')
    const offset = (page - 1) * limit
    const status = searchParams.get('status')
    const serviceProviderId = searchParams.get('serviceProviderId')
    const search = searchParams.get('search')

    const conditions = []
    if (status) conditions.push(eq(templates.status, status as any))
    if (serviceProviderId) conditions.push(eq(templates.serviceProviderId, parseInt(serviceProviderId)))
    if (search) {
      conditions.push(sql`(${templates.name} ILIKE ${`%${search}%`} OR ${templates.content} ILIKE ${`%${search}%`})`)
    }

    const whereClause = conditions.length > 0 ? and(...conditions) : undefined

    const [totalResult, templateList] = await Promise.all([
      db.select({ count: count() }).from(templates).where(whereClause),
      db.select({
        id: templates.id,
        name: templates.name,
        content: templates.content,
        status: templates.status,
        serviceProviderId: templates.serviceProviderId,
        createdAt: templates.createdAt,
        updatedAt: templates.updatedAt,
      })
        .from(templates)
        .where(whereClause)
        .orderBy(desc(templates.createdAt))
        .limit(limit)
        .offset(offset),
    ])

    const total = totalResult[0]?.count || 0

    return NextResponse.json({
      templates: templateList,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    })
  } catch (error) {
    console.error('[v0] Error fetching templates:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { name, content, serviceProviderId, status } = body

    if (!name || !content) {
      return NextResponse.json({ error: 'Name and content are required' }, { status: 400 })
    }

    const result = await db.insert(templates).values({
      name,
      content,
      serviceProviderId: serviceProviderId || null,
      status: status || 'active',
    }).returning()

    return NextResponse.json({ success: true, template: result[0] })
  } catch (error: any) {
    console.error('[v0] Error creating template:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'A template with this name already exists for this service provider' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to create template' }, { status: 500 })
  }
}
