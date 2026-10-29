import { randomInt } from 'crypto'
import bcrypt from 'bcryptjs'
import { eq } from 'drizzle-orm'
import { db } from '@/lib/db'
import { spCredentials } from '@/lib/schema'

// SMPP 3.4 limits: system_id up to 15 characters, password up to 8.
export const SMPP_SYSTEM_ID_MAX = 15
export const SMPP_PASSWORD_MAX = 8

const ALPHABET = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789'

export function generateSmppPassword(): string {
  let out = ''
  for (let i = 0; i < SMPP_PASSWORD_MAX; i++) out += ALPHABET[randomInt(ALPHABET.length)]
  return out
}

// Returns an error message, or null when the values can be used for a bind.
export function validateSmppLogin(systemId: string, password: string): string | null {
  if (systemId.length < 1 || systemId.length > SMPP_SYSTEM_ID_MAX) {
    return `System ID must be 1 to ${SMPP_SYSTEM_ID_MAX} characters`
  }
  if (password.length < 1 || password.length > SMPP_PASSWORD_MAX) {
    return `Password must be 1 to ${SMPP_PASSWORD_MAX} characters (SMPP limit)`
  }
  return null
}

export async function systemIdTaken(systemId: string): Promise<boolean> {
  const rows = await db.select({ id: spCredentials.id }).from(spCredentials).where(eq(spCredentials.systemId, systemId)).limit(1)
  return rows.length > 0
}

// The gateway checks binds against this bcrypt hash; the password itself is only shown once.
export function hashSmppPassword(password: string): Promise<string> {
  return bcrypt.hash(password, 12)
}
