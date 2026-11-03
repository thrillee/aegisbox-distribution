import { NextResponse } from 'next/server'
import { and, eq, isNull } from 'drizzle-orm'
import { db } from '@/lib/db'
import { serviceProviders, simulatorWhitelistNumbers } from '@/lib/schema'
import { isValidMsisdn, normalizeMsisdn } from '@/lib/msisdn'

// Adds a whitelisted number. Body: { msisdn: string, service_provider_id?: number | null, note?: string }
export async function POST(request: Request) {
  const body = await request.json().catch(() => null)
  const msisdn = normalizeMsisdn(String(body?.msisdn ?? ''))
  if (!isValidMsisdn(msisdn)) {
    return NextResponse.json({ error: 'Enter a phone number in international format, for example 2348031234567' }, { status: 400 })
  }
  const spRaw = body?.service_provider_id
  const spId = spRaw === null || spRaw === undefined || spRaw === '' ? null : Number(spRaw)
  if (spId !== null && !Number.isInteger(spId)) {
    return NextResponse.json({ error: 'Invalid customer' }, { status: 400 })
  }
  const note = typeof body?.note === 'string' ? body.note.trim().slice(0, 200) || null : null

  try {
    if (spId !== null) {
      const [sp] = await db.select({ id: serviceProviders.id, isTest: serviceProviders.isTest })
        .from(serviceProviders).where(eq(serviceProviders.id, spId))
      if (!sp) return NextResponse.json({ error: 'Customer not found' }, { status: 404 })
      // The whitelist only changes routing for test customers; for others it would do nothing.
      if (!sp.isTest) return NextResponse.json({ error: 'Only test customers use the simulator whitelist' }, { status: 400 })
    }
    const existing = await db.select({ id: simulatorWhitelistNumbers.id }).from(simulatorWhitelistNumbers).where(and(
      eq(simulatorWhitelistNumbers.msisdn, msisdn),
      spId === null ? isNull(simulatorWhitelistNumbers.serviceProviderId) : eq(simulatorWhitelistNumbers.serviceProviderId, spId),
    ))
    if (existing.length) return NextResponse.json({ error: 'This number is already whitelisted' }, { status: 409 })
    const [row] = await db.insert(simulatorWhitelistNumbers)
      .values({ msisdn, serviceProviderId: spId, note })
      .returning()
    return NextResponse.json({ number: row }, { status: 201 })
  } catch (error) {
    // Two admins adding the same number at once: the unique index catches the second.
    const code = (error as { code?: string; cause?: { code?: string } })?.code ?? (error as { cause?: { code?: string } })?.cause?.code
    if (code === '23505') return NextResponse.json({ error: 'This number is already whitelisted' }, { status: 409 })
    console.error('Error adding whitelisted number:', error)
    return NextResponse.json({ error: 'Failed to add the number' }, { status: 500 })
  }
}
