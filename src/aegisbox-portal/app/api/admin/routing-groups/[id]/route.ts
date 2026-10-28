import { db } from '@/lib/db'
import { routingGroups, routingAssignments, mnos, msisdnPrefixGroups, msisdnPrefixEntries } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq, desc } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const id = parseInt(resolvedParams.id)

    if (isNaN(id)) {
      return NextResponse.json({ error: 'Invalid routing group ID' }, { status: 400 })
    }

    const [group] = await db
      .select({
        id: routingGroups.id,
        name: routingGroups.name,
        reference: routingGroups.reference,
        description: routingGroups.description,
        createdAt: routingGroups.createdAt,
        updatedAt: routingGroups.updatedAt,
      })
      .from(routingGroups)
      .where(eq(routingGroups.id, id))
      .limit(1)

    if (!group) {
      return NextResponse.json({ error: 'Routing group not found' }, { status: 404 })
    }

    const assignments = await db
      .select({
        id: routingAssignments.id,
        priority: routingAssignments.priority,
        status: routingAssignments.status,
        mnoId: routingAssignments.mnoId,
        mnoName: mnos.name,
        mnoCountryCode: mnos.countryCode,
        mnoNetworkCode: mnos.networkCode,
        msisdnPrefixGroupId: routingAssignments.msisdnPrefixGroupId,
        prefixGroupName: msisdnPrefixGroups.name,
        prefixGroupReference: msisdnPrefixGroups.reference,
      })
      .from(routingAssignments)
      .leftJoin(mnos, eq(routingAssignments.mnoId, mnos.id))
      .leftJoin(msisdnPrefixGroups, eq(routingAssignments.msisdnPrefixGroupId, msisdnPrefixGroups.id))
      .where(eq(routingAssignments.routingGroupId, id))
      .orderBy(desc(routingAssignments.priority))

    const prefixEntries = await db
      .select({
        id: msisdnPrefixEntries.id,
        msisdnPrefix: msisdnPrefixEntries.msisdnPrefix,
        prefixGroupId: msisdnPrefixEntries.msisdnPrefixGroupId,
      })
      .from(msisdnPrefixEntries)
      .innerJoin(msisdnPrefixGroups, eq(msisdnPrefixEntries.msisdnPrefixGroupId, msisdnPrefixGroups.id))
      .where(eq(msisdnPrefixGroups.id, assignments[0]?.msisdnPrefixGroupId || 0))

    return NextResponse.json({
      routingGroup: group,
      assignments,
      prefixes: prefixEntries,
    })
  } catch (error) {
    console.error('[v0] Error fetching routing group:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const id = parseInt(resolvedParams.id)

    if (isNaN(id)) {
      return NextResponse.json({ error: 'Invalid routing group ID' }, { status: 400 })
    }

    const body = await request.json()
    const { name, reference, description } = body

    const [group] = await db
      .update(routingGroups)
      .set({
        name: name,
        reference: reference,
        description: description,
        updatedAt: new Date(),
      })
      .where(eq(routingGroups.id, id))
      .returning()

    if (!group) {
      return NextResponse.json({ error: 'Routing group not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, routingGroup: group })
  } catch (error: any) {
    console.error('[v0] Error updating routing group:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'A routing group with this reference already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to update routing group' }, { status: 500 })
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const id = parseInt(resolvedParams.id)

    if (isNaN(id)) {
      return NextResponse.json({ error: 'Invalid routing group ID' }, { status: 400 })
    }

    const [deletedGroup] = await db
      .delete(routingGroups)
      .where(eq(routingGroups.id, id))
      .returning()

    if (!deletedGroup) {
      return NextResponse.json({ error: 'Routing group not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, message: 'Routing group deleted' })
  } catch (error: any) {
    console.error('[v0] Error deleting routing group:', error)
    if (error.code === '23503') {
      return NextResponse.json({ error: 'Cannot delete routing group with existing assignments' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to delete routing group' }, { status: 500 })
  }
}
