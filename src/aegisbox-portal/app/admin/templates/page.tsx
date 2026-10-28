import Link from 'next/link'
import { Plus, FileText, RefreshCw, Settings } from 'lucide-react'
import { db } from '@/lib/db'
import { templates, serviceProviders } from '@/lib/schema'
import { desc, eq } from 'drizzle-orm'
import { TemplatesList } from './templates-list'

async function getTemplates() {
  try {
    const templateList = await db.select({
      id: templates.id,
      name: templates.name,
      content: templates.content,
      status: templates.status,
      serviceProviderId: templates.serviceProviderId,
      createdAt: templates.createdAt,
      updatedAt: templates.updatedAt,
    })
      .from(templates)
      .orderBy(desc(templates.createdAt))

    return templateList
  } catch (error) {
    console.error('Error fetching templates:', error)
    return []
  }
}

async function getServiceProviders() {
  try {
    const spList = await db.select({
      id: serviceProviders.id,
      name: serviceProviders.name,
    })
      .from(serviceProviders)
      .orderBy(serviceProviders.name)

    return spList
  } catch (error) {
    console.error('Error fetching service providers:', error)
    return []
  }
}

export default async function TemplatesPage() {
  const [initialTemplates, serviceProviders] = await Promise.all([
    getTemplates(),
    getServiceProviders(),
  ])

  const stats = {
    total: initialTemplates.length,
    active: initialTemplates.filter((t: any) => t.status === 'active').length,
    inactive: initialTemplates.filter((t: any) => t.status === 'inactive').length,
  }

  return (
    <div className="min-h-screen bg-black">
      <TemplatesList
        initialTemplates={initialTemplates}
        serviceProviders={serviceProviders}
        stats={stats}
      />
    </div>
  )
}
