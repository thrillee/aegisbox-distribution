import { sql } from 'drizzle-orm'
import { REPORT_TIME_ZONE } from '@/lib/time'

// The report time zone as an SQL literal (not a bind parameter), so the same day expression can
// appear in both SELECT and GROUP BY.
export const REPORT_TIME_ZONE_SQL = sql.raw(`'${REPORT_TIME_ZONE}'`)
