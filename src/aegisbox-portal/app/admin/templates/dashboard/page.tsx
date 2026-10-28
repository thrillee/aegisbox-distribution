import Link from 'next/link'
import { db } from '@/lib/db'
import { templates, otpAlternativeSenders, otpMessageTemplates, otpSenderTemplateAssignments } from '@/lib/schema'
import { count, eq } from 'drizzle-orm'
import { Shield, FileText, MessageSquare, ArrowRight } from 'lucide-react'
import { TemplatesDashboard } from './templates-dashboard'

async function getTemplatesStats() {
  try {
    const [totalTemplates, activeTemplates] = await Promise.all([
      db.select({ count: count() }).from(templates),
      db.select({ count: count() }).from(templates).where(eq(templates.status, 'active')),
    ])

    return {
      total: totalTemplates[0]?.count || 0,
      active: activeTemplates[0]?.count || 0,
      bySp: 0,
    }
  } catch (error) {
    console.error('Error fetching templates stats:', error)
    return { total: 0, active: 0, bySp: 0 }
  }
}

async function getOtpStats() {
  try {
    const [totalSenders, activeSenders, totalTemplates, activeAssignments] = await Promise.all([
      db.select({ count: count() }).from(otpAlternativeSenders),
      db.select({ count: count() }).from(otpAlternativeSenders).where(eq(otpAlternativeSenders.status, 'active' as any)),
      db.select({ count: count() }).from(otpMessageTemplates),
      db.select({ count: count() }).from(otpSenderTemplateAssignments).where(eq(otpSenderTemplateAssignments.status, 'active' as any)),
    ])

    return {
      senders: totalSenders[0]?.count || 0,
      activeSenders: activeSenders[0]?.count || 0,
      templates: totalTemplates[0]?.count || 0,
      assignments: activeAssignments[0]?.count || 0,
    }
  } catch (error) {
    console.error('Error fetching OTP stats:', error)
    return { senders: 0, activeSenders: 0, templates: 0, assignments: 0 }
  }
}

async function getRecentActivity() {
  try {
    const [recentTemplates, recentOtpTemplates] = await Promise.all([
      db.select({
        id: templates.id,
        name: templates.name,
        updatedAt: templates.updatedAt,
      })
        .from(templates)
        .orderBy(templates.updatedAt)
        .limit(5),
      db.select({
        id: otpMessageTemplates.id,
        name: otpMessageTemplates.name,
        updatedAt: otpMessageTemplates.updatedAt,
      })
        .from(otpMessageTemplates)
        .orderBy(otpMessageTemplates.updatedAt)
        .limit(5),
    ])

    return {
      templates: recentTemplates,
      otpTemplates: recentOtpTemplates,
    }
  } catch (error) {
    console.error('Error fetching recent activity:', error)
    return { templates: [], otpTemplates: [] }
  }
}

export default async function TemplatesDashboardPage() {
  const [templatesStats, otpStats, recentActivity] = await Promise.all([
    getTemplatesStats(),
    getOtpStats(),
    getRecentActivity(),
  ])

  return (
    <div className="min-h-screen bg-black">
      <TemplatesDashboard
        templatesStats={templatesStats}
        otpStats={otpStats}
        recentActivity={recentActivity}
      />
    </div>
  )
}
