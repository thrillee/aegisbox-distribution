'use client'

import { KeyRound } from 'lucide-react'

export interface SmppConnection {
  host: string | null
  port: string | null
}

// Shows a new SMPP login once. Only the password's hash is stored, so this is the only time it
// can be read.
export function SmppLoginCreated({
  systemId,
  password,
  connection,
  onDone,
}: Readonly<{ systemId: string; password: string; connection: SmppConnection | null; onDone: () => void }>) {
  const rows: [string, string][] = [
    ['Host', connection?.host ?? 'Not set (SMPP_PUBLIC_HOST)'],
    ['Port', connection?.port ?? 'Not set (SMPP_PUBLIC_PORT)'],
    ['System ID', systemId],
    ['Password', password],
  ]
  return (
    <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6 space-y-4">
      <h2 className="font-semibold text-white flex items-center gap-2">
        <KeyRound className="h-4 w-4" />
        SMPP login created
      </h2>
      <p className="text-sm text-amber-400">
        Copy the password now and send it to the customer. It is stored hashed and cannot be shown again. The login
        starts as pending: activate it on the customer page before they bind.
      </p>
      <dl className="grid grid-cols-[auto_1fr] gap-x-6 gap-y-2 text-sm">
        {rows.map(([label, value]) => (
          <div key={label} className="contents">
            <dt className="text-zinc-400">{label}</dt>
            <dd className="text-white font-mono break-all">{value}</dd>
          </div>
        ))}
      </dl>
      <button
        type="button"
        onClick={onDone}
        className="px-4 py-2 bg-white text-black rounded-xl hover:bg-zinc-200 transition-colors"
      >
        Done
      </button>
    </div>
  )
}
