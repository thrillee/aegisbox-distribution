import { db } from '@/lib/db'
import { mnos, mnoConnections } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq, count } from 'drizzle-orm'

export async function GET() {
  try {
    const mnoList = await db.select({
      id: mnos.id,
      name: mnos.name,
      countryCode: mnos.countryCode,
      networkCode: mnos.networkCode,
      status: mnos.status,
      createdAt: mnos.createdAt,
    })
      .from(mnos)
      .orderBy(desc(mnos.name))

    const results = []
    for (const mno of mnoList) {
      const connections = await db.select({
        id: mnoConnections.id,
        status: mnoConnections.status,
      })
        .from(mnoConnections)
        .where(eq(mnoConnections.mnoId, mno.id))

      const activeConnections = connections.filter(c => c.status === 'active').length

      results.push({
        id: mno.id,
        name: mno.name,
        countryCode: mno.countryCode,
        networkCode: mno.networkCode,
        status: mno.status,
        createdAt: mno.createdAt,
        connectionsCount: connections.length,
        activeConnections,
        lastActivity: 'N/A',
        dailyMessages: 0,
        deliveryRate: 0,
      })
    }

    return NextResponse.json({ mnos: results })
  } catch (error) {
    console.error('[v0] Error fetching MNOs:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { name, countryCode, networkCode } = body

    if (!name || !countryCode) {
      return NextResponse.json({ error: 'Name and country code are required' }, { status: 400 })
    }

    const result = await db.insert(mnos).values({
      name,
      countryCode,
      networkCode,
      status: 'active',
    }).returning()

    return NextResponse.json({ success: true, mno: result[0] })
  } catch (error: any) {
    console.error('[v0] Error creating MNO:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'An MNO with this name already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to create MNO' }, { status: 500 })
  }
}
