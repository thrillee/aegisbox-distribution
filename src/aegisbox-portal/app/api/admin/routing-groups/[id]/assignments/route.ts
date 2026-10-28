import { db } from '@/lib/db'
import { routingAssignments } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const routingGroupId = parseInt(resolvedParams.id)

    if (isNaN(routingGroupId)) {
      return NextResponse.json({ error: 'Invalid routing group ID' }, { status: 400 })
    }

    const body = await request.json()
    const { msisdnPrefixGroupId, mnoId, priority, comment } = body

    if (!msisdnPrefixGroupId || !mnoId) {
      return NextResponse.json({ error: 'MSISDN prefix group ID and MNO ID are required' }, { status: 400 })
    }

    const [assignment] = await db
      .insert(routingAssignments)
      .values({
        routingGroupId,
        msisdnPrefixGroupId,
        mnoId,
        priority: priority || 0,
        comment,
      })
      .returning()

    return NextResponse.json({ success: true, assignment })
  } catch (error: any) {
    console.error('[v0] Error creating routing assignment:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'An assignment for this prefix group already exists in this routing group' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to create routing assignment' }, { status: 500 })
  }
}

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const routingGroupId = parseInt(resolvedParams.id)

    if (isNaN(routingGroupId)) {
      return NextResponse.json({ error: 'Invalid routing group ID' }, { status: 400 })
    }

    const assignments = await db
      .select({
        id: routingAssignments.id,
        msisdnPrefixGroupId: routingAssignments.msisdnPrefixGroupId,
        mnoId: routingAssignments.mnoId,
        priority: routingAssignments.priority,
        status: routingAssignments.status,
        comment: routingAssignments.comment,
        createdAt: routingAssignments.createdAt,
      })
      .from(routingAssignments)
      .where(eq(routingAssignments.routingGroupId, routingGroupId))
      .orderBy(routingAssignments.priority)

    return NextResponse.json({ assignments })
  } catch (error) {
    console.error('[v0] Error fetching routing assignments:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
