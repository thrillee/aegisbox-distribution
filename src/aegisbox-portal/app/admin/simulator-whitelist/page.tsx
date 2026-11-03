import { asc, desc, eq } from 'drizzle-orm'
import { db } from '@/lib/db'
import { serviceProviders, simulatorWhitelistNumbers } from '@/lib/schema'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { reportDateTime } from '@/lib/time'
import { WhitelistForm, RemoveButton } from './whitelist-client'

export const dynamic = 'force-dynamic'

export default async function SimulatorWhitelistPage() {
  const [rows, testCustomers] = await Promise.all([
    db.select({
      id: simulatorWhitelistNumbers.id,
      msisdn: simulatorWhitelistNumbers.msisdn,
      note: simulatorWhitelistNumbers.note,
      createdAt: simulatorWhitelistNumbers.createdAt,
      spId: simulatorWhitelistNumbers.serviceProviderId,
      spName: serviceProviders.name,
    })
      .from(simulatorWhitelistNumbers)
      .leftJoin(serviceProviders, eq(serviceProviders.id, simulatorWhitelistNumbers.serviceProviderId))
      .orderBy(desc(simulatorWhitelistNumbers.createdAt)),
    db.select({ id: serviceProviders.id, name: serviceProviders.name })
      .from(serviceProviders)
      .where(eq(serviceProviders.isTest, true))
      .orderBy(asc(serviceProviders.name)),
  ])

  return (
    <div className="p-4 md:p-8 space-y-6">
      <PageHeader
        title="Simulator Whitelist"
        description="Numbers that get real delivery. A test customer's message to one of these numbers goes through a real network instead of the simulated MNO, so the phone receives it and the real DLR comes back."
      />

      <DataCard title="Add a number" subtitle="The test customer's routing group must also contain a real route for the number's network; the customer is charged its normal price.">
        <WhitelistForm testCustomers={testCustomers} />
      </DataCard>

      <DataCard title="Whitelisted numbers" subtitle={`${rows.length} number${rows.length === 1 ? '' : 's'}`}>
        {rows.length === 0 ? (
          <p className="text-sm text-zinc-400 py-6 text-center">No numbers yet. Every test message goes to the simulator.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-zinc-800/50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Number</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Applies to</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Note</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Added</th>
                  <th className="px-4 py-3 text-right text-xs font-semibold text-zinc-400 uppercase"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/10">
                {rows.map((r) => (
                  <tr key={r.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-4 py-3 font-mono text-white">{r.msisdn}</td>
                    <td className="px-4 py-3 text-zinc-300">{r.spId === null ? 'All test customers' : (r.spName ?? `Customer ${r.spId}`)}</td>
                    <td className="px-4 py-3 text-zinc-400">{r.note ?? ''}</td>
                    <td className="px-4 py-3 text-zinc-400 whitespace-nowrap">{reportDateTime(r.createdAt)}</td>
                    <td className="px-4 py-3 text-right"><RemoveButton id={r.id} msisdn={r.msisdn} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </DataCard>
    </div>
  )
}
