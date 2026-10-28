const CURRENCY_SYMBOLS: Record<string, string> = {
  NGN: '₦',
  USD: '$',
  EUR: '€',
  GBP: '£',
  GHS: 'GH₵',
  KES: 'KSh',
}

export function currencySymbol(currency: string | null | undefined): string {
  if (!currency) return ''
  return CURRENCY_SYMBOLS[currency.toUpperCase()] ?? `${currency.toUpperCase()} `
}

// Formats an amount in its own currency. Per-SMS prices and charges are fractions of a unit
// (e.g. USD 0.0056), so up to 6 decimals are shown instead of rounding them away.
export function formatMoney(amount: number | string | null | undefined, currency: string | null | undefined): string {
  const value = Number(amount ?? 0)
  const safe = Number.isFinite(value) ? value : 0
  const digits = Math.abs(safe).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 6 })
  return `${safe < 0 ? '-' : ''}${currencySymbol(currency)}${digits}`
}
