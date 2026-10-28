import { db } from '@/lib/db'
import { routingAssignments, mnos, msisdnPrefixGroups } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string; assignmentId: string }> }
) {
  try {
    const resolvedParams = await params
    const assignmentId = parseInt(resolvedParams.assignmentId)

    if (isNaN(assignmentId)) {
      return NextResponse.json({ error: 'Invalid assignment ID' }, { status: 400 })
    }

    const [assignment] = await db
      .select({
        id: routingAssignments.id,
        routingGroupId: routingAssignments.routingGroupId,
        msisdnPrefixGroupId: routingAssignments.msisdnPrefixGroupId,
        mnoId: routingAssignments.mnoId,
        priority: routingAssignments.priority,
        status: routingAssignments.status,
        comment: routingAssignments.comment,
        createdAt: routingAssignments.createdAt,
        mnoName: mnos.name,
        mnoCountryCode: mnos.countryCode,
        prefixGroupName: msisdnPrefixGroups.name,
        prefixGroupReference: msisdnPrefixGroups.reference,
      })
      .from(routingAssignments)
      .leftJoin(mnos, eq(routingAssignments.mnoId, mnos.id))
      .leftJoin(msisdnPrefixGroups, eq(routingAssignments.msisdnPrefixGroupId, msisdnPrefixGroups.id))
      .where(eq(routingAssignments.id, assignmentId))
      .limit(1)

    if (!assignment) {
      return NextResponse.json({ error: 'Routing assignment not found' }, { status: 404 })
    }

    return NextResponse.json({ assignment })
  } catch (error) {
    console.error('[v0] Error fetching routing assignment:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string; assignmentId: string }> }
) {
  try {
    const resolvedParams = await params
    const assignmentId = parseInt(resolvedParams.assignmentId)

    if (isNaN(assignmentId)) {
      return NextResponse.json({ error: 'Invalid assignment ID' }, { status: 400 })
    }

    const [deletedAssignment] = await db
      .delete(routingAssignments)
      .where(eq(routingAssignments.id, assignmentId))
      .returning()

    if (!deletedAssignment) {
      return NextResponse.json({ error: 'Routing assignment not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, message: 'Routing assignment deleted' })
  } catch (error: any) {
    console.error('[v0] Error deleting routing assignment:', error)
    return NextResponse.json({ error: 'Failed to delete routing assignment' }, { status: 500 })
  }
}

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string; assignmentId: string }> }
) {
  try {
    const resolvedParams = await params
    const assignmentId = parseInt(resolvedParams.assignmentId)

    if (isNaN(assignmentId)) {
      return NextResponse.json({ error: 'Invalid assignment ID' }, { status: 400 })
    }

    const body = await request.json()
    const { priority, status, comment } = body

    const [updatedAssignment] = await db
      .update(routingAssignments)
      .set({
        priority: priority,
        status: status,
        comment: comment,
        updatedAt: new Date(),
      })
      .where(eq(routingAssignments.id, assignmentId))
      .returning()

    if (!updatedAssignment) {
      return NextResponse.json({ error: 'Routing assignment not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, assignment: updatedAssignment })
  } catch (error: any) {
    console.error('[v0] Error updating routing assignment:', error)
    return NextResponse.json({ error: 'Failed to update routing assignment' }, { status: 500 })
  }
}
