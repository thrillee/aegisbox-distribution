// Normalizes a phone number the way the gateway does for the simulator whitelist: digits only,
// international format; local Nigerian numbers (11 digits starting with 0) get 234.
export function normalizeMsisdn(input: string): string {
  let d = input.replace(/\D/g, '')
  if (d.length === 11 && d.startsWith('0')) d = '234' + d.slice(1)
  return d
}

// Accepts 8 to 15 digits (E.164 length) after normalizing.
export function isValidMsisdn(d: string): boolean {
  return /^\d{8,15}$/.test(d)
}
