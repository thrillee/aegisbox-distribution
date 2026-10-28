import { db } from '@/lib/db'
import { serviceProviders } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { desc } from 'drizzle-orm'

export async function GET() {
  try {
    const providers = await db.select({
      id: serviceProviders.id,
      name: serviceProviders.name,
      email: serviceProviders.email,
      status: serviceProviders.status,
      defaultCurrencyCode: serviceProviders.defaultCurrencyCode,
      createdAt: serviceProviders.createdAt,
    })
      .from(serviceProviders)
      .orderBy(desc(serviceProviders.createdAt))

    return NextResponse.json({ providers })
  } catch (error) {
    console.error('[v0] Error fetching providers:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
