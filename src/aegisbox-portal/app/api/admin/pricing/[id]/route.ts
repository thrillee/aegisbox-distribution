import { db } from '@/lib/db'
import { pricingRules } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const ruleId = parseInt(resolvedParams.id)

    if (isNaN(ruleId)) {
      return NextResponse.json({ error: 'Invalid pricing rule ID' }, { status: 400 })
    }

    const rule = await db.query.pricingRules.findFirst({
      where: eq(pricingRules.id, ruleId),
    })

    if (!rule) {
      return NextResponse.json({ error: 'Pricing rule not found' }, { status: 404 })
    }

    return NextResponse.json({ pricingRule: rule })
  } catch (error) {
    console.error('[v0] Error fetching pricing rule:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const ruleId = parseInt(resolvedParams.id)

    if (isNaN(ruleId)) {
      return NextResponse.json({ error: 'Invalid pricing rule ID' }, { status: 400 })
    }

    const body = await request.json()
    const { mnoId, currencyCode, pricePerSms } = body

    await db.update(pricingRules)
      .set({
        mnoId,
        currencyCode,
        pricePerSms,
        updatedAt: new Date(),
      })
      .where(eq(pricingRules.id, ruleId))

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('[v0] Error updating pricing rule:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const ruleId = parseInt(resolvedParams.id)

    if (isNaN(ruleId)) {
      return NextResponse.json({ error: 'Invalid pricing rule ID' }, { status: 400 })
    }

    await db.delete(pricingRules).where(eq(pricingRules.id, ruleId))

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('[v0] Error deleting pricing rule:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
