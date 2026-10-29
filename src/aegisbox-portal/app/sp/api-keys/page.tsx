import { Info } from 'lucide-react'
import { desc, eq } from 'drizzle-orm'
import { db } from '@/lib/db'
import { getCurrentServiceProviderId, NOT_LINKED_MESSAGE } from '@/lib/service-provider'
import { spCredentials } from '@/lib/schema'
import { httpApiPublicUrl, smppConnectionInfo } from '@/lib/connection-info'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'

// The API key is the bearer token itself, so only a short prefix is shown here. Logins and keys
// are issued and reset by your account manager.
const VISIBLE_KEY_PREFIX = 8

async function getCredentials(serviceProviderId: number) {
  return db
    .select({
      id: spCredentials.id,
      protocol: spCredentials.protocol,
      systemId: spCredentials.systemId,
      apiKeyIdentifier: spCredentials.apiKeyIdentifier,
      bindType: spCredentials.bindType,
      status: spCredentials.status,
      createdAt: spCredentials.createdAt,
    })
    .from(spCredentials)
    .where(eq(spCredentials.serviceProviderId, serviceProviderId))
    .orderBy(desc(spCredentials.createdAt))
}

const STATUS_STYLES: Record<string, string> = {
  active: 'bg-emerald-500/20 text-emerald-400',
  pending: 'bg-amber-500/20 text-amber-400',
}

function CredentialStatus({ status }: Readonly<{ status: string }>) {
  return (
    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium capitalize ${STATUS_STYLES[status] ?? 'bg-zinc-500/20 text-zinc-400'}`}>
      {status}
    </span>
  )
}

const BIND_TYPES: Record<string, string> = { trx: 'Transceiver', tx: 'Transmitter', rx: 'Receiver' }

function Detail({ label, value }: Readonly<{ label: string; value: string | null }>) {
  return (
    <div>
      <p className="text-xs text-zinc-500">{label}</p>
      <p className={`font-mono text-sm ${value ? 'text-white' : 'text-zinc-500'}`}>{value ?? 'Ask your account manager'}</p>
    </div>
  )
}

export default async function ConnectionPage() {
  const serviceProviderId = await getCurrentServiceProviderId()
  if (!serviceProviderId) {
    return (
      <div className="p-8 text-center">
        <p className="text-zinc-400">{NOT_LINKED_MESSAGE}</p>
      </div>
    )
  }

  const credentials = await getCredentials(serviceProviderId)
  const smppLogins = credentials.filter((c) => c.protocol === 'smpp' && c.systemId)
  const apiKeys = credentials.filter((c) => c.protocol === 'http' && c.apiKeyIdentifier)
  const smpp = smppConnectionInfo()
  const httpUrl = httpApiPublicUrl()
  const endpoint = httpUrl ? `${httpUrl}/sms/incoming` : null

  return (
    <div className="p-8 space-y-6 max-w-5xl">
      <PageHeader title="Connection" description="How your systems connect to send SMS" />

      <div className="flex items-start gap-3 p-4 bg-white/5 border border-white/10 rounded-xl">
        <Info className="h-5 w-5 text-zinc-400 mt-0.5 shrink-0" />
        <p className="text-sm text-zinc-400">
          Passwords and API keys are sent to you once when they are created and are not shown here. If you have lost
          one, or need a new login or key, contact your account manager.
        </p>
      </div>

      <DataCard title="SMPP" subtitle="SMPP 3.4, transceiver, transmitter or receiver binds">
        <div className="grid gap-4 sm:grid-cols-2 mb-6">
          <Detail label="Host" value={smpp.host} />
          <Detail label="Port" value={smpp.port} />
        </div>
        {smppLogins.length === 0 ? (
          <p className="text-sm text-zinc-500">No SMPP logins yet.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-zinc-400 border-b border-white/10">
                  <th className="py-2 pr-4 font-medium">System ID</th>
                  <th className="py-2 pr-4 font-medium">Bind type</th>
                  <th className="py-2 pr-4 font-medium">Status</th>
                  <th className="py-2 font-medium">Created</th>
                </tr>
              </thead>
              <tbody>
                {smppLogins.map((c) => (
                  <tr key={c.id} className="border-b border-white/5">
                    <td className="py-2 pr-4 font-mono text-white">{c.systemId}</td>
                    <td className="py-2 pr-4 text-zinc-300">{BIND_TYPES[c.bindType ?? ''] ?? c.bindType ?? '-'}</td>
                    <td className="py-2 pr-4"><CredentialStatus status={c.status} /></td>
                    <td className="py-2 text-zinc-400">{c.createdAt.toISOString().slice(0, 10)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <p className="text-xs text-zinc-500 mt-4">
          Up to 10 sessions per login. Delivery reports are sent to receiver and transceiver sessions.
        </p>
      </DataCard>

      <DataCard title="HTTP API" subtitle="Send one message per request">
        <div className="mb-6">
          <Detail label="Endpoint (POST)" value={endpoint} />
        </div>
        {apiKeys.length === 0 ? (
          <p className="text-sm text-zinc-500 mb-6">No API keys yet.</p>
        ) : (
          <div className="overflow-x-auto mb-6">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-zinc-400 border-b border-white/10">
                  <th className="py-2 pr-4 font-medium">API key</th>
                  <th className="py-2 pr-4 font-medium">Status</th>
                  <th className="py-2 font-medium">Created</th>
                </tr>
              </thead>
              <tbody>
                {apiKeys.map((c) => (
                  <tr key={c.id} className="border-b border-white/5">
                    <td className="py-2 pr-4 font-mono text-white">{c.apiKeyIdentifier!.slice(0, VISIBLE_KEY_PREFIX)}…</td>
                    <td className="py-2 pr-4"><CredentialStatus status={c.status} /></td>
                    <td className="py-2 text-zinc-400">{c.createdAt.toISOString().slice(0, 10)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <p className="text-sm text-zinc-300 mb-2">Example request</p>
        <pre className="p-4 bg-black/50 rounded-xl text-xs text-zinc-300 overflow-x-auto">{`curl -X POST ${endpoint ?? '<endpoint>'} \\
  -H "Authorization: Bearer <your API key>" \\
  -H "Content-Type: application/json" \\
  -d '{"from": "YourSender", "to": "2348031234567", "text": "Hello", "client_ref": "order-1001", "request_dlr": true}'`}</pre>
        <p className="text-sm text-zinc-300 mt-4 mb-2">Responses</p>
        <ul className="text-sm text-zinc-400 space-y-1">
          <li><span className="font-mono text-zinc-300">202</span> {'{"status": "accepted", "message_id": "..."}'}</li>
          <li><span className="font-mono text-zinc-300">400</span> {'{"status": "rejected", "error": "..."}'}, the error says why</li>
          <li><span className="font-mono text-zinc-300">402</span> insufficient wallet balance</li>
          <li><span className="font-mono text-zinc-300">401</span> missing, wrong or inactive API key</li>
        </ul>
      </DataCard>
    </div>
  )
}
