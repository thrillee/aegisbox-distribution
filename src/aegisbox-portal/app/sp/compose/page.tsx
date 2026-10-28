'use client'

import { useState } from 'react'
import { Send, Clock, Calendar, MessageSquare, Users, DollarSign, AlertCircle } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

export default function ComposePage() {
  const [recipients, setRecipients] = useState('')
  const [message, setMessage] = useState('')
  const [senderId, setSenderId] = useState('')
  const [messageType, setMessageType] = useState('sms')
  const [scheduleDate, setScheduleDate] = useState('')

  const senders = [
    { id: '1', name: 'PROMO-NG', type: 'regular', status: 'approved' },
    { id: '2', name: 'ALERT-NG', type: 'regular', status: 'approved' },
    { id: '3', name: 'OTP-NG', type: 'otp', status: 'approved' },
  ]

  const messageLength = message.length
  const segments = Math.ceil(messageLength / 160) || 1
  const estimatedCost = segments * 3.50

  const recipientCount = recipients.split(/[\n,]/).filter(r => r.trim()).length

  return (
    <div className="p-8 space-y-6 max-w-4xl">
      <PageHeader
        title="Send SMS"
        description="Compose and send SMS to your recipients"
      />

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <DataCard title="" className="!p-6">
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-1">
                  Sender ID *
                </label>
                <select
                  value={senderId}
                  onChange={(e) => setSenderId(e.target.value)}
                  className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
                >
                  <option value="">Select sender ID</option>
                  {senders.map((sender) => (
                    <option key={sender.id} value={sender.id}>
                      {sender.name} ({sender.type.toUpperCase()})
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-1">
                  Recipients *
                </label>
                <textarea
                  value={recipients}
                  onChange={(e) => setRecipients(e.target.value)}
                  placeholder="Enter phone numbers, one per line or comma-separated&#10;+2347031234567&#10;+2348029876543"
                  className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white placeholder:text-zinc-500 focus:ring-2 focus:ring-white/20"
                  rows={5}
                />
                <div className="flex items-center justify-between mt-2">
                  <span className="text-sm text-zinc-500">
                    {recipientCount} recipient{recipientCount !== 1 ? 's' : ''}
                  </span>
                  <button className="text-sm text-white hover:underline">
                    Upload CSV
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-1">
                  Message *
                </label>
                <textarea
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  placeholder="Type your message here..."
                  className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white placeholder:text-zinc-500 focus:ring-2 focus:ring-white/20"
                  rows={6}
                />
                <div className="flex items-center justify-between mt-2">
                  <div className="flex items-center gap-4">
                    <span className={`text-sm ${messageLength > 160 ? 'text-amber-400' : 'text-zinc-500'}`}>
                      {messageLength}/160 characters
                    </span>
                    <span className="text-sm text-zinc-500">
                      {segments} segment{segments !== 1 ? 's' : ''}
                    </span>
                  </div>
                  <span className="text-sm text-zinc-500">
                    Est. Cost: ₦{estimatedCost.toFixed(2)}
                  </span>
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-3">
                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-1">
                    Message Type
                  </label>
                  <select
                    value={messageType}
                    onChange={(e) => setMessageType(e.target.value)}
                    className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
                  >
                    <option value="sms">Regular SMS</option>
                    <option value="otp">OTP</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-1">
                    Schedule (Optional)
                  </label>
                  <input
                    type="datetime-local"
                    value={scheduleDate}
                    onChange={(e) => setScheduleDate(e.target.value)}
                    className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
                  />
                </div>
              </div>
            </div>
          </DataCard>

          <div className="flex items-center justify-end gap-4">
            <Button variant="outline" className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5">
              Save as Draft
            </Button>
            <Button className="bg-white text-black hover:bg-zinc-200">
              <Send className="h-4 w-4" />
              {scheduleDate ? 'Schedule Send' : 'Send Now'}
            </Button>
          </div>
        </div>

        <div className="space-y-6">
          <DataCard title="Quick Stats" subtitle="Message overview">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-zinc-400">Recipients</span>
                <span className="font-medium text-white">{recipientCount}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-zinc-400">Segments</span>
                <span className="font-medium text-white">{segments}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-zinc-400">Est. Total Cost</span>
                <span className="font-medium text-emerald-400">₦{(estimatedCost * recipientCount).toFixed(2)}</span>
              </div>
            </div>
          </DataCard>

          <DataCard title="Tips" subtitle="Best practices">
            <ul className="space-y-3 text-sm text-zinc-400">
              <li className="flex items-start gap-2">
                <MessageSquare className="h-4 w-4 text-white mt-0.5" />
                <span>Keep messages under 160 characters to avoid extra charges</span>
              </li>
              <li className="flex items-start gap-2">
                <Users className="h-4 w-4 text-white mt-0.5" />
                <span>Use recipient groups for bulk messaging</span>
              </li>
              <li className="flex items-start gap-2">
                <Clock className="h-4 w-4 text-white mt-0.5" />
                <span>Schedule messages for optimal delivery times</span>
              </li>
              <li className="flex items-start gap-2">
                <AlertCircle className="h-4 w-4 text-white mt-0.5" />
                <span>Ensure sender IDs are approved before sending</span>
              </li>
            </ul>
          </DataCard>

          <div className="rounded-2xl border border-amber-500/20 bg-amber-500/10 p-4">
            <div className="flex items-start gap-3">
              <AlertCircle className="h-5 w-5 text-amber-400 mt-0.5" />
              <div>
                <p className="text-sm font-medium text-amber-400">
                  Low Balance Warning
                </p>
                <p className="text-sm text-amber-400/80 mt-1">
                  Your wallet balance is low. Top up to continue sending messages.
                </p>
                <button className="text-sm font-medium text-amber-400 hover:underline mt-2">
                  Fund Wallet →
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
