'use client'

import { useState } from 'react'
import { Save, Globe, Clock, Mail, Shield, Zap, Bell, DollarSign } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

export default function SettingsPage() {
  const [loading, setLoading] = useState(false)
  const [settings, setSettings] = useState({
    platformName: 'Aegisbox Portal',
    timezone: 'Africa/Lagos',
    defaultCurrency: 'NGN',
    supportEmail: 'support@aegisbox.com',
    adminEmail: 'admin@aegisbox.com',
    maxMessagesPerSecond: '100',
    defaultRoute: 'system_default',
    lowBalanceThreshold: '10000',
    enable2FA: true,
    enableEmailNotifications: true,
    enableWebhookNotifications: true,
  })

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target
    if (type === 'checkbox') {
      setSettings(prev => ({ ...prev, [name]: (e.target as HTMLInputElement).checked }))
    } else {
      setSettings(prev => ({ ...prev, [name]: value }))
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    await new Promise(resolve => setTimeout(resolve, 1000))
    setLoading(false)
    alert('Settings saved successfully!')
  }

  return (
    <div className="p-8 space-y-6 max-w-4xl">
      <PageHeader
        title="Settings"
        description="Configure platform-wide settings"
      />

      <form onSubmit={handleSubmit} className="space-y-6">
        <DataCard
          title="General Settings"
          subtitle="Basic platform configuration"
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Platform Name</label>
              <input 
                type="text" 
                name="platformName" 
                value={settings.platformName} 
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white" 
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Timezone</label>
              <select 
                name="timezone" 
                value={settings.timezone} 
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
              >
                <option value="Africa/Lagos">Africa/Lagos (GMT+1)</option>
                <option value="Africa/Accra">Africa/Accra (GMT+0)</option>
                <option value="UTC">UTC</option>
                <option value="Europe/London">Europe/London</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Default Currency</label>
              <select 
                name="defaultCurrency" 
                value={settings.defaultCurrency} 
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
              >
                <option value="NGN">NGN - Nigerian Naira</option>
                <option value="USD">USD - US Dollar</option>
                <option value="GHS">GHS - Ghanaian Cedi</option>
              </select>
            </div>
          </div>
        </DataCard>

        <DataCard
          title="Email Settings"
          subtitle="Email configuration for notifications"
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Support Email</label>
              <input 
                type="email" 
                name="supportEmail" 
                value={settings.supportEmail} 
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white" 
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Admin Email</label>
              <input 
                type="email" 
                name="adminEmail" 
                value={settings.adminEmail} 
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white" 
              />
            </div>
          </div>
        </DataCard>

        <DataCard
          title="Processing Settings"
          subtitle="Message processing configuration"
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Max Messages/Second</label>
              <input 
                type="number" 
                name="maxMessagesPerSecond" 
                value={settings.maxMessagesPerSecond} 
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white" 
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Default Route</label>
              <select 
                name="defaultRoute" 
                value={settings.defaultRoute} 
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
              >
                <option value="system_default">System Default</option>
                <option value="mtn_primary">MTN Primary</option>
              </select>
            </div>
          </div>
        </DataCard>

        <DataCard
          title="Security & Notifications"
          subtitle="Security and alert configuration"
        >
          <div className="space-y-4">
            <label className="flex items-center gap-3 cursor-pointer">
              <input 
                type="checkbox" 
                name="enable2FA" 
                checked={settings.enable2FA} 
                onChange={handleChange}
                className="w-5 h-5 rounded border-white/10 bg-zinc-800 text-white" 
              />
              <span className="text-white">Enable Two-Factor Authentication</span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer">
              <input 
                type="checkbox" 
                name="enableEmailNotifications" 
                checked={settings.enableEmailNotifications} 
                onChange={handleChange}
                className="w-5 h-5 rounded border-white/10 bg-zinc-800 text-white" 
              />
              <span className="text-white">Enable Email Notifications</span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer">
              <input 
                type="checkbox" 
                name="enableWebhookNotifications" 
                checked={settings.enableWebhookNotifications} 
                onChange={handleChange}
                className="w-5 h-5 rounded border-white/10 bg-zinc-800 text-white" 
              />
              <span className="text-white">Enable Webhook Notifications</span>
            </label>
          </div>
        </DataCard>

        <div className="flex justify-end">
          <Button type="submit" disabled={loading} className="bg-white text-black hover:bg-zinc-200">
            {loading ? 'Saving...' : 'Save Settings'}
          </Button>
        </div>
      </form>
    </div>
  )
}
