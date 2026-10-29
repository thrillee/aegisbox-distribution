import { NextResponse } from 'next/server'
import { sql } from 'drizzle-orm'
import { db } from '@/lib/db'

// The router picks between a network's usable routes at random in proportion to their weight
// (routing_assignments.weight, 0 to 100). A route is usable when the assignment is active, its MNO
// is active, and route health has not taken it out: the same rules as the gateway's route lookup.
// It also needs an active connection to actually send.

interface RouteRow {
  assignment_id: number
  network_id: number
  network: string
  mno_id: number
  mno_name: string
  weight: number
  priority: number
  assignment_active: boolean
  mno_active: boolean
  has_connection: boolean
  manually_unavailable: boolean
  health_excluded: boolean
  consecutive_failures: number
}

async function loadRoutes(groupId: number) {
  const result = await db.execute(sql`
    WITH cfg AS (
      SELECT COALESCE(
        (SELECT route_failure_threshold FROM routing_config WHERE routing_group_id = ${groupId}),
        (SELECT route_failure_threshold FROM routing_config WHERE routing_group_id IS NULL LIMIT 1),
        5) AS failure_threshold
    )
    SELECT ra.id AS assignment_id,
           pg.id AS network_id,
           regexp_replace(pg.name, '\\s*Prefixes$', '') AS network,
           m.id AS mno_id,
           m.name AS mno_name,
           COALESCE(ra.weight, 100) AS weight,
           ra.priority,
           (ra.status = 'active' AND COALESCE(ra.is_active, true)) AS assignment_active,
           (m.status = 'active') AS mno_active,
           EXISTS (SELECT 1 FROM mno_connections mc WHERE mc.mno_id = m.id AND mc.status = 'active') AS has_connection,
           -- Mirrors the gateway: an admin's "unavailable" keeps a route out; an automatic
           -- exclusion (too many failed sends in a row) is retried every minute.
           (COALESCE(rh.is_available, true) = false AND COALESCE(rh.auto_unavailable, false) = false
              AND COALESCE(rh.consecutive_failures, 0) < cfg.failure_threshold) AS manually_unavailable,
           (COALESCE(rh.consecutive_failures, 0) >= cfg.failure_threshold
              OR (COALESCE(rh.is_available, true) = false AND COALESCE(rh.auto_unavailable, false) = true
                  AND COALESCE(rh.consecutive_failures, 0) > 0)) AS health_excluded,
           COALESCE(rh.consecutive_failures, 0) AS consecutive_failures
    FROM routing_assignments ra
    JOIN msisdn_prefix_groups pg ON pg.id = ra.msisdn_prefix_group_id
    JOIN mnos m ON m.id = ra.mno_id
    LEFT JOIN route_health rh ON rh.routing_assignment_id = ra.id AND rh.sender_id IS NULL
    CROSS JOIN cfg
    WHERE ra.routing_group_id = ${groupId}
    ORDER BY pg.name, ra.priority, m.name`)
  return result.rows as unknown as RouteRow[]
}

function parseGroupId(id: string) {
  return /^\d+$/.test(id) ? Number(id) : null
}

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const groupId = parseGroupId((await params).id)
  if (!groupId) {
    return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
  }
  try {
    return NextResponse.json({ routes: await loadRoutes(groupId) })
  } catch (error) {
    console.error('Error loading routing split:', error)
    return NextResponse.json({ error: 'Failed to load the traffic split' }, { status: 500 })
  }
}

// Sets one weight per MNO for every network in the group. Body: { weights: { [mnoId]: weight } }
// with a value for each MNO in the group.
export async function PUT(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const groupId = parseGroupId((await params).id)
  if (!groupId) {
    return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
  }
  const body = await request.json().catch(() => null)
  const input = body?.weights
  if (!input || typeof input !== 'object' || Array.isArray(input)) {
    return NextResponse.json({ error: 'weights must be an object of MNO ID to weight' }, { status: 400 })
  }

  try {
    const routes = await loadRoutes(groupId)
    if (routes.length === 0) {
      return NextResponse.json({ error: 'This routing group has no routes' }, { status: 400 })
    }

    const groupMnos = new Map(routes.map((r) => [Number(r.mno_id), r.mno_name]))
    const weights = new Map<number, number>()
    for (const [key, value] of Object.entries(input)) {
      const mnoId = Number(key)
      if (!groupMnos.has(mnoId)) {
        return NextResponse.json({ error: `MNO ${key} has no routes in this group` }, { status: 400 })
      }
      if (typeof value !== 'number' || !Number.isInteger(value) || value < 0 || value > 100) {
        return NextResponse.json({ error: `Weight for ${groupMnos.get(mnoId)} must be a whole number from 0 to 100` }, { status: 400 })
      }
      weights.set(mnoId, value)
    }
    const missing = [...groupMnos.keys()].filter((id) => !weights.has(id))
    if (missing.length > 0) {
      return NextResponse.json(
        { error: `Give a weight for every route: missing ${missing.map((id) => groupMnos.get(id)).join(', ')}` },
        { status: 400 }
      )
    }

    // Every network with an active route needs a positive weight on one of them, or the router
    // ignores the weights and splits that network's traffic evenly.
    const networks = new Map<string, number>()
    for (const r of routes) {
      if (!r.assignment_active || !r.mno_active) continue
      networks.set(r.network, (networks.get(r.network) ?? 0) + (weights.get(Number(r.mno_id)) ?? 0))
    }
    const zeroNetworks = [...networks.entries()].filter(([, total]) => total === 0).map(([n]) => n)
    if (zeroNetworks.length > 0) {
      return NextResponse.json(
        { error: `At least one active route needs a weight above 0 for: ${zeroNetworks.join(', ')}` },
        { status: 400 }
      )
    }

    await db.transaction(async (tx) => {
      for (const [mnoId, weight] of weights) {
        await tx.execute(sql`
          UPDATE routing_assignments
          SET weight = ${weight}, updated_at = NOW()
          WHERE routing_group_id = ${groupId} AND mno_id = ${mnoId}`)
      }
    })

    return NextResponse.json({ routes: await loadRoutes(groupId) })
  } catch (error) {
    console.error('Error saving routing split:', error)
    return NextResponse.json({ error: 'Failed to save the traffic split' }, { status: 500 })
  }
}
