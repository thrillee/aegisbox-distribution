import { Plus, Copy, Check, X, Key, Calendar, Eye, EyeOff, RefreshCw } from 'lucide-react'

async function getApiKeys() {
  return [
    { id: '1', name: 'Production API', prefix: 'aegisbox_live_', created: '2024-01-15', lastUsed: '2024-06-15 10:30:00', status: 'active', rateLimit: 100 },
    { id: '2', name: 'Test API', prefix: 'aegisbox_test_', created: '2024-02-20', lastUsed: '2024-05-10 14:20:00', status: 'active', rateLimit: 50 },
    { id: '3', name: 'Development', prefix: 'aegisbox_dev_', created: '2024-03-01', lastUsed: '2024-06-14 09:15:00', status: 'revoked', rateLimit: 10 },
  ]
}

function getStatusBadge(status: string) {
  const styles: Record<string, { bg: string; text: string }> = {
    active: { bg: 'bg-emerald-500/20 text-emerald-400', text: 'Active' },
    revoked: { bg: 'bg-red-500/20 text-red-400', text: 'Revoked' },
  }

  const style = styles[status] || styles.revoked

  return (
    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${style.bg}`}>
      {style.text}
    </span>
  )
}

export default async function ApiKeysPage() {
  const apiKeys = await getApiKeys()

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">API Keys</h1>
          <p className="text-zinc-400 mt-1">
            Manage your API keys for programmatic access
          </p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors">
          <Plus className="h-4 w-4" />
          Create New Key
        </button>
      </div>

      <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
        <div className="flex items-start gap-4 p-4 bg-white/5 border border-white/10 rounded-xl mb-6">
          <Key className="h-5 w-5 text-zinc-400 mt-0.5" />
          <div>
            <p className="text-sm font-medium text-white">
              Keep your API keys secure
            </p>
            <p className="text-sm text-zinc-400 mt-1">
              Never share your API keys publicly or commit them to version control.
              Regenerate keys immediately if compromised.
            </p>
          </div>
        </div>

        <div className="space-y-4">
          {apiKeys.map((key) => (
            <div key={key.id} className="p-4 border border-white/10 rounded-xl">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <span className="font-medium text-white">{key.name}</span>
                  {getStatusBadge(key.status)}
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-zinc-500">{key.rateLimit}/min</span>
                  {key.status === 'active' && (
                    <button className="p-2 hover:bg-white/10 rounded-lg transition-colors">
                      <RefreshCw className="h-4 w-4 text-zinc-400" />
                    </button>
                  )}
                  {key.status === 'active' ? (
                    <button className="px-3 py-1.5 text-sm font-medium text-red-400 hover:bg-red-500/10 rounded-lg transition-colors">
                      Revoke
                    </button>
                  ) : (
                    <button className="px-3 py-1.5 text-sm font-medium text-emerald-400 hover:bg-emerald-500/10 rounded-lg transition-colors">
                      Regenerate
                    </button>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-2 p-3 bg-black/50 rounded-xl">
                <code className="flex-1 font-mono text-sm text-zinc-400">
                  {key.prefix}abc123xyz789def456ghi789jkl012
                </code>
                <button className="p-2 hover:bg-white/10 rounded-lg transition-colors">
                  <Copy className="h-4 w-4 text-zinc-400" />
                </button>
              </div>

              <div className="flex items-center justify-between mt-3 text-sm text-zinc-500">
                <div className="flex items-center gap-4">
                  <span className="flex items-center gap-1">
                    <Calendar className="h-4 w-4" />
                    Created: {new Date(key.created).toLocaleDateString()}
                  </span>
                  <span>Last used: {new Date(key.lastUsed).toLocaleString()}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
        <h3 className="font-semibold text-white mb-4">API Documentation</h3>
        <p className="text-sm text-zinc-400 mb-4">
          Learn how to integrate our SMS API into your applications.
        </p>
        <div className="grid gap-4 md:grid-cols-3">
          <a href="#" className="p-4 border border-white/10 rounded-xl hover:bg-white/5 transition-colors">
            <p className="font-medium text-white">Quick Start Guide</p>
            <p className="text-sm text-zinc-500 mt-1">Get started with your first API call</p>
          </a>
          <a href="#" className="p-4 border border-white/10 rounded-xl hover:bg-white/5 transition-colors">
            <p className="font-medium text-white">API Reference</p>
            <p className="text-sm text-zinc-500 mt-1">Complete endpoint documentation</p>
          </a>
          <a href="#" className="p-4 border border-white/10 rounded-xl hover:bg-white/5 transition-colors">
            <p className="font-medium text-white">Code Examples</p>
            <p className="text-sm text-zinc-500 mt-1">Sample code in popular languages</p>
          </a>
        </div>
      </div>
    </div>
  )
}
