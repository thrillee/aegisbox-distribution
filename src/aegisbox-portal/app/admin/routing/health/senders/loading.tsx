import { TableLoadingSkeleton } from '@/components/routing/loading-components'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'

export default function Loading() {
  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Sender ID Health"
        description="Monitor delivery performance for all sender IDs across MNOs"
      />
      <DataCard title="">
        <TableLoadingSkeleton rows={10} />
      </DataCard>
    </div>
  )
}
