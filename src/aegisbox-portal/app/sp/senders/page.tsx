import { Plus, Search, Tag, CheckCircle, Clock, AlertCircle, Zap, Settings } from 'lucide-react'
import { db } from '@/lib/db'
import { getCurrentServiceProviderId, NOT_LINKED_MESSAGE } from '@/lib/service-provider'
import { senderIds } from '@/lib/schema'
import { desc, eq } from 'drizzle-orm'

async function getSenderIds(serviceProviderId: number) {
  const senderList = await db
    .select({
      id: senderIds.id,
      senderIdString: senderIds.senderIdString,
      status: senderIds.status,
      rateLimit: senderIds.rateLimit,
      currentUsageCount: senderIds.currentUsageCount,
      maxUsageCount: senderIds.maxUsageCount,
      createdAt: senderIds.createdAt,
    })
    .from(senderIds)
    .where(eq(senderIds.serviceProviderId, serviceProviderId))
    .orderBy(desc(senderIds.createdAt))

  return senderList.map(s => ({
    id: String(s.id),
    senderId: s.senderIdString || 'Unknown',
    status: (s.status as string) || 'pending',
    rateLimit: Number(s.rateLimit) || 100,
    usedToday: Number(s.currentUsageCount) || 0,
    maxDaily: Number(s.maxUsageCount) || 10000,
  }))
}

function getStatusBadge(status: string) {
  const styles: Record<string, { bg: string; icon: any; text: string }> = {
    approved: { bg: 'bg-emerald-500/20 text-emerald-400', icon: CheckCircle, text: 'Approved' },
    pending: { bg: 'bg-amber-500/20 text-amber-400', icon: Clock, text: 'Pending' },
    rejected: { bg: 'bg-red-500/20 text-red-400', icon: AlertCircle, text: 'Rejected' },
    active: { bg: 'bg-emerald-500/20 text-emerald-400', icon: CheckCircle, text: 'Active' },
  }

  const style = styles[status] || styles.pending
  const Icon = style.icon

  return (
    <span className={`flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${style.bg}`}>
      <Icon className="h-3 w-3" />
      <span>{style.text}</span>
    </span>
  )
}

export default async function SendersPage() {
  const serviceProviderId = await getCurrentServiceProviderId()

  if (!serviceProviderId) {
    return (
      <div className="p-8 text-center">
        <p className="text-zinc-400">{NOT_LINKED_MESSAGE}</p>
      </div>
    )
  }

  const senderList = await getSenderIds(serviceProviderId)

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Sender IDs</h1>
          <p className="text-zinc-400 mt-1">
            Manage your approved sender IDs for sending messages
          </p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors">
          <Plus className="h-4 w-4" />
          Request New Sender ID
        </button>
      </div>

      <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
        <div className="p-4 border-b border-white/10">
          <div className="flex items-center gap-4">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
              <input
                type="text"
                placeholder="Search sender IDs..."
                className="w-full pl-10 pr-4 py-2 bg-black/50 border border-white/10 rounded-xl text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
              />
            </div>
            <select className="px-4 py-2 bg-black/50 border border-white/10 rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-white/20">
              <option value="">All Status</option>
              <option value="approved">Approved</option>
              <option value="pending">Pending</option>
              <option value="rejected">Rejected</option>
            </select>
          </div>
        </div>

        {senderList.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-black/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Sender ID</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Rate Limit</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Used Today</th>
                  <th className="px-6 py-3 text-right text-xs font-semibold text-zinc-400 uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/10">
                {senderList.map((sender) => (
                  <tr key={sender.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-xl flex items-center justify-center bg-white/10">
                          <Tag className="h-5 w-5 text-white" />
                        </div>
                        <span className="font-medium text-white">{sender.senderId}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">{getStatusBadge(sender.status)}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-white">{sender.rateLimit}/min</td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <span className="text-white">{sender.usedToday.toLocaleString()}</span>
                        <span className="text-zinc-600">/ {sender.maxDaily.toLocaleString()}</span>
                      </div>
                      <div className="w-24 h-1.5 bg-zinc-800 rounded-full mt-1">
                        <div
                          className="h-full bg-white rounded-full"
                          style={{ width: `${Math.min((sender.usedToday / sender.maxDaily) * 100, 100)}%` }}
                        />
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <div className="flex items-center justify-end gap-2">
                        {(sender.status === 'approved' || sender.status === 'active') && (
                          <button className="p-2 hover:bg-white/10 rounded-lg transition-colors">
                            <Settings className="h-4 w-4 text-zinc-400" />
                          </button>
                        )}
                        <button className="px-3 py-1.5 text-sm font-medium text-white hover:bg-white/10 rounded-lg transition-colors">
                          View Details
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="p-8 text-center text-zinc-500">No sender IDs found</div>
        )}
      </div>

      <div className="bg-blue-500/10 border border-blue-500/20 rounded-2xl p-6">
        <h3 className="font-semibold text-white mb-2">Need a New Sender ID?</h3>
        <p className="text-sm text-zinc-400 mb-4">
          Request a new sender ID for your brand. Sender IDs must be approved before use.
          Regular sender IDs typically take 1-3 business days. OTP sender IDs are usually approved within 24 hours.
        </p>
        <button className="px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors text-sm font-medium">
          Request Sender ID
        </button>
      </div>
    </div>
  )
}
