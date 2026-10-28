'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Save, Network, Loader2, Globe, Server, Clock, AlertCircle, CheckCircle } from 'lucide-react'
import { JsonEditor } from '@/components/ui/json-editor'

export default function NewConnectionPage() {
  const params = useParams()
  const router = useRouter()
  const mnoId = params.id as string

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
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

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target
    if (type === 'checkbox') {
      setFormData(prev => ({ ...prev, [name]: (e.target as HTMLInputElement).checked }))
    } else {
      setFormData(prev => ({ ...prev, [name]: value }))
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const payload: Record<string, any> = {
        mnoId: parseInt(mnoId),
        protocol: formData.protocol,
        status: formData.status,
        priority: parseInt(formData.priority),
      }

      if (formData.protocol === 'smpp') {
        payload.systemId = formData.systemId
        payload.password = formData.password
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

      const response = await fetch('/api/admin/mno-connections', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })

      if (response.ok) {
        router.push(`/admin/mnos/${mnoId}`)
      } else {
        const data = await response.json()
        setError(data.error || 'Failed to create connection')
      }
    } catch (err) {
      setError('An error occurred')
    } finally {
      setLoading(false)
    }
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
        <h1 className="text-3xl font-bold text-white">New MNO Connection</h1>
        <p className="text-zinc-400 mt-1">Create a new connection for this MNO</p>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 flex items-center gap-2">
          <AlertCircle className="h-5 w-5" />
          {error}
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
              <label htmlFor="protocol" className="block text-sm font-medium text-zinc-400 mb-1">Protocol *</label>
              <select
                id="protocol"
                name="protocol"
                value={formData.protocol}
                onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              >
                <option value="smpp">SMPP</option>
                <option value="http">HTTP</option>
              </select>
            </div>

            <div>
              <label htmlFor="status" className="block text-sm font-medium text-zinc-400 mb-1">Status</label>
              <select
                id="status"
                name="status"
                value={formData.status}
                onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>

            <div>
              <label htmlFor="priority" className="block text-sm font-medium text-zinc-400 mb-1">Priority</label>
              <input
                type="number"
                id="priority"
                name="priority"
                value={formData.priority}
                onChange={handleChange}
                min="1"
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
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
                  <label htmlFor="systemId" className="block text-sm font-medium text-zinc-400 mb-1">System ID *</label>
                  <input
                    type="text"
                    id="systemId"
                    name="systemId"
                    value={formData.systemId}
                    onChange={handleChange}
                    placeholder="e.g., 30020D1A8"
                    required
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
                  />
                </div>

                <div>
                  <label htmlFor="password" className="block text-sm font-medium text-zinc-400 mb-1">Password *</label>
                  <input
                    type="password"
                    id="password"
                    name="password"
                    value={formData.password}
                    onChange={handleChange}
                    placeholder="Enter password"
                    required
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
                  />
                </div>

                <div>
                  <label htmlFor="host" className="block text-sm font-medium text-zinc-400 mb-1">Host *</label>
                  <input
                    type="text"
                    id="host"
                    name="host"
                    value={formData.host}
                    onChange={handleChange}
                    placeholder="e.g., 196.13.112.164"
                    required
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
                  />
                </div>

                <div>
                  <label htmlFor="port" className="block text-sm font-medium text-zinc-400 mb-1">Port</label>
                  <input
                    type="number"
                    id="port"
                    name="port"
                    value={formData.port}
                    onChange={handleChange}
                    placeholder="e.g., 6000"
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
                  />
                </div>

                <div>
                  <label htmlFor="bindType" className="block text-sm font-medium text-zinc-400 mb-1">Bind Type</label>
                  <select
                    id="bindType"
                    name="bindType"
                    value={formData.bindType}
                    onChange={handleChange}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
                  >
                    <option value="trx">Transceiver (TRX)</option>
                    <option value="tx">Transmitter (TX)</option>
                    <option value="rx">Receiver (RX)</option>
                  </select>
                </div>

                <div>
                  <label htmlFor="systemType" className="block text-sm font-medium text-zinc-400 mb-1">System Type</label>
                  <input
                    type="text"
                    id="systemType"
                    name="systemType"
                    value={formData.systemType}
                    onChange={handleChange}
                    placeholder="e.g., 54"
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
                  />
                </div>
              </div>
            </div>

            <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
              <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                <Clock className="h-5 w-5" />
                SMPP Timing & Performance
              </h2>

              <div className="grid gap-6 md:grid-cols-2">
                <div>
                  <label htmlFor="enquireLinkIntervalSecs" className="block text-sm font-medium text-zinc-400 mb-1">Enquire Link Interval (sec)</label>
                  <input
                    type="number"
                    id="enquireLinkIntervalSecs"
                    name="enquireLinkIntervalSecs"
                    value={formData.enquireLinkIntervalSecs}
                    onChange={handleChange}
                    min="1"
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
                  />
                </div>

                <div>
                  <label htmlFor="requestTimeoutSecs" className="block text-sm font-medium text-zinc-400 mb-1">Request Timeout (sec)</label>
                  <input
                    type="number"
                    id="requestTimeoutSecs"
                    name="requestTimeoutSecs"
                    value={formData.requestTimeoutSecs}
                    onChange={handleChange}
                    min="1"
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
                  />
                </div>

                <div>
                  <label htmlFor="connectRetryDelaySecs" className="block text-sm font-medium text-zinc-400 mb-1">Retry Delay (sec)</label>
                  <input
                    type="number"
                    id="connectRetryDelaySecs"
                    name="connectRetryDelaySecs"
                    value={formData.connectRetryDelaySecs}
                    onChange={handleChange}
                    min="1"
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
                  />
                </div>

                <div>
                  <label htmlFor="maxWindowSize" className="block text-sm font-medium text-zinc-400 mb-1">Max Window Size</label>
                  <input
                    type="number"
                    id="maxWindowSize"
                    name="maxWindowSize"
                    value={formData.maxWindowSize}
                    onChange={handleChange}
                    min="1"
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
                  />
                </div>

                <div>
                  <label htmlFor="defaultDataCoding" className="block text-sm font-medium text-zinc-400 mb-1">Default Data Coding</label>
                  <input
                    type="number"
                    id="defaultDataCoding"
                    name="defaultDataCoding"
                    value={formData.defaultDataCoding}
                    onChange={handleChange}
                    min="0"
                    max="15"
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
                  />
                </div>

                <div>
                  <label htmlFor="sourceAddrTon" className="block text-sm font-medium text-zinc-400 mb-1">Address TON/NPI</label>
                  <div className="grid grid-cols-2 gap-2">
                    <input
                      type="number"
                      id="sourceAddrTon"
                      name="sourceAddrTon"
                      value={formData.sourceAddrTon}
                      onChange={handleChange}
                      placeholder="TON"
                      className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
                    />
                    <input
                      type="number"
                      id="sourceAddrNpi"
                      name="sourceAddrNpi"
                      value={formData.sourceAddrNpi}
                      onChange={handleChange}
                      placeholder="NPI"
                      className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
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
                <label htmlFor="endpointUrl" className="block text-sm font-medium text-zinc-400 mb-1">Endpoint URL *</label>
                <input
                  type="url"
                  id="endpointUrl"
                  name="endpointUrl"
                  value={formData.endpointUrl}
                  onChange={handleChange}
                  placeholder="http://52.10.34.217/aegisbox/send"
                  required
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
                />
              </div>

              <div className="md:col-span-2">
                <label htmlFor="callbackUrl" className="block text-sm font-medium text-zinc-400 mb-1">Callback URL</label>
                <input
                  type="url"
                  id="callbackUrl"
                  name="callbackUrl"
                  value={formData.callbackUrl}
                  onChange={handleChange}
                  placeholder="http://localhost:9001/aegis-dlr/callback?connID=[connID]&msgID=[msgID]"
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
                />
              </div>

              <div className="md:col-span-2">
                <label htmlFor="authToken" className="block text-sm font-medium text-zinc-400 mb-1">Auth Token</label>
                <input
                  type="text"
                  id="authToken"
                  name="authToken"
                  value={formData.authToken}
                  onChange={handleChange}
                  placeholder="Enter authentication token"
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
                />
              </div>

              <div>
                <label htmlFor="httpTimeoutSecs" className="block text-sm font-medium text-zinc-400 mb-1">Timeout (seconds)</label>
                <input
                  type="number"
                  id="httpTimeoutSecs"
                  name="httpTimeoutSecs"
                  value={formData.httpTimeoutSecs}
                  onChange={handleChange}
                  min="1"
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
                />
              </div>

              <div>
                <label htmlFor="maxRetries" className="block text-sm font-medium text-zinc-400 mb-1">Max Retries</label>
                <input
                  type="number"
                  id="maxRetries"
                  name="maxRetries"
                  value={formData.maxRetries}
                  onChange={handleChange}
                  min="0"
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
                />
              </div>

              <div>
                <label htmlFor="backoffSeconds" className="block text-sm font-medium text-zinc-400 mb-1">Backoff (seconds)</label>
                <input
                  type="number"
                  id="backoffSeconds"
                  name="backoffSeconds"
                  value={formData.backoffSeconds}
                  onChange={handleChange}
                  min="1"
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
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
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-50"
          >
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {loading ? 'Creating...' : 'Create Connection'}
          </button>
        </div>
      </form>
    </div>
  )
}
