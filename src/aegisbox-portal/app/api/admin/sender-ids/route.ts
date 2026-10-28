import { db } from '@/lib/db'
import { senderIds } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, count, and, like, or } from 'drizzle-orm'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    const page = parseInt(searchParams.get('page') || '1')
    const limit = parseInt(searchParams.get('limit') || '10')
    const offset = (page - 1) * limit
    const status = searchParams.get('status')
    const serviceProviderId = searchParams.get('serviceProviderId')
    const mnoConnectionId = searchParams.get('mnoConnectionId')
    const search = searchParams.get('search')

    const conditions = []
    if (status && status !== 'all') conditions.push(eq(senderIds.status, status as any))
    if (serviceProviderId) conditions.push(eq(senderIds.serviceProviderId, parseInt(serviceProviderId)))
    if (mnoConnectionId) conditions.push(eq(senderIds.mnoConnectionId, parseInt(mnoConnectionId)))
    if (search) {
      conditions.push(like(senderIds.senderIdString, `%${search}%`))
    }

    const whereClause = conditions.length > 0 ? and(...conditions) : undefined

    const [totalResult, senderList] = await Promise.all([
      db.select({ count: count() }).from(senderIds).where(whereClause),
      db.select({
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
        .where(whereClause)
        .orderBy(desc(senderIds.createdAt))
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
    console.error('[v0] Error fetching sender IDs:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { senderIdString, serviceProviderId, mnoConnectionId, status, rateLimit, otpTemplate, networks, maxUsageCount, resetIntervalHours } = body

    if (!senderIdString) {
      return NextResponse.json({ error: 'Sender ID string is required' }, { status: 400 })
    }

    const result = await db.insert(senderIds).values({
      senderIdString,
      serviceProviderId: serviceProviderId || null,
      mnoConnectionId: mnoConnectionId || null,
      status: status || 'pending',
      rateLimit: rateLimit || 100,
      otpTemplate: otpTemplate || null,
      networks: networks || null,
      maxUsageCount: maxUsageCount || 10000,
      resetIntervalHours: resetIntervalHours || null,
    }).returning()

    return NextResponse.json({ success: true, sender: result[0] })
  } catch (error: any) {
    console.error('[v0] Error creating sender ID:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'This sender ID string already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to create sender ID' }, { status: 500 })
  }
}
