import { db } from '@/lib/db'
import { msisdnPrefixGroups, msisdnPrefixEntries } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq, desc } from 'drizzle-orm'

export async function GET() {
  try {
    const groups = await db.select({
      id: msisdnPrefixGroups.id,
      name: msisdnPrefixGroups.name,
      description: msisdnPrefixGroups.description,
      reference: msisdnPrefixGroups.reference,
      createdAt: msisdnPrefixGroups.createdAt,
    })
      .from(msisdnPrefixGroups)
      .orderBy(desc(msisdnPrefixGroups.createdAt))

    for (const group of groups) {
      const prefixes = await db.select({
        id: msisdnPrefixEntries.id,
        prefix: msisdnPrefixEntries.msisdnPrefix,
      })
        .from(msisdnPrefixEntries)
        .where(eq(msisdnPrefixEntries.msisdnPrefixGroupId, group.id))
      ;(group as any).prefixes = prefixes
    }

    return NextResponse.json({ groups })
  } catch (error) {
    console.error('[v0] Error fetching prefix groups:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { name, description, reference, prefixes } = body

    if (!name || !reference) {
      return NextResponse.json({ error: 'Name and reference are required' }, { status: 400 })
    }

    const result = await db.insert(msisdnPrefixGroups).values({
      name,
      description,
      reference,
    }).returning()

    const group = result[0]

    if (prefixes && Array.isArray(prefixes) && prefixes.length > 0) {
      for (const prefix of prefixes) {
        if (prefix && prefix.trim()) {
          await db.insert(msisdnPrefixEntries).values({
            msisdnPrefixGroupId: group.id,
            msisdnPrefix: prefix.trim(),
          })
        }
      }
    }

    return NextResponse.json({ success: true, group: { ...group, prefixes: [] } })
  } catch (error: any) {
    console.error('[v0] Error creating prefix group:', error)
    if (error.code === '23505') {
      return NextResponse.json({ error: 'A group with this name or reference already exists' }, { status: 409 })
    }
    return NextResponse.json({ error: 'Failed to create prefix group' }, { status: 500 })
  }
}
