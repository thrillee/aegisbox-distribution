// Reports count days in Lagos time, whatever time zone the server runs in.
// Africa/Lagos is UTC+1 all year (no daylight saving), so a Lagos day starts at 23:00 UTC.
export const REPORT_TIME_ZONE = 'Africa/Lagos'
const REPORT_UTC_OFFSET = '+01:00'

const ymdFormat = new Intl.DateTimeFormat('en-CA', {
  timeZone: REPORT_TIME_ZONE,
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
})

// The Lagos calendar date of an instant, as YYYY-MM-DD.
export function reportDate(at: Date = new Date()): string {
  return ymdFormat.format(at)
}

// Start of a Lagos day: today when daysAgo is 0, yesterday when 1, tomorrow when -1.
export function reportDayStart(daysAgo = 0, from: Date = new Date()): Date {
  const start = new Date(`${reportDate(from)}T00:00:00${REPORT_UTC_OFFSET}`)
  start.setUTCDate(start.getUTCDate() - daysAgo)
  return start
}

// Start of the current Lagos month.
export function reportMonthStart(from: Date = new Date()): Date {
  return new Date(`${reportDate(from).slice(0, 8)}01T00:00:00${REPORT_UTC_OFFSET}`)
}

// Whole Lagos days from start through end (both YYYY-MM-DD): from is the start of the first day,
// to is the start of the day after the last, so filter with >= from and < to.
export function reportDayRange(start: string, end: string): { from: Date; to: Date } {
  const from = new Date(`${start}T00:00:00${REPORT_UTC_OFFSET}`)
  const to = new Date(`${end}T00:00:00${REPORT_UTC_OFFSET}`)
  to.setUTCDate(to.getUTCDate() + 1)
  return { from, to }
}

export const isReportDate = (value: string) => /^\d{4}-\d{2}-\d{2}$/.test(value) && !Number.isNaN(Date.parse(value))

const dateTimeFormat = new Intl.DateTimeFormat('en-GB', {
  timeZone: REPORT_TIME_ZONE,
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
  hourCycle: 'h23', // 00:00 at midnight, never 24:00
})

// An instant as Lagos date and time, YYYY-MM-DD HH:MM:SS.
export function reportDateTime(at: Date): string {
  const p = Object.fromEntries(dateTimeFormat.formatToParts(at).map((x) => [x.type, x.value]))
  return `${p.year}-${p.month}-${p.day} ${p.hour}:${p.minute}:${p.second}`
}
