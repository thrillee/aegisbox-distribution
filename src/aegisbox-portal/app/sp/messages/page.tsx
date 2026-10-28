import { Download, ChevronLeft, ChevronRight, Eye } from 'lucide-react'
import { db } from '@/lib/db'
import { getCurrentServiceProviderId, NOT_LINKED_MESSAGE } from '@/lib/service-provider'
import { messages } from '@/lib/schema'
import { desc, eq, gte, and, sql } from 'drizzle-orm'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { StatusBadge } from '@/components/dashboard/status-badge'
import { Button } from '@/components/ui/button'
import { formatMoney } from '@/lib/money'

async function getMessages(
  serviceProviderId: number,
  searchParams: { status?: string; sender?: string; dateFrom?: string; dateTo?: string; page?: string }
) {
  const pageNum = parseInt(searchParams.page || '1')
  const pageSize = 20
  const offset = (pageNum - 1) * pageSize

  const conditions = []

  if (serviceProviderId) {
    conditions.push(eq(messages.serviceProviderId, serviceProviderId))
  }
  if (searchParams.status) {
    conditions.push(eq(messages.finalStatus, searchParams.status))
  }
  if (searchParams.sender) {
    conditions.push(eq(messages.originalSourceAddr, searchParams.sender))
  }
  if (searchParams.dateFrom) {
    conditions.push(gte(messages.submittedAt, new Date(searchParams.dateFrom)))
  }
  if (searchParams.dateTo) {
    conditions.push(sql`${messages.submittedAt} <= ${new Date(searchParams.dateTo)}`)
  }

  const whereClause = conditions.length > 0 ? and(...conditions) : undefined

  const [messageList, totalResult] = await Promise.all([
    db
      .select({
        id: messages.id,
        clientMessageId: messages.clientMessageId,
        sender: messages.originalSourceAddr,
        recipient: messages.originalDestinationAddr,
        segments: messages.totalSegments,
        cost: messages.cost,
        currency: messages.currencyCode,
        status: messages.finalStatus,
        submittedAt: messages.submittedAt,
        completedAt: messages.completedAt,
      })
      .from(messages)
      .where(whereClause)
      .orderBy(desc(messages.submittedAt))
      .limit(pageSize)
      .offset(offset),
    db
      .select({ count: sql<number>`count(*)` })
      .from(messages)
      .where(whereClause),
  ])

  return {
    messages: messageList.map(msg => ({
      id: `MSG${String(msg.id).padStart(6, '0')}`,
      clientMessageId: msg.clientMessageId,
      sender: msg.sender || 'Unknown',
      recipient: msg.recipient ? `${msg.recipient.slice(0, 6)}***${msg.recipient.slice(-4)}` : 'Unknown',
      segments: Number(msg.segments),
      cost: msg.cost ? Number(msg.cost) : 0,
      currency: msg.currency,
      status: msg.status || 'pending',
      sentAt: msg.submittedAt?.toISOString() || null,
      deliveredAt: msg.completedAt?.toISOString() || null,
    })),
    total: Number(totalResult[0]?.count) || 0,
    page: pageNum,
    pageSize,
  }
}

export default async function MessagesPage({
  searchParams,
}: {
  searchParams: Promise<{ status?: string; sender?: string; dateFrom?: string; dateTo?: string; page?: string }>
}) {
  const serviceProviderId = await getCurrentServiceProviderId()

  if (!serviceProviderId) {
    return (
      <div className="p-8 text-center">
        <p className="text-zinc-400">{NOT_LINKED_MESSAGE}</p>
      </div>
    )
  }

  const params = await searchParams
  const { messages: messageList, total, page, pageSize } = await getMessages(serviceProviderId, params)

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Message History"
        description="View and manage your sent messages"
        actions={
          <Button
            variant="outline"
            className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5"
          >
            <Download className="h-4 w-4" />
            Export CSV
          </Button>
        }
      />

      <DataCard title="" className="!p-0">
        <div className="p-4 border-b border-white/10">
          <div className="flex flex-wrap items-center gap-4">
            <div className="relative flex-1 min-w-[200px] max-w-md">
              <input
                type="text"
                placeholder="Search by message ID or phone number..."
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white placeholder:text-zinc-500 focus:ring-2 focus:ring-white/20 focus:border-transparent"
              />
            </div>
            <select className="px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20">
              <option value="">All Status</option>
              <option value="delivered">Delivered</option>
              <option value="pending">Pending</option>
              <option value="failed">Failed</option>
            </select>
            <select className="px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20">
              <option value="">All Senders</option>
              <option value="PROMO-NG">PROMO-NG</option>
              <option value="ALERT-NG">ALERT-NG</option>
              <option value="OTP-NG">OTP-NG</option>
            </select>
            <div className="flex items-center gap-2">
              <input
                type="date"
                className="px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              />
              <span className="text-zinc-500">to</span>
              <input
                type="date"
                className="px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              />
            </div>
            <Button className="bg-white text-black hover:bg-zinc-200">
              <span className="flex items-center gap-2">
                Apply Filters
              </span>
            </Button>
          </div>
        </div>

        {messageList.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-zinc-800/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Message ID</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Sender</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Recipient</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Segments</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Cost</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Sent At</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Delivered</th>
                  <th className="px-6 py-3 text-right text-xs font-semibold text-zinc-400 uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/10">
                {messageList.map((msg) => (
                  <tr key={msg.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap font-mono text-sm text-white">{msg.id}</td>
                    <td className="px-6 py-4 whitespace-nowrap font-medium text-white">{msg.sender}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-zinc-400">{msg.recipient}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-white">{msg.segments}</td>
                    <td className="px-6 py-4 whitespace-nowrap font-medium text-white">{formatMoney(msg.cost, msg.currency)}</td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <StatusBadge status={msg.status as 'delivered' | 'pending' | 'failed'} />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-zinc-400">
                      {msg.sentAt ? new Date(msg.sentAt).toLocaleString() : '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-zinc-400">
                      {msg.deliveredAt ? new Date(msg.deliveredAt).toLocaleString() : '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <button className="p-2 hover:bg-white/5 rounded-lg transition-colors">
                        <Eye className="h-4 w-4 text-zinc-400" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="p-8 text-center text-zinc-500">No messages found</div>
        )}

        <div className="p-4 border-t border-white/10">
          <div className="flex items-center justify-between">
            <p className="text-sm text-zinc-400">
              Showing {(page - 1) * pageSize + 1} to {Math.min(page * pageSize, total)} of {total} messages
            </p>
            <div className="flex items-center gap-2">
              <button
                className="p-2 border border-white/10 rounded-lg text-zinc-400 hover:bg-white/5 disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={page <= 1}
              >
                <ChevronLeft className="h-4 w-4" />
              </button>
              <span className="px-3 py-1 text-sm text-white">
                {page} of {Math.ceil(total / pageSize)}
              </span>
              <button
                className="p-2 border border-white/10 rounded-lg text-zinc-400 hover:bg-white/5 disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={page >= Math.ceil(total / pageSize)}
              >
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
      </DataCard>
    </div>
  )
}
