'use client'

import { useState, useEffect } from 'react'
import { Loader2, ToggleLeft, ToggleRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DLR_STATUSES, FAILURE_DLR_STATUSES } from '@/components/messages/manual-dlr-form'

interface AutoDlrConfig {
  enabled: boolean
  status: string
  delay_seconds: number
  enabled_at: string | null
}

const MIN_DELAY_MINUTES = 1
const MAX_DELAY_MINUTES = 2880

export function AutoDlrSettings({ providerId }: Readonly<{ providerId: number }>) {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [enabled, setEnabled] = useState(false)
  const [status, setStatus] = useState('DELIVRD')
  const [delayMinutes, setDelayMinutes] = useState('5')
  const [enabledAt, setEnabledAt] = useState<string | null>(null)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [savedMessage, setSavedMessage] = useState<string | null>(null)

  const applyConfig = (config: AutoDlrConfig) => {
    setEnabled(config.enabled)
    setStatus(config.status)
    setDelayMinutes(String(Math.round(config.delay_seconds / 60)))
    setEnabledAt(config.enabled_at)
  }

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const res = await fetch(`/api/admin/service-providers/${providerId}/auto-dlr`)
        const data = await res.json().catch(() => ({}))
        if (res.ok) {
          applyConfig(data)
        } else {
          setErrorMessage(data.error || 'Failed to load auto DLR settings')
        }
      } catch {
        setErrorMessage('Failed to load auto DLR settings')
      } finally {
        setLoading(false)
      }
    }
    fetchConfig()
  }, [providerId])

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    setErrorMessage(null)
    setSavedMessage(null)

    const minutes = Number(delayMinutes)
    if (!Number.isInteger(minutes) || minutes < MIN_DELAY_MINUTES || minutes > MAX_DELAY_MINUTES) {
      setErrorMessage(`Delay must be a whole number of minutes between ${MIN_DELAY_MINUTES} and ${MAX_DELAY_MINUTES}`)
      return
    }

    setSaving(true)
    try {
      const res = await fetch(`/api/admin/service-providers/${providerId}/auto-dlr`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled, status, delay_seconds: minutes * 60 }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setErrorMessage(data.error || 'Failed to save auto DLR settings')
        return
      }
      applyConfig(data)
      setSavedMessage('Auto DLR settings saved')
    } catch {
      setErrorMessage('Failed to save auto DLR settings')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-zinc-500" />
      </div>
    )
  }

  return (
    <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
      <h3 className="text-lg font-semibold text-white">Auto DLR</h3>
      <p className="text-sm text-zinc-400 mt-0.5 mb-6">
        When the MNO does not return a DLR within the delay, the gateway sets the chosen status and forwards the DLR to this provider. Only messages sent after auto DLR is turned on are affected. Use Send DLR on the Messages page for older messages.
      </p>

      <form onSubmit={handleSave} className="space-y-5">
        <div className="flex items-center justify-between rounded-lg border border-white/10 px-4 py-3">
          <div>
            <p className="font-medium text-white">Enable auto DLR</p>
            <p className="text-sm text-zinc-500">
              {enabled && enabledAt
                ? `On since ${new Date(enabledAt).toLocaleString()}`
                : 'Off: messages wait for a real DLR from the MNO'}
            </p>
          </div>
          <button
            type="button"
            role="switch"
            aria-checked={enabled}
            aria-label="Enable auto DLR"
            onClick={() => setEnabled(!enabled)}
            className="text-zinc-400 hover:text-white"
          >
            {enabled ? <ToggleRight className="h-8 w-8 text-emerald-400" /> : <ToggleLeft className="h-8 w-8" />}
          </button>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-1">DLR status to send</label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
            >
              {DLR_STATUSES.map((s) => (
                <option key={s.value} value={s.value}>{s.label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-1">Wait for MNO DLR (minutes)</label>
            <input
              type="number"
              min={MIN_DELAY_MINUTES}
              max={MAX_DELAY_MINUTES}
              step={1}
              value={delayMinutes}
              onChange={(e) => setDelayMinutes(e.target.value)}
              className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
            />
          </div>
        </div>

        {FAILURE_DLR_STATUSES.includes(status) && (
          <p className="text-sm text-amber-400">
            A failure status marks messages as failed. Charges are not refunded automatically.
          </p>
        )}
        {errorMessage && <p className="text-sm text-red-400">{errorMessage}</p>}
        {savedMessage && <p className="text-sm text-emerald-400">{savedMessage}</p>}

        <Button type="submit" disabled={saving} className="bg-white text-black hover:bg-zinc-200">
          {saving && <Loader2 className="h-4 w-4 animate-spin" />}
          {saving ? 'Saving...' : 'Save'}
        </Button>
      </form>
    </div>
  )
}
