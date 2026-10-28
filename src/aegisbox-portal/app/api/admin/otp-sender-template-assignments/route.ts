import { db } from '@/lib/db'
import { otpSenderTemplateAssignments, otpAlternativeSenders, otpMessageTemplates, mnos } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, count, and, sql } from 'drizzle-orm'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    const page = parseInt(searchParams.get('page') || '1')
    const limit = parseInt(searchParams.get('limit') || '50')
    const offset = (page - 1) * limit
    const status = searchParams.get('status')
    const otpAlternativeSenderId = searchParams.get('otpAlternativeSenderId')
    const otpMessageTemplateId = searchParams.get('otpMessageTemplateId')
    const mnoId = searchParams.get('mnoId')

    const conditions = []
    if (status) conditions.push(eq(otpSenderTemplateAssignments.status, status as any))
    if (otpAlternativeSenderId) conditions.push(eq(otpSenderTemplateAssignments.otpAlternativeSenderId, parseInt(otpAlternativeSenderId)))
    if (otpMessageTemplateId) conditions.push(eq(otpSenderTemplateAssignments.otpMessageTemplateId, parseInt(otpMessageTemplateId)))
    if (mnoId) conditions.push(eq(otpSenderTemplateAssignments.mnoId, parseInt(mnoId)))

    const whereClause = conditions.length > 0 ? and(...conditions) : undefined

    const [totalResult, assignmentList] = await Promise.all([
      db.select({ count: count() }).from(otpSenderTemplateAssignments).where(whereClause),
      db.select({
        id: otpSenderTemplateAssignments.id,
        otpAlternativeSenderId: otpSenderTemplateAssignments.otpAlternativeSenderId,
        otpMessageTemplateId: otpSenderTemplateAssignments.otpMessageTemplateId,
        mnoId: otpSenderTemplateAssignments.mnoId,
        priority: otpSenderTemplateAssignments.priority,
        status: otpSenderTemplateAssignments.status,
        assignmentUsageCount: otpSenderTemplateAssignments.assignmentUsageCount,
        createdAt: otpSenderTemplateAssignments.createdAt,
        updatedAt: otpSenderTemplateAssignments.updatedAt,
        senderName: otpAlternativeSenders.senderIdString,
        templateName: otpMessageTemplates.name,
        mnoName: mnos.name,
        mnoCountryCode: mnos.countryCode,
      })
        .from(otpSenderTemplateAssignments)
        .leftJoin(otpAlternativeSenders, eq(otpSenderTemplateAssignments.otpAlternativeSenderId, otpAlternativeSenders.id))
        .leftJoin(otpMessageTemplates, eq(otpSenderTemplateAssignments.otpMessageTemplateId, otpMessageTemplates.id))
        .leftJoin(mnos, eq(otpSenderTemplateAssignments.mnoId, mnos.id))
        .where(whereClause)
        .orderBy(desc(otpSenderTemplateAssignments.priority))
        .limit(limit)
        .offset(offset),
    ])

    const total = totalResult[0]?.count || 0

    return NextResponse.json({
      assignments: assignmentList,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    })
  } catch (error) {
    console.error('[v0] Error fetching OTP sender template assignments:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { otpAlternativeSenderId, otpMessageTemplateId, mnoId, priority, status } = body

    if (!otpAlternativeSenderId || !otpMessageTemplateId) {
      return NextResponse.json({ error: 'OTP alternative sender ID and OTP message template ID are required' }, { status: 400 })
    }

    const result = await db.insert(otpSenderTemplateAssignments).values({
      otpAlternativeSenderId,
      otpMessageTemplateId,
      mnoId: mnoId || null,
      priority: priority || 0,
      status: status || 'active',
    }).returning()

    return NextResponse.json({ success: true, assignment: result[0] })
  } catch (error: any) {
    console.error('[v0] Error creating OTP sender template assignment:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'This assignment already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to create OTP sender template assignment' }, { status: 500 })
  }
}
