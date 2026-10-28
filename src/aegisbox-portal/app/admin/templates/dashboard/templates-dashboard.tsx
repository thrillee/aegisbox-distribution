'use client'

import Link from 'next/link'
import { FileText, Shield, MessageSquare, ArrowRight, Plus, Settings } from 'lucide-react'

interface TemplatesStats {
  total: number
  active: number
  bySp: number
}

interface OtpStats {
  senders: number
  activeSenders: number
  templates: number
  assignments: number
}

interface RecentActivity {
  templates: Array<{ id: number; name: string; updatedAt: Date }>
  otpTemplates: Array<{ id: number; name: string; updatedAt: Date }>
}

interface TemplatesDashboardProps {
  templatesStats: TemplatesStats
  otpStats: OtpStats
  recentActivity: RecentActivity
}

function StatCard({
  title,
  value,
  icon: Icon,
  color,
  href,
}: {
  title: string
  value: number
  icon: any
  color: string
  href?: string
}) {
  const content = (
    <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6 hover:bg-white/5 transition-colors">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-zinc-400">{title}</p>
          <p className="text-3xl font-bold text-white mt-1">{value}</p>
        </div>
        <div className={`p-3 rounded-lg ${color}`}>
          <Icon className="h-6 w-6" />
        </div>
      </div>
    </div>
  )

  if (href) {
    return <Link href={href}>{content}</Link>
  }

  return content
}

export function TemplatesDashboard({ templatesStats, otpStats, recentActivity }: TemplatesDashboardProps) {
  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Message Templates Dashboard</h1>
          <p className="text-zinc-400 mt-1">
            Overview of all message templates and OTP configurations
          </p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Templates"
          value={templatesStats.total}
          icon={FileText}
          color="bg-blue-500/20 text-blue-400"
          href="/admin/templates"
        />
        <StatCard
          title="Active Templates"
          value={templatesStats.active}
          icon={Settings}
          color="bg-emerald-500/20 text-emerald-400"
          href="/admin/templates"
        />
        <StatCard
          title="OTP Senders"
          value={otpStats.senders}
          icon={Shield}
          color="bg-purple-500/20 text-purple-400"
          href="/admin/otp-senders"
        />
        <StatCard
          title="OTP Templates"
          value={otpStats.templates}
          icon={MessageSquare}
          color="bg-amber-500/20 text-amber-400"
          href="/admin/otp-templates"
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-white">Message Templates</h2>
            <Link href="/admin/templates" className="text-white hover:underline text-sm flex items-center gap-1">
              View all <ArrowRight className="h-4 w-4" />
            </Link>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-black/50 rounded-xl">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-blue-500/20 rounded-lg">
                  <FileText className="h-5 w-5 text-blue-400" />
                </div>
                <div>
                  <p className="font-medium text-white">Total Templates</p>
                  <p className="text-sm text-zinc-500">All message templates</p>
                </div>
              </div>
              <p className="text-2xl font-bold text-white">{templatesStats.total}</p>
            </div>

            <div className="flex items-center justify-between p-4 bg-black/50 rounded-xl">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-emerald-500/20 rounded-lg">
                  <Settings className="h-5 w-5 text-emerald-400" />
                </div>
                <div>
                  <p className="font-medium text-white">Active Templates</p>
                  <p className="text-sm text-zinc-500">Currently in use</p>
                </div>
              </div>
              <p className="text-2xl font-bold text-emerald-400">{templatesStats.active}</p>
            </div>
          </div>

          <Link
            href="/admin/templates"
            className="mt-6 flex items-center justify-center gap-2 w-full py-3 border border-white/10 rounded-xl text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
          >
            <Plus className="h-4 w-4" />
            Create New Template
          </Link>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-white">OTP Configuration</h2>
            <Link href="/admin/otp-senders" className="text-white hover:underline text-sm flex items-center gap-1">
              View all <ArrowRight className="h-4 w-4" />
            </Link>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-black/50 rounded-xl">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-purple-500/20 rounded-lg">
                  <Shield className="h-5 w-5 text-purple-400" />
                </div>
                <div>
                  <p className="font-medium text-white">Active Senders</p>
                  <p className="text-sm text-zinc-500">Available OTP senders</p>
                </div>
              </div>
              <p className="text-2xl font-bold text-purple-400">{otpStats.activeSenders}</p>
            </div>

            <div className="flex items-center justify-between p-4 bg-black/50 rounded-xl">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-amber-500/20 rounded-lg">
                  <MessageSquare className="h-5 w-5 text-amber-400" />
                </div>
                <div>
                  <p className="font-medium text-white">Active Assignments</p>
                  <p className="text-sm text-zinc-500">Sender-template mappings</p>
                </div>
              </div>
              <p className="text-2xl font-bold text-amber-400">{otpStats.assignments}</p>
            </div>
          </div>

          <div className="mt-6 grid grid-cols-2 gap-3">
            <Link
              href="/admin/otp-senders"
              className="flex items-center justify-center gap-2 py-3 border border-white/10 rounded-xl text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
            >
              <Shield className="h-4 w-4" />
              OTP Senders
            </Link>
            <Link
              href="/admin/otp-templates"
              className="flex items-center justify-center gap-2 py-3 border border-white/10 rounded-xl text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
            >
              <MessageSquare className="h-4 w-4" />
              OTP Templates
            </Link>
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4">Recent Templates</h2>

          {recentActivity.templates.length === 0 ? (
            <div className="text-center py-8">
              <FileText className="h-10 w-10 text-zinc-500 mx-auto mb-3" />
              <p className="text-zinc-500">No templates yet</p>
              <Link href="/admin/templates" className="text-white hover:underline text-sm">
                Create your first template
              </Link>
            </div>
          ) : (
            <div className="space-y-3">
              {recentActivity.templates.map((template) => (
                <div key={template.id} className="flex items-center justify-between p-3 bg-black/50 rounded-xl">
                  <div className="flex items-center gap-3">
                    <FileText className="h-5 w-5 text-zinc-500" />
                    <span className="text-sm font-medium text-white">{template.name}</span>
                  </div>
                  <span className="text-xs text-zinc-500">
                    {new Date(template.updatedAt).toISOString().split('T')[0]}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4">Recent OTP Templates</h2>

          {recentActivity.otpTemplates.length === 0 ? (
            <div className="text-center py-8">
              <MessageSquare className="h-10 w-10 text-zinc-500 mx-auto mb-3" />
              <p className="text-zinc-500">No OTP templates yet</p>
              <Link href="/admin/otp-templates" className="text-white hover:underline text-sm">
                Create your first OTP template
              </Link>
            </div>
          ) : (
            <div className="space-y-3">
              {recentActivity.otpTemplates.map((template) => (
                <div key={template.id} className="flex items-center justify-between p-3 bg-black/50 rounded-xl">
                  <div className="flex items-center gap-3">
                    <MessageSquare className="h-5 w-5 text-zinc-500" />
                    <span className="text-sm font-medium text-white">{template.name}</span>
                  </div>
                  <span className="text-xs text-zinc-500">
                    {new Date(template.updatedAt).toISOString().split('T')[0]}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
