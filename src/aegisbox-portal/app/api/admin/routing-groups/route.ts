import { db } from '@/lib/db'
import { routingGroups } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc, eq } from 'drizzle-orm'

export async function GET() {
  try {
    const groups = await db
      .select({
        id: routingGroups.id,
        name: routingGroups.name,
        reference: routingGroups.reference,
        description: routingGroups.description,
        createdAt: routingGroups.createdAt,
      })
      .from(routingGroups)
      .orderBy(routingGroups.name)

    return NextResponse.json({ routingGroups: groups })
  } catch (error) {
    console.error('[v0] Error fetching routing groups:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { name, reference, description } = body

    if (!name || !reference) {
      return NextResponse.json({ error: 'Name and reference are required' }, { status: 400 })
    }

    const [group] = await db
      .insert(routingGroups)
      .values({
        name,
        reference,
        description,
      })
      .returning()

    return NextResponse.json({ success: true, routingGroup: group })
  } catch (error: any) {
    console.error('[v0] Error creating routing group:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'A routing group with this reference already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to create routing group' }, { status: 500 })
  }
}
