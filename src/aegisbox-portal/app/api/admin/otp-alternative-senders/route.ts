import { db } from '@/lib/db'
import { otpAlternativeSenders } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, count, and, like } from 'drizzle-orm'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    const page = parseInt(searchParams.get('page') || '1')
    const limit = parseInt(searchParams.get('limit') || '10')
    const offset = (page - 1) * limit
    const status = searchParams.get('status')
    const serviceProviderId = searchParams.get('serviceProviderId')
    const mnoId = searchParams.get('mnoId')
    const search = searchParams.get('search')

    const conditions = []
    if (status) conditions.push(eq(otpAlternativeSenders.status, status as any))
    if (serviceProviderId) conditions.push(eq(otpAlternativeSenders.serviceProviderId, parseInt(serviceProviderId)))
    if (mnoId) conditions.push(eq(otpAlternativeSenders.mnoId, parseInt(mnoId)))
    if (search) {
      conditions.push(like(otpAlternativeSenders.senderIdString, `%${search}%`))
    }

    const whereClause = conditions.length > 0 ? and(...conditions) : undefined

    const [totalResult, senderList] = await Promise.all([
      db.select({ count: count() }).from(otpAlternativeSenders).where(whereClause),
      db.select({
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
        .where(whereClause)
        .orderBy(desc(otpAlternativeSenders.createdAt))
        .limit(limit)
        .offset(offset),
    ])

    const total = totalResult[0]?.count || 0

    return NextResponse.json({
      senders: senderList,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    })
  } catch (error) {
    console.error('[v0] Error fetching OTP alternative senders:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { senderIdString, serviceProviderId, mnoId, status, maxUsageCount, resetIntervalHours, notes } = body

    if (!senderIdString) {
      return NextResponse.json({ error: 'Sender ID string is required' }, { status: 400 })
    }

    const result = await db.insert(otpAlternativeSenders).values({
      senderIdString,
      serviceProviderId: serviceProviderId || null,
      mnoId: mnoId || null,
      status: status || 'active',
      maxUsageCount: maxUsageCount || 10000,
      resetIntervalHours: resetIntervalHours || null,
      notes: notes || null,
    }).returning()

    return NextResponse.json({ success: true, sender: result[0] })
  } catch (error: any) {
    console.error('[v0] Error creating OTP alternative sender:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'This sender ID string already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to create OTP alternative sender' }, { status: 500 })
  }
}
