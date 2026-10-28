import { db } from '@/lib/db'
import { pricingRules, mnos } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq, desc, asc } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const spId = parseInt(resolvedParams.id)

    if (isNaN(spId)) {
      return NextResponse.json({ error: 'Invalid service provider ID' }, { status: 400 })
    }

    const { searchParams } = new URL(request.url)
    const sort = searchParams.get('sort') === 'asc' ? asc : desc

    const rules = await db
      .select({
        id: pricingRules.id,
        mnoId: pricingRules.mnoId,
        mnoName: mnos.name,
        mnoCountryCode: mnos.countryCode,
        currencyCode: pricingRules.currencyCode,
        pricePerSms: pricingRules.pricePerSms,
        createdAt: pricingRules.createdAt,
        updatedAt: pricingRules.updatedAt,
      })
      .from(pricingRules)
      .leftJoin(mnos, eq(pricingRules.mnoId, mnos.id))
      .where(eq(pricingRules.serviceProviderId, spId))
      .orderBy(sort(pricingRules.createdAt))

    return NextResponse.json({ pricingRules: rules })
  } catch (error) {
    console.error('[v0] Error fetching pricing rules:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const spId = parseInt(resolvedParams.id)

    if (isNaN(spId)) {
      return NextResponse.json({ error: 'Invalid service provider ID' }, { status: 400 })
    }

    const body = await request.json()
    const { mnoId, currencyCode, pricePerSms } = body

    if (!pricePerSms) {
      return NextResponse.json({ error: 'Price per SMS is required' }, { status: 400 })
    }

    const [rule] = await db
      .insert(pricingRules)
      .values({
        serviceProviderId: spId,
        mnoId: mnoId ? parseInt(mnoId) : null,
        currencyCode: currencyCode || 'NGN',
        pricePerSms: pricePerSms.toString(),
      })
      .returning()

    return NextResponse.json({ success: true, pricingRule: rule })
  } catch (error) {
    console.error('[v0] Error creating pricing rule:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
