import { auth, currentUser } from '@clerk/nextjs/server'
import { NextResponse } from 'next/server'
import { db } from '@/lib/db'
import { serviceProviders, wallets, orgMappings } from '@/lib/schema'

export async function POST(req: Request) {
  try {
    const { userId } = await auth()
    if (!userId) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const user = await currentUser()
    if (!user) {
      return NextResponse.json({ error: 'User not found' }, { status: 404 })
    }

    const body = await req.json()
    const { orgName, slug } = body

    if (!orgName || !slug) {
      return NextResponse.json({ error: 'Organization name and slug are required' }, { status: 400 })
    }

    const slugRegex = /^[a-z0-9-]+$/
    if (!slugRegex.test(slug)) {
      return NextResponse.json({ error: 'Slug must contain only lowercase letters, numbers, and hyphens' }, { status: 400 })
    }

    const clerkResponse = await fetch(`https://api.clerk.com/v1/organizations`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${process.env.CLERK_SECRET_KEY}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: orgName,
        slug: slug,
        created_by: userId,
      }),
    })

    if (!clerkResponse.ok) {
      const errorData = await clerkResponse.json()
      if (errorData.errors?.[0]?.code === 'form_identifier_exists') {
        return NextResponse.json({ error: 'Organization slug already exists' }, { status: 400 })
      }
      throw new Error('Failed to create organization in Clerk')
    }

    const organization = await clerkResponse.json()

    const sp = await db.insert(serviceProviders).values({
      name: orgName,
      email: user.emailAddresses[0]?.emailAddress || '',
      status: 'active',
      defaultCurrencyCode: 'NGN',
    }).returning().execute()

    await db.insert(wallets).values({
      serviceProviderId: sp[0].id,
      currencyCode: 'NGN',
      balance: '0.000000',
      walletType: 'prepaid',
    }).execute()

    await db.insert(orgMappings).values({
      clerkOrgId: organization.id,
      serviceProviderId: sp[0].id,
    }).execute()

    const updateResponse = await fetch(`https://api.clerk.com/v1/organizations/${organization.id}/metadata`, {
      method: 'PATCH',
      headers: {
        'Authorization': `Bearer ${process.env.CLERK_SECRET_KEY}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        public_metadata: {
          serviceProviderId: sp[0].id,
        },
      }),
    })

    if (!updateResponse.ok) {
      console.warn('Failed to update organization metadata in Clerk')
    }

    return NextResponse.json({ 
      orgId: organization.id,
      message: 'Organization created successfully' 
    })
  } catch (error) {
    console.error('Error creating organization:', error)
    
    if (error instanceof Error && error.message.includes('slug')) {
      return NextResponse.json({ error: 'Organization slug already exists' }, { status: 400 })
    }
    
    return NextResponse.json({ error: 'Failed to create organization' }, { status: 500 })
  }
}
