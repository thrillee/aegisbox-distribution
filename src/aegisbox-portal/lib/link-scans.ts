import { currentUser } from '@clerk/nextjs/server'

export const VERDICTS = ['safe', 'suspicious', 'malicious'] as const
export type Verdict = (typeof VERDICTS)[number]
export const SCAN_MODES = ['off', 'monitor', 'hold', 'block'] as const
export type ScanMode = (typeof SCAN_MODES)[number]

export const VERDICT_STYLES: Record<string, string> = {
  safe: 'bg-emerald-500/20 text-emerald-400',
  suspicious: 'bg-amber-500/20 text-amber-300',
  malicious: 'bg-red-500/20 text-red-400',
  pending: 'bg-zinc-500/20 text-zinc-300',
  failed: 'bg-zinc-500/20 text-zinc-400',
  none: 'bg-zinc-500/10 text-zinc-500',
}

export const VERDICT_RANK: Record<string, number> = { safe: 1, suspicious: 2, malicious: 3 }

// Who is acting, for the view log and decisions: the admin's email, or their Clerk user ID.
export async function actorName(): Promise<string> {
  try {
    const user = await currentUser()
    return user?.emailAddresses?.[0]?.emailAddress || user?.id || 'admin'
  } catch {
    return 'admin'
  }
}

// The stored form of a list domain: lowercase host without scheme, path, port or "*.". Mirrors
// linkscan.NormalizeListDomain in the gateway (which also converts IDN to punycode).
export function normalizeListDomain(input: string): string {
  let s = input.trim().toLowerCase()
  const scheme = s.indexOf('://')
  if (scheme >= 0) s = s.slice(scheme + 3)
  s = s.split(/[/?#]/)[0]
  const at = s.lastIndexOf('@')
  if (at >= 0) s = s.slice(at + 1)
  s = s.replace(/:\d+$/, '')
  s = s.replace(/^\*\./, '').replace(/^\.+|\.+$/g, '')
  try {
    // URL turns an IDN into punycode.
    s = new URL(`http://${s}`).hostname
  } catch {
    return ''
  }
  return s
}

export function isValidListDomain(d: string): boolean {
  return d.length > 0 && d.length <= 253 && d.includes('.') && /^[a-z0-9.-]+$/.test(d) && !d.includes('..')
}
