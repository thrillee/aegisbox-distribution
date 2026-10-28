import { NextRequest, NextResponse } from 'next/server'
import { db } from '@/lib/db'
import { senderIds } from '@/lib/schema'
import { requireServiceProvider } from '@/lib/service-provider'
import { desc, eq, and, like } from 'drizzle-orm'

const SENDER_STATUSES = ['pending', 'active', 'suspended', 'cancelled', 'inactive'] as const
type SenderStatus = (typeof SENDER_STATUSES)[number]

export async function GET(request: NextRequest) {
  try {
    const sp = await requireServiceProvider()
    if (sp.response) return sp.response

    const { searchParams } = new URL(request.url)
    const status = searchParams.get('status')
    const search = searchParams.get('search')

    const conditions = [eq(senderIds.serviceProviderId, sp.serviceProviderId)]
    if (status && (SENDER_STATUSES as readonly string[]).includes(status)) {
      conditions.push(eq(senderIds.status, status as SenderStatus))
    }
    if (search) {
      conditions.push(like(senderIds.senderIdString, `%${search}%`))
    }

    const senders = await db
      .select()
      .from(senderIds)
      .where(and(...conditions))
      .orderBy(desc(senderIds.createdAt))

    return NextResponse.json({ senderIds: senders })
  } catch (error) {
    console.error('Error fetching sender IDs:', error)
    return NextResponse.json({ error: 'Failed to fetch sender IDs' }, { status: 500 })
  }
}

// Requests a new sender ID. It is created as pending and only becomes usable once an
// administrator approves it and assigns its routing.
export async function POST(request: NextRequest) {
  try {
    const sp = await requireServiceProvider()
    if (sp.response) return sp.response

    const body = await request.json()
    const senderIdString = typeof body.senderIdString === 'string' ? body.senderIdString.trim() : ''
    const networks = Array.isArray(body.networks) ? body.networks.filter((n: unknown) => typeof n === 'string') : []
    const otpTemplate = typeof body.otpTemplate === 'string' && body.otpTemplate.trim() ? body.otpTemplate.trim() : null

    if (!senderIdString) {
      return NextResponse.json({ error: 'Sender ID is required' }, { status: 400 })
    }
    if (senderIdString.length > 16) {
      return NextResponse.json({ error: 'Sender ID must be at most 16 characters' }, { status: 400 })
    }

    const newSender = await db.insert(senderIds).values({
      serviceProviderId: sp.serviceProviderId,
      senderIdString,
      networks,
      otpTemplate,
      status: 'pending',
    }).returning()

    return NextResponse.json({ success: true, senderId: newSender[0] })
  } catch (error) {
    console.error('Error creating sender ID:', error)
    return NextResponse.json({ error: 'Failed to create sender ID' }, { status: 500 })
  }
}
