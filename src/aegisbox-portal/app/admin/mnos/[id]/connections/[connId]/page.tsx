'use client'

import { use, useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Save, Network, Loader2, Globe, Server, Clock, AlertCircle, CheckCircle } from 'lucide-react'
import { JsonEditor } from '@/components/ui/json-editor'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

interface Connection {
  id: number
  mnoId: number
  protocol: string
  status: string
  priority: number
  systemId: string | null
  host: string | null
  port: number | null
  useTls: boolean
  bindType: string | null
  systemType: string | null
  enquireLinkIntervalSecs: number
  requestTimeoutSecs: number
  connectRetryDelaySecs: number
  maxWindowSize: number
  defaultDataCoding: number
  sourceAddrTon: number
  sourceAddrNpi: number
  destAddrTon: number
  destAddrNpi: number
  baseUrl: string
  timeoutSecs: number
  providerConfig: Record<string, any> | null
  createdAt: string
}

export default function ConnectionDetailPage() {
  const params = useParams()
  const router = useRouter()
  const mnoId = params.id as string
  const connId = params.connId as string

  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [connection, setConnection] = useState<Connection | null>(null)
  const [providerConfig, setProviderConfig] = useState<Record<string, any> | null>(null)
  const [formData, setFormData] = useState({
    protocol: 'smpp',
    status: 'active',
    priority: '1',
    systemId: '',
    password: '',
    host: '',
    port: '',
    useTls: false,
    bindType: 'trx',
    systemType: '',
    enquireLinkIntervalSecs: '30',
    requestTimeoutSecs: '5',
    connectRetryDelaySecs: '10',
    maxWindowSize: '50',
    defaultDataCoding: '0',
    sourceAddrTon: '1',
    sourceAddrNpi: '1',
    destAddrTon: '1',
    destAddrNpi: '1',
    endpointUrl: '',
    callbackUrl: '',
    authToken: '',
    httpTimeoutSecs: '15',
    maxRetries: '3',
    backoffSeconds: '2',
  })

  useEffect(() => {
    fetchConnection()
  }, [connId])

  const fetchConnection = async () => {
    try {
      const response = await fetch(`/api/admin/mno-connections/${connId}`)
      if (response.ok) {
        const data = await response.json()
        const conn = data.connection
        setConnection(conn)
        setProviderConfig(conn.providerConfig)
        setFormData({
          protocol: conn.protocol,
          status: conn.status,
          priority: String(conn.priority),
          systemId: conn.systemId || '',
          password: '',
          host: conn.host || '',
          port: conn.port ? String(conn.port) : '',
          useTls: conn.useTls || false,
          bindType: conn.bindType || 'trx',
          systemType: conn.systemType || '',
          enquireLinkIntervalSecs: String(conn.enquireLinkIntervalSecs || 30),
          requestTimeoutSecs: String(conn.requestTimeoutSecs || 5),
          connectRetryDelaySecs: String(conn.connectRetryDelaySecs || 10),
          maxWindowSize: String(conn.maxWindowSize || 50),
          defaultDataCoding: String(conn.defaultDataCoding || 0),
          sourceAddrTon: String(conn.sourceAddrTon || 1),
          sourceAddrNpi: String(conn.sourceAddrNpi || 1),
          destAddrTon: String(conn.destAddrTon || 1),
          destAddrNpi: String(conn.destAddrNpi || 1),
          endpointUrl: conn.providerConfig?.endpoint_url || conn.baseUrl || '',
          callbackUrl: conn.providerConfig?.callback_url || '',
          authToken: conn.providerConfig?.auth_token || '',
          httpTimeoutSecs: String(conn.providerConfig?.timeout_secs || conn.timeoutSecs || 15),
          maxRetries: String(conn.providerConfig?.retry_policy?.max_retries || 3),
          backoffSeconds: String(conn.providerConfig?.retry_policy?.backoff_seconds || 2),
        })
      } else {
        setError('Connection not found')
      }
    } catch (err) {
      setError('Failed to fetch connection')
    } finally {
      setLoading(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target
    if (type === 'checkbox') {
      setFormData(prev => ({ ...prev, [name]: (e.target as HTMLInputElement).checked }))
    } else {
      setFormData(prev => ({ ...prev, [name]: value }))
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError('')
    setSuccess('')

    try {
      const payload: Record<string, any> = {
        status: formData.status,
        priority: parseInt(formData.priority),
      }

      if (formData.protocol === 'smpp') {
        payload.systemId = formData.systemId
        if (formData.password) payload.password = formData.password
        payload.host = formData.host
        payload.port = parseInt(formData.port) || null
        payload.useTls = formData.useTls
        payload.bindType = formData.bindType
        payload.systemType = formData.systemType || null
        payload.enquireLinkIntervalSecs = parseInt(formData.enquireLinkIntervalSecs)
        payload.requestTimeoutSecs = parseInt(formData.requestTimeoutSecs)
        payload.connectRetryDelaySecs = parseInt(formData.connectRetryDelaySecs)
        payload.maxWindowSize = parseInt(formData.maxWindowSize)
        payload.defaultDataCoding = parseInt(formData.defaultDataCoding)
        payload.sourceAddrTon = parseInt(formData.sourceAddrTon)
        payload.sourceAddrNpi = parseInt(formData.sourceAddrNpi)
        payload.destAddrTon = parseInt(formData.destAddrTon)
        payload.destAddrNpi = parseInt(formData.destAddrNpi)
      } else {
        payload.httpConfig = {
          endpoint_url: formData.endpointUrl,
          callback_url: formData.callbackUrl,
          auth_token: formData.authToken,
          timeout_secs: parseInt(formData.httpTimeoutSecs),
          retry_policy: {
            max_retries: parseInt(formData.maxRetries),
            backoff_seconds: parseInt(formData.backoffSeconds),
          },
        }
      }

      if (providerConfig && Object.keys(providerConfig).length > 0) {
        payload.providerConfig = providerConfig
      }

      const response = await fetch(`/api/admin/mno-connections/${connId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })

      if (response.ok) {
        setSuccess('Connection updated successfully')
        fetchConnection()
      } else {
        const data = await response.json()
        setError(data.error || 'Failed to update connection')
      }
    } catch (err) {
      setError('An error occurred')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return <PageLoadingSkeleton />
  }

  if (!connection) {
    return (
      <div className="p-8">
        <Link href={`/admin/mnos/${mnoId}`} className="flex items-center gap-2 text-zinc-400">
          <ArrowLeft className="h-4 w-4" />
          Back to MNO
        </Link>
        <div className="mt-8 text-center">
          <h1 className="text-2xl font-bold text-white">Connection Not Found</h1>
        </div>
      </div>
    )
  }

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <div className="mb-6">
        <Link href={`/admin/mnos/${mnoId}`} className="flex items-center gap-2 text-zinc-400 hover:text-white transition-colors">
          <ArrowLeft className="h-4 w-4" />
          Back to MNO
        </Link>
      </div>

      <div className="mb-6">
        <h1 className="text-3xl font-bold text-white">
          Edit Connection
        </h1>
        <p className="text-zinc-400 mt-1">
          {connection.protocol.toUpperCase()} Connection #{connection.id}
        </p>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 flex items-center gap-2">
          <AlertCircle className="h-5 w-5" />
          {error}
        </div>
      )}

      {success && (
        <div className="mb-6 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-emerald-400 flex items-center gap-2">
          <CheckCircle className="h-5 w-5" />
          {success}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <Network className="h-5 w-5" />
            Connection Details
          </h2>

          <div className="grid gap-6 md:grid-cols-3">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Protocol</label>
              <div className="px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white capitalize">
                {connection.protocol}
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Status</label>
              <select
                name="status"
                value={formData.status}
                onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Priority</label>
              <input
                type="number"
                name="priority"
                value={formData.priority}
                onChange={handleChange}
                min="1"
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
              />
            </div>
          </div>
        </div>

        {formData.protocol === 'smpp' ? (
          <>
            <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
              <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                <Server className="h-5 w-5" />
                SMPP Configuration
              </h2>

              <div className="grid gap-6 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">System ID</label>
                  <input
                    type="text"
                    name="systemId"
                    value={formData.systemId}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">Password (leave empty to keep current)</label>
                  <input
                    type="password"
                    name="password"
                    value={formData.password}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">Host</label>
                  <input
                    type="text"
                    name="host"
                    value={formData.host}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">Port</label>
                  <input
                    type="number"
                    name="port"
                    value={formData.port}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">Bind Type</label>
                  <select
                    name="bindType"
                    value={formData.bindType}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  >
                    <option value="trx">Transceiver (TRX)</option>
                    <option value="tx">Transmitter (TX)</option>
                    <option value="rx">Receiver (RX)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">System Type</label>
                  <input
                    type="text"
                    name="systemType"
                    value={formData.systemType}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  />
                </div>
              </div>
            </div>

            <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
              <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                <Clock className="h-5 w-5" />
                SMPP Timing & Performance
              </h2>

              <div className="grid gap-6 md:grid-cols-3">
                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">Enquire Link (sec)</label>
                  <input
                    type="number"
                    name="enquireLinkIntervalSecs"
                    value={formData.enquireLinkIntervalSecs}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">Request Timeout (sec)</label>
                  <input
                    type="number"
                    name="requestTimeoutSecs"
                    value={formData.requestTimeoutSecs}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">Retry Delay (sec)</label>
                  <input
                    type="number"
                    name="connectRetryDelaySecs"
                    value={formData.connectRetryDelaySecs}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">Max Window Size</label>
                  <input
                    type="number"
                    name="maxWindowSize"
                    value={formData.maxWindowSize}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">Data Coding</label>
                  <input
                    type="number"
                    name="defaultDataCoding"
                    value={formData.defaultDataCoding}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">Address TON/NPI</label>
                  <div className="grid grid-cols-2 gap-2">
                    <input
                      type="number"
                      name="sourceAddrTon"
                      value={formData.sourceAddrTon}
                      onChange={handleChange}
                      placeholder="TON"
                      className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                    />
                    <input
                      type="number"
                      name="sourceAddrNpi"
                      value={formData.sourceAddrNpi}
                      onChange={handleChange}
                      placeholder="NPI"
                      className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                    />
                  </div>
                </div>
              </div>
            </div>
          </>
        ) : (
          <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
            <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
              <Globe className="h-5 w-5" />
              HTTP Configuration
            </h2>

            <div className="grid gap-6 md:grid-cols-2">
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-zinc-400 mb-1">Endpoint URL</label>
                <input
                  type="url"
                  name="endpointUrl"
                  value={formData.endpointUrl}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                />
              </div>

              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-zinc-400 mb-1">Callback URL</label>
                <input
                  type="url"
                  name="callbackUrl"
                  value={formData.callbackUrl}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                />
              </div>

              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-zinc-400 mb-1">Auth Token</label>
                <input
                  type="text"
                  name="authToken"
                  value={formData.authToken}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-400 mb-1">Timeout (sec)</label>
                <input
                  type="number"
                  name="httpTimeoutSecs"
                  value={formData.httpTimeoutSecs}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-400 mb-1">Max Retries</label>
                <input
                  type="number"
                  name="maxRetries"
                  value={formData.maxRetries}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-400 mb-1">Backoff (sec)</label>
                <input
                  type="number"
                  name="backoffSeconds"
                  value={formData.backoffSeconds}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                />
              </div>
            </div>
          </div>
        )}

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <JsonEditor
            value={providerConfig}
            onChange={setProviderConfig}
            label="Provider Config (JSON)"
            description="Additional provider-specific configuration for both SMPP and HTTP connections"
          />
        </div>

        <div className="flex items-center justify-end gap-3">
          <Link
            href={`/admin/mnos/${mnoId}`}
            className="px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition-colors"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={saving}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-50"
          >
            {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  )
}
