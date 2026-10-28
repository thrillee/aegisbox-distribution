import Link from 'next/link'
import { Plus, RefreshCw, FileText, Settings } from 'lucide-react'
import { db } from '@/lib/db'
import { otpMessageTemplates } from '@/lib/schema'
import { desc } from 'drizzle-orm'
import { OtpTemplatesList } from './otp-templates-list'

async function getOtpTemplates() {
  try {
    const templates = await db.select({
      id: otpMessageTemplates.id,
      name: otpMessageTemplates.name,
      contentTemplate: otpMessageTemplates.contentTemplate,
      defaultBrandName: otpMessageTemplates.defaultBrandName,
      status: otpMessageTemplates.status,
      createdAt: otpMessageTemplates.createdAt,
      updatedAt: otpMessageTemplates.updatedAt,
    })
      .from(otpMessageTemplates)
      .orderBy(desc(otpMessageTemplates.createdAt))

    return templates
  } catch (error) {
    console.error('Error fetching OTP message templates:', error)
    return []
  }
}

export default async function OtpTemplatesPage() {
  const initialTemplates = await getOtpTemplates()

  const stats = {
    total: initialTemplates.length,
    active: initialTemplates.filter((t: any) => t.status === 'active').length,
    inactive: initialTemplates.filter((t: any) => t.status === 'inactive').length,
  }

  return (
    <div className="min-h-screen bg-black">
      <OtpTemplatesList initialTemplates={initialTemplates} stats={stats} />
    </div>
  )
}
