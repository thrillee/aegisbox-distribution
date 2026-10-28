import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'
import { NextResponse } from 'next/server'

const isPublicRoute = createRouteMatcher([
  '/',
  '/sign-in(.*)',
  '/sign-up(.*)',
  '/auth-check',
  '/api/health',
])

const isAdminRoute = createRouteMatcher(['/admin(.*)'])
const isSPRoute = createRouteMatcher(['/sp(.*)'])
const isAPIAdminRoute = createRouteMatcher(['/api/admin(.*)'])
const isAPISPRoute = createRouteMatcher(['/api/sp(.*)'])

export default clerkMiddleware(async (auth, req) => {
  const { userId, orgId, sessionClaims, redirectToSignIn } = await auth()

  // Allow public routes without auth
  if (isPublicRoute(req)) {
    return NextResponse.next()
  }

  // Require authentication for all other routes
  if (!userId) {
    return redirectToSignIn({ returnBackUrl: req.url })
  }

  // Admin routes - check admin permissions
  if (isAPIAdminRoute(req) || isAdminRoute(req)) {
    const adminOrgId = process.env.CLERK_ADMIN_ORG_ID
    const isAdmin = orgId === adminOrgId || 
      (sessionClaims?.metadata as Record<string, unknown>)?.role === 'admin'
    
    if (!isAdmin) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 403 })
    }
    return NextResponse.next()
  }

  // SP routes - require org membership
  if (isSPRoute(req) || isAPISPRoute(req)) {
    // Check if user is admin - redirect to admin dashboard
    const adminOrgId = process.env.CLERK_ADMIN_ORG_ID
    const isAdmin = orgId === adminOrgId || 
      (sessionClaims?.metadata as Record<string, unknown>)?.role === 'admin'
    
    if (isAdmin) {
      return NextResponse.redirect(new URL('/admin/dashboard', req.url))
    }

    if (!orgId) {
      const url = new URL('/auth-check', req.url)
      url.searchParams.set('reason', 'no_org')
      return NextResponse.redirect(url)
    }
    return NextResponse.next()
  }

  return NextResponse.next()
})

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml|.*\\.(?:ico|png|jpg|jpeg|gif|svg|woff|woff2|ttf|eot)).*)',
  ],
}
