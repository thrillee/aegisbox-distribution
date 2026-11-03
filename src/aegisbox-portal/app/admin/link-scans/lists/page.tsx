import Link from 'next/link'
import { sql } from 'drizzle-orm'
import { db } from '@/lib/db'
import { reportDateTime } from '@/lib/time'
import { ListEntryForm, RemoveListEntry } from '../actions'

export const dynamic = 'force-dynamic'

interface Entry {
  id: number
  domain: string
  list_type: string
  service_provider_id: number | null
  customer: string | null
  note: string | null
  created_by: string | null
  updated_at: Date
}

export default async function LinkListsPage() {
  const [entriesRes, customersRes] = await Promise.all([
    db.execute(sql`
      SELECT l.id, l.domain, l.list_type, l.service_provider_id, sp.name AS customer, l.note, l.created_by, l.updated_at
      FROM link_domain_lists l
      LEFT JOIN service_providers sp ON sp.id = l.service_provider_id
      ORDER BY l.list_type DESC, l.domain`),
    db.execute(sql`SELECT id, name FROM service_providers ORDER BY name`),
  ])
  const entries = entriesRes.rows as unknown as Entry[]
  const customers = customersRes.rows as unknown as { id: number; name: string }[]

  return (
    <div className="p-4 md:p-8 space-y-6">
      <div>
        <Link href="/admin/link-scans" className="text-sm text-zinc-400 hover:text-white">← Link Scans</Link>
        <h1 className="text-3xl font-bold text-white mt-2">Allow and block lists</h1>
        <p className="text-zinc-400 mt-1">
          A domain covers its subdomains. Blocked links are judged malicious and allowed links safe, without opening them;
          a link that leads to a listed domain is judged by the list too. A customer&apos;s own entry wins over an entry for
          every customer, and block wins over allow. Changes apply to new scans within 30 seconds.
        </p>
      </div>

      <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
        <h2 className="font-semibold text-white mb-4">Add a domain</h2>
        <ListEntryForm customers={customers} />
      </div>

      <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
        <h2 className="font-semibold text-white mb-4">{entries.length} entr{entries.length === 1 ? 'y' : 'ies'}</h2>
        {entries.length === 0 ? (
          <p className="text-sm text-zinc-500 py-4 text-center">No entries yet.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-zinc-400 border-b border-white/10">
                  <th className="py-2 pr-4 font-medium">Domain</th>
                  <th className="py-2 pr-4 font-medium">List</th>
                  <th className="py-2 pr-4 font-medium">Applies to</th>
                  <th className="py-2 pr-4 font-medium">Note</th>
                  <th className="py-2 pr-4 font-medium">Changed</th>
                  <th className="py-2 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {entries.map((e) => (
                  <tr key={e.id} className="border-b border-white/5">
                    <td className="py-2 pr-4 font-mono text-white">{e.domain}</td>
                    <td className="py-2 pr-4">
                      <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${e.list_type === 'block' ? 'bg-red-500/20 text-red-400' : 'bg-emerald-500/20 text-emerald-400'}`}>{e.list_type}</span>
                    </td>
                    <td className="py-2 pr-4 text-zinc-300">{e.service_provider_id === null ? 'Every customer' : (e.customer ?? `Customer ${e.service_provider_id}`)}</td>
                    <td className="py-2 pr-4 text-zinc-400">{e.note ?? ''}</td>
                    <td className="py-2 pr-4 text-zinc-500 whitespace-nowrap">{reportDateTime(new Date(e.updated_at))}{e.created_by ? ` by ${e.created_by}` : ''}</td>
                    <td className="py-2 text-right"><RemoveListEntry id={e.id} domain={e.domain} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
