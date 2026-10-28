import { TableLoadingSkeleton } from '@/components/routing/loading-components'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'

export default function Loading() {
  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Route Health"
        description="Monitor health metrics for all routing assignments"
      />
      <DataCard title="">
        <TableLoadingSkeleton rows={10} />
      </DataCard>
    </div>
  )
}
