import { db } from '@/lib/db'
import { spCredentials } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq, and } from 'drizzle-orm'

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string; credId: string }> }
) {
  try {
    const resolvedParams = await params
    const spId = parseInt(resolvedParams.id)
    const credId = parseInt(resolvedParams.credId)

    if (isNaN(spId) || isNaN(credId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const body = await request.json()
    const { status, routingGroupId } = body

    const updateData: Record<string, any> = {
      updatedAt: new Date(),
    }

    if (status !== undefined) {
      updateData.status = status
    }

    if (routingGroupId !== undefined) {
      updateData.routingGroupId = routingGroupId ? parseInt(routingGroupId) : null
    }

    const [credential] = await db
      .update(spCredentials)
      .set(updateData)
      .where(
        and(
          eq(spCredentials.id, credId),
          eq(spCredentials.serviceProviderId, spId)
        )
      )
      .returning()

    if (!credential) {
      return NextResponse.json({ error: 'Credential not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, credential })
  } catch (error) {
    console.error('[v0] Error updating SMPP credential:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
