import { db } from '@/lib/db'
import { otpMessageTemplates } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, count, like } from 'drizzle-orm'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    const page = parseInt(searchParams.get('page') || '1')
    const limit = parseInt(searchParams.get('limit') || '10')
    const offset = (page - 1) * limit
    const status = searchParams.get('status')
    const search = searchParams.get('search')

    let conditions: any[] = []
    if (status) conditions.push(eq(otpMessageTemplates.status, status as any))
    if (search) {
      conditions.push(like(otpMessageTemplates.name, `%${search}%`))
    }

    const whereClause = conditions.length > 0 ? conditions[0] : undefined

    const [totalResult, templateList] = await Promise.all([
      db.select({ count: count() }).from(otpMessageTemplates).where(whereClause),
      db.select({
        id: otpMessageTemplates.id,
        name: otpMessageTemplates.name,
        contentTemplate: otpMessageTemplates.contentTemplate,
        defaultBrandName: otpMessageTemplates.defaultBrandName,
        status: otpMessageTemplates.status,
        createdAt: otpMessageTemplates.createdAt,
        updatedAt: otpMessageTemplates.updatedAt,
      })
        .from(otpMessageTemplates)
        .where(whereClause)
        .orderBy(desc(otpMessageTemplates.createdAt))
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
    console.error('[v0] Error fetching OTP message templates:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { name, contentTemplate, defaultBrandName, status } = body

    if (!name || !contentTemplate) {
      return NextResponse.json({ error: 'Name and content template are required' }, { status: 400 })
    }

    const result = await db.insert(otpMessageTemplates).values({
      name,
      contentTemplate,
      defaultBrandName: defaultBrandName || null,
      status: status || 'active',
    }).returning()

    return NextResponse.json({ success: true, template: result[0] })
  } catch (error: any) {
    console.error('[v0] Error creating OTP message template:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'An OTP message template with this name already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to create OTP message template' }, { status: 500 })
  }
}
