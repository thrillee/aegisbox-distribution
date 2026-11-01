import { sql, type SQL } from 'drizzle-orm'

// Test customers (service_providers.is_test) send traffic to simulator MNOs. Platform totals
// (dashboard, analytics, daily report) leave them out unless asked to include them.

// True when the service provider with this ID is not a test customer. spId is a column or SQL
// expression for a service_providers.id.
export function notTestCustomer(spId: SQL | unknown): SQL {
  return sql`NOT EXISTS (SELECT 1 FROM service_providers tsp WHERE tsp.id = ${spId} AND tsp.is_test)`
}

export const includeTestParam = (v: string | null) => v === '1' || v === 'true'
