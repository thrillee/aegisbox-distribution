'use client'

import * as React from 'react'
import { useState, useEffect, useCallback } from 'react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'
import { Settings2, Save, Loader2, Plus, Pencil, X } from 'lucide-react'

interface GroupConfig {
  id: number
  routing_group_id: number
  routing_group_name: string
  selection_strategy: string
  max_route_retries: number
  route_failure_threshold: number
}

interface GroupConfigForm {
  selection_strategy: string
  max_route_retries: number
  route_failure_threshold: number
}

interface WorkerSchedule {
  dlr_timeout_check_interval_seconds: number
  sender_auto_unblock_check_interval_minutes: number
  sender_auto_unblock_after_inactive_days: number
  ring_buffer_cleanup_interval_hours: number
  ring_buffer_retention_days: number
}

interface GlobalConfig {
  selection_strategy: string
  max_route_retries: number
  route_failure_threshold: number
  route_recovery_threshold: number
  route_health_window_size: number
  sender_block_threshold: number
  sender_unblock_threshold: number
  sender_health_window_size: number
  dlr_timeout_seconds: number
  schedule?: WorkerSchedule
}

interface RoutingGroup {
  id: number
  name: string
}

const STRATEGIES = [
  { value: 'priority_weighted', label: 'Priority, then weight' },
  { value: 'health_based', label: 'Best health first' },
  { value: 'round_robin', label: 'Round robin' },
]

const INPUT_CLASS =
  'w-full px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-white disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-blue-500'

async function readError(response: Response, fallback: string): Promise<string> {
  const data = await response.json().catch(() => ({}))
  return data.error || fallback
}

function NumberField({
  label,
  hint,
  value,
  onChange,
  disabled,
  step = 1,
  min,
  max,
}: Readonly<{
  label: string
  hint?: string
  value: number
  onChange: (value: number) => void
  disabled?: boolean
  step?: number
  min?: number
  max?: number
}>) {
  const id = React.useId()
  return (
    <div>
      <label htmlFor={id} className="block text-sm font-medium text-zinc-400 mb-1">{label}</label>
      <input
        id={id}
        type="number"
        value={Number.isNaN(value) ? '' : value}
        step={step}
        min={min}
        max={max}
        onChange={(e) => onChange(step < 1 ? parseFloat(e.target.value) : parseInt(e.target.value, 10))}
        disabled={disabled}
        className={INPUT_CLASS}
      />
      {hint && <p className="text-xs text-zinc-500 mt-1">{hint}</p>}
    </div>
  )
}

function StrategyField({
  value,
  onChange,
  disabled,
}: Readonly<{ value: string; onChange: (value: string) => void; disabled?: boolean }>) {
  const id = React.useId()
  return (
    <div>
      <label htmlFor={id} className="block text-sm font-medium text-zinc-400 mb-1">Selection Strategy</label>
      <select id={id} value={value} onChange={(e) => onChange(e.target.value)} disabled={disabled} className={INPUT_CLASS}>
        {STRATEGIES.map((s) => (
          <option key={s.value} value={s.value} className="bg-zinc-900">
            {s.label}
          </option>
        ))}
      </select>
    </div>
  )
}

export default function RoutingConfigPage() {
  const [configs, setConfigs] = useState<GroupConfig[]>([])
  const [groups, setGroups] = useState<RoutingGroup[]>([])
  const [globalConfig, setGlobalConfig] = useState<GlobalConfig | null>(null)
  const [savedGlobal, setSavedGlobal] = useState<GlobalConfig | null>(null)
  const [editingGlobal, setEditingGlobal] = useState(false)
  const [editingGroupId, setEditingGroupId] = useState<number | null>(null)
  const [groupForm, setGroupForm] = useState<GroupConfigForm | null>(null)
  const [newGroupId, setNewGroupId] = useState<string>('')
  const [saving, setSaving] = useState(false)
  const [initialLoading, setInitialLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  const loadConfigs = useCallback(async () => {
    const response = await fetch('/api/admin/routing-config')
    if (!response.ok) {
      setErrorMessage(await readError(response, 'Failed to load routing group configs'))
      return
    }
    const data = await response.json()
    setConfigs(data.configs || [])
  }, [])

  const loadGlobalConfig = useCallback(async () => {
    const response = await fetch('/api/admin/routing-config/global')
    if (!response.ok) {
      setErrorMessage(await readError(response, 'Failed to load global routing config'))
      return
    }
    const data = await response.json()
    setGlobalConfig(data)
    setSavedGlobal(data)
  }, [])

  const loadGroups = useCallback(async () => {
    const response = await fetch('/api/admin/routing-groups')
    if (response.ok) {
      const data = await response.json()
      setGroups(data.routingGroups || [])
    }
  }, [])

  useEffect(() => {
    const loadAll = async () => {
      try {
        await Promise.all([loadConfigs(), loadGlobalConfig(), loadGroups()])
      } catch {
        setErrorMessage('Failed to load routing configuration')
      } finally {
        setInitialLoading(false)
      }
    }
    loadAll()
  }, [loadConfigs, loadGlobalConfig, loadGroups])

  const saveGlobal = async () => {
    if (!globalConfig) return
    if (globalConfig.sender_unblock_threshold < globalConfig.sender_block_threshold) {
      setErrorMessage('Sender unblock threshold must be at or above the block threshold')
      return
    }
    setSaving(true)
    setErrorMessage(null)
    setNotice(null)
    try {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { schedule, ...body } = globalConfig
      const response = await fetch('/api/admin/routing-config/global', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!response.ok) {
        setErrorMessage(await readError(response, 'Failed to save global routing config'))
        return
      }
      const data = await response.json()
      setGlobalConfig(data)
      setSavedGlobal(data)
      setEditingGlobal(false)
      setNotice('Global routing config saved. The gateway uses it for the next messages.')
    } catch {
      setErrorMessage('Failed to save global routing config')
    } finally {
      setSaving(false)
    }
  }

  const startGroupEdit = (groupId: number, initial: GroupConfigForm) => {
    setEditingGroupId(groupId)
    setGroupForm(initial)
    setErrorMessage(null)
    setNotice(null)
  }

  const saveGroup = async () => {
    if (editingGroupId === null || !groupForm) return
    setSaving(true)
    setErrorMessage(null)
    setNotice(null)
    try {
      const response = await fetch(`/api/admin/routing-config/${editingGroupId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(groupForm),
      })
      if (!response.ok) {
        setErrorMessage(await readError(response, 'Failed to save routing group config'))
        return
      }
      setEditingGroupId(null)
      setGroupForm(null)
      setNewGroupId('')
      setNotice('Routing group config saved.')
      await loadConfigs()
    } catch {
      setErrorMessage('Failed to save routing group config')
    } finally {
      setSaving(false)
    }
  }

  if (initialLoading) {
    return <PageLoadingSkeleton />
  }

  const groupsWithoutOverride = groups.filter((g) => !configs.some((c) => c.routing_group_id === g.id))
  const defaultsFromGlobal: GroupConfigForm = {
    selection_strategy: savedGlobal?.selection_strategy || 'priority_weighted',
    max_route_retries: savedGlobal?.max_route_retries ?? 2,
    route_failure_threshold: savedGlobal?.route_failure_threshold ?? 5,
  }
  const addingNew = editingGroupId !== null && !configs.some((c) => c.routing_group_id === editingGroupId)

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Routing Configuration"
        description="Route selection, retries and the health thresholds the gateway uses"
      />

      {errorMessage && (
        <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-sm">{errorMessage}</div>
      )}
      {notice && (
        <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-emerald-400 text-sm">{notice}</div>
      )}

      {globalConfig && (
        <DataCard title="Global Settings" subtitle="Defaults for every routing group, plus the sender and route health rules">
          <div className="flex justify-end mb-4 gap-2">
            {editingGlobal ? (
              <>
                <button
                  onClick={() => {
                    setGlobalConfig(savedGlobal)
                    setEditingGlobal(false)
                  }}
                  disabled={saving}
                  className="flex items-center gap-2 px-3 py-1.5 text-sm text-zinc-400 hover:text-white"
                >
                  <X className="h-4 w-4" /> Cancel
                </button>
                <button
                  onClick={saveGlobal}
                  disabled={saving}
                  className="flex items-center gap-2 px-3 py-1.5 text-sm bg-white text-black rounded-lg hover:bg-zinc-200 disabled:opacity-50"
                >
                  {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />} Save
                </button>
              </>
            ) : (
              <button
                onClick={() => {
                  setEditingGlobal(true)
                  setErrorMessage(null)
                  setNotice(null)
                }}
                className="flex items-center gap-2 px-3 py-1.5 text-sm bg-white/5 border border-white/10 rounded-lg text-white hover:bg-white/10"
              >
                <Pencil className="h-4 w-4" /> Edit
              </button>
            )}
          </div>

          <h4 className="text-sm font-semibold text-zinc-300 mb-3">Route selection (default for groups without an override)</h4>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <StrategyField
              value={globalConfig.selection_strategy}
              onChange={(v) => setGlobalConfig({ ...globalConfig, selection_strategy: v })}
              disabled={!editingGlobal}
            />
            <NumberField
              label="Max Route Retries"
              hint="Alternate routes tried after a failed send (0 to 5)"
              value={globalConfig.max_route_retries}
              min={0}
              max={5}
              onChange={(v) => setGlobalConfig({ ...globalConfig, max_route_retries: v })}
              disabled={!editingGlobal}
            />
            <NumberField
              label="Route Failure Threshold"
              hint="Consecutive failures before a route is marked unavailable"
              value={globalConfig.route_failure_threshold}
              min={1}
              max={20}
              onChange={(v) => setGlobalConfig({ ...globalConfig, route_failure_threshold: v })}
              disabled={!editingGlobal}
            />
          </div>

          <h4 className="text-sm font-semibold text-zinc-300 mb-3">Route health</h4>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <NumberField
              label="Route Recovery Threshold"
              hint="Consecutive successes before a route is available again"
              value={globalConfig.route_recovery_threshold}
              min={1}
              max={20}
              onChange={(v) => setGlobalConfig({ ...globalConfig, route_recovery_threshold: v })}
              disabled={!editingGlobal}
            />
            <NumberField
              label="Route Health Window"
              hint="Recent attempts used to score a route (100 to 10000)"
              value={globalConfig.route_health_window_size}
              min={100}
              max={10000}
              onChange={(v) => setGlobalConfig({ ...globalConfig, route_health_window_size: v })}
              disabled={!editingGlobal}
            />
          </div>

          <h4 className="text-sm font-semibold text-zinc-300 mb-3">Sender ID health</h4>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <NumberField
              label="Block Threshold (%)"
              hint="Block a sender when its delivery rate falls below this"
              value={globalConfig.sender_block_threshold}
              step={0.01}
              min={0}
              max={100}
              onChange={(v) => setGlobalConfig({ ...globalConfig, sender_block_threshold: v })}
              disabled={!editingGlobal}
            />
            <NumberField
              label="Unblock Threshold (%)"
              hint="Delivery rate considered recovered (at or above the block threshold)"
              value={globalConfig.sender_unblock_threshold}
              step={0.01}
              min={0}
              max={100}
              onChange={(v) => setGlobalConfig({ ...globalConfig, sender_unblock_threshold: v })}
              disabled={!editingGlobal}
            />
            <NumberField
              label="Sender Health Window"
              hint="Recent messages used to score a sender (100 to 10000)"
              value={globalConfig.sender_health_window_size}
              min={100}
              max={10000}
              onChange={(v) => setGlobalConfig({ ...globalConfig, sender_health_window_size: v })}
              disabled={!editingGlobal}
            />
            <NumberField
              label="DLR Timeout (seconds)"
              hint="Wait this long for a DLR before marking it missing (60 to 3600)"
              value={globalConfig.dlr_timeout_seconds}
              min={60}
              max={3600}
              onChange={(v) => setGlobalConfig({ ...globalConfig, dlr_timeout_seconds: v })}
              disabled={!editingGlobal}
            />
          </div>

          {globalConfig.schedule && (
            <div className="border-t border-white/10 pt-4">
              <h4 className="text-sm font-semibold text-zinc-300 mb-2">Fixed gateway schedules (read only)</h4>
              <ul className="text-sm text-zinc-400 space-y-1">
                <li>Missing DLR check: every {globalConfig.schedule.dlr_timeout_check_interval_seconds} seconds</li>
                <li>
                  Sender auto unblock: checked every {globalConfig.schedule.sender_auto_unblock_check_interval_minutes / 60} hours,
                  after {globalConfig.schedule.sender_auto_unblock_after_inactive_days} days without traffic
                </li>
                <li>
                  Health history cleanup: every {globalConfig.schedule.ring_buffer_cleanup_interval_hours} hours, keeping{' '}
                  {globalConfig.schedule.ring_buffer_retention_days} days
                </li>
              </ul>
            </div>
          )}
        </DataCard>
      )}

      <DataCard
        title="Routing Group Overrides"
        subtitle="A group with an override uses its own strategy, retries and failure threshold; everything else comes from the global settings"
      >
        <div className="space-y-3">
          {configs.length === 0 && !addingNew && (
            <p className="text-zinc-500 text-sm">No overrides: every routing group uses the global settings.</p>
          )}

          {configs.map((config) => {
            const editing = editingGroupId === config.routing_group_id && groupForm
            return (
              <div key={config.id} className="border border-white/10 rounded-lg p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <Settings2 className="h-4 w-4 text-blue-400" />
                    <span className="font-medium text-white">{config.routing_group_name}</span>
                  </div>
                  {!editing && (
                    <button
                      onClick={() =>
                        startGroupEdit(config.routing_group_id, {
                          selection_strategy: config.selection_strategy,
                          max_route_retries: config.max_route_retries,
                          route_failure_threshold: config.route_failure_threshold,
                        })
                      }
                      className="flex items-center gap-1 text-sm text-zinc-400 hover:text-white"
                    >
                      <Pencil className="h-4 w-4" /> Edit
                    </button>
                  )}
                </div>
                {editing ? (
                  <GroupForm
                    form={groupForm}
                    onChange={setGroupForm}
                    onSave={saveGroup}
                    onCancel={() => {
                      setEditingGroupId(null)
                      setGroupForm(null)
                    }}
                    saving={saving}
                  />
                ) : (
                  <div className="grid grid-cols-3 gap-4 text-sm">
                    <div>
                      <span className="text-zinc-500">Strategy</span>
                      <p className="text-white">
                        {STRATEGIES.find((s) => s.value === config.selection_strategy)?.label || config.selection_strategy}
                      </p>
                    </div>
                    <div>
                      <span className="text-zinc-500">Max retries</span>
                      <p className="text-white">{config.max_route_retries}</p>
                    </div>
                    <div>
                      <span className="text-zinc-500">Failure threshold</span>
                      <p className="text-white">{config.route_failure_threshold}</p>
                    </div>
                  </div>
                )}
              </div>
            )
          })}

          {addingNew && groupForm && (
            <div className="border border-blue-500/30 rounded-lg p-4">
              <p className="font-medium text-white mb-3">
                New override: {groups.find((g) => g.id === editingGroupId)?.name}
              </p>
              <GroupForm
                form={groupForm}
                onChange={setGroupForm}
                onSave={saveGroup}
                onCancel={() => {
                  setEditingGroupId(null)
                  setGroupForm(null)
                }}
                saving={saving}
              />
            </div>
          )}

          {groupsWithoutOverride.length > 0 && editingGroupId === null && (
            <div className="flex items-center gap-2 pt-2">
              <select
                aria-label="Routing group without an override"
                value={newGroupId}
                onChange={(e) => setNewGroupId(e.target.value)}
                className="px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-white text-sm"
              >
                <option value="" className="bg-zinc-900">
                  Choose a routing group
                </option>
                {groupsWithoutOverride.map((g) => (
                  <option key={g.id} value={g.id} className="bg-zinc-900">
                    {g.name}
                  </option>
                ))}
              </select>
              <button
                onClick={() => startGroupEdit(Number(newGroupId), defaultsFromGlobal)}
                disabled={!newGroupId}
                className="flex items-center gap-2 px-3 py-2 text-sm bg-white text-black rounded-lg hover:bg-zinc-200 disabled:opacity-50"
              >
                <Plus className="h-4 w-4" /> Add override
              </button>
            </div>
          )}
        </div>
      </DataCard>
    </div>
  )
}

function GroupForm({
  form,
  onChange,
  onSave,
  onCancel,
  saving,
}: Readonly<{
  form: GroupConfigForm
  onChange: (form: GroupConfigForm) => void
  onSave: () => void
  onCancel: () => void
  saving: boolean
}>) {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StrategyField value={form.selection_strategy} onChange={(v) => onChange({ ...form, selection_strategy: v })} />
        <NumberField
          label="Max Route Retries"
          hint="0 to 5"
          value={form.max_route_retries}
          min={0}
          max={5}
          onChange={(v) => onChange({ ...form, max_route_retries: v })}
        />
        <NumberField
          label="Route Failure Threshold"
          hint="Consecutive failures before a route is marked unavailable (1 to 20)"
          value={form.route_failure_threshold}
          min={1}
          max={20}
          onChange={(v) => onChange({ ...form, route_failure_threshold: v })}
        />
      </div>
      <div className="flex justify-end gap-2">
        <button onClick={onCancel} disabled={saving} className="flex items-center gap-2 px-3 py-1.5 text-sm text-zinc-400 hover:text-white">
          <X className="h-4 w-4" /> Cancel
        </button>
        <button
          onClick={onSave}
          disabled={saving}
          className="flex items-center gap-2 px-3 py-1.5 text-sm bg-white text-black rounded-lg hover:bg-zinc-200 disabled:opacity-50"
        >
          {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />} Save
        </button>
      </div>
    </div>
  )
}
