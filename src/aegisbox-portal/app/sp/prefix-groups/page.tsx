'use client'

import { useState } from 'react'
import { Plus, Search, Edit2, Trash2, Eye, Phone, Users, ArrowRight } from 'lucide-react'

const mockPrefixGroups = [
  {
    id: '1',
    name: 'Corporate Contacts',
    description: 'Corporate clients and business partners',
    prefixes: ['0803', '0805', '0806', '0810', '0811', '0812', '0813', '0814', '0815', '0816'],
    contactCount: 15000,
    status: 'active',
    createdAt: '2024-01-15 10:30:00',
  },
  {
    id: '2',
    name: 'Nigerian Diaspora',
    description: 'Contacts registered from abroad',
    prefixes: ['+1', '+44', '+234'],
    contactCount: 8500,
    status: 'active',
    createdAt: '2024-02-20 14:15:00',
  },
  {
    id: '3',
    name: 'Enterprise Clients',
    description: 'Premium enterprise customers',
    prefixes: ['0701', '0702', '0703', '0704', '0705'],
    contactCount: 3200,
    status: 'active',
    createdAt: '2024-03-10 09:00:00',
  },
  {
    id: '4',
    name: 'VIP Contacts',
    description: 'VIP and high-value customers',
    prefixes: ['0802', '0809', '0817', '0818'],
    contactCount: 1200,
    status: 'inactive',
    createdAt: '2024-04-05 11:45:00',
  },
]

const allPrefixes = [
  { prefix: '0701', mno: 'MTN', type: 'Mobile' },
  { prefix: '0702', mno: 'Airtel', type: 'Mobile' },
  { prefix: '0703', mno: 'Airtel', type: 'Mobile' },
  { prefix: '0704', mno: 'Airtel', type: 'Mobile' },
  { prefix: '0705', mno: 'Airtel', type: 'Mobile' },
  { prefix: '0802', mno: 'MTN', type: 'Mobile' },
  { prefix: '0803', mno: 'MTN', type: 'Mobile' },
  { prefix: '0804', mno: 'MTN', type: 'Mobile' },
  { prefix: '0805', mno: 'MTN', type: 'Mobile' },
  { prefix: '0806', mno: 'MTN', type: 'Mobile' },
  { prefix: '0807', mno: 'MTN', type: 'Mobile' },
  { prefix: '0808', mno: 'MTN', type: 'Mobile' },
  { prefix: '0809', mno: 'MTN', type: 'Mobile' },
  { prefix: '0810', mno: 'MTN', type: 'Mobile' },
  { prefix: '0811', mno: 'MTN', type: 'Mobile' },
  { prefix: '0812', mno: 'MTN', type: 'Mobile' },
  { prefix: '0813', mno: 'MTN', type: 'Mobile' },
  { prefix: '0814', mno: 'MTN', type: 'Mobile' },
  { prefix: '0815', mno: 'MTN', type: 'Mobile' },
  { prefix: '0816', mno: 'MTN', type: 'Mobile' },
  { prefix: '0901', mno: 'Glo', type: 'Mobile' },
  { prefix: '0902', mno: 'Glo', type: 'Mobile' },
  { prefix: '0903', mno: 'Glo', type: 'Mobile' },
  { prefix: '0904', mno: 'Glo', type: 'Mobile' },
  { prefix: '0905', mno: 'Glo', type: 'Mobile' },
  { prefix: '0906', mno: 'Glo', type: 'Mobile' },
  { prefix: '0805', mno: 'Vmobile', type: 'Mobile' },
  { prefix: '0806', mno: 'Vmobile', type: 'Mobile' },
  { prefix: '0817', mno: '9mobile', type: 'Mobile' },
  { prefix: '0818', mno: '9mobile', type: 'Mobile' },
  { prefix: '0819', mno: '9mobile', type: 'Mobile' },
  { prefix: '0908', mno: '9mobile', type: 'Mobile' },
  { prefix: '0909', mno: '9mobile', type: 'Mobile' },
]

export default function PrefixGroupsPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [editingGroup, setEditingGroup] = useState<typeof mockPrefixGroups[0] | null>(null)
  const [viewingGroup, setViewingGroup] = useState<typeof mockPrefixGroups[0] | null>(null)
  const [selectedPrefixes, setSelectedPrefixes] = useState<string[]>([])

  const filteredGroups = mockPrefixGroups.filter((group) =>
    group.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    group.description.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const getMNOForPrefix = (prefix: string) => {
    const exactMatch = allPrefixes.find(p => p.prefix === prefix)
    if (exactMatch) return exactMatch.mno
    if (prefix.startsWith('+')) return 'International'
    return 'Unknown'
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Prefix Groups</h1>
          <p className="text-zinc-400 mt-1">
            Organize contacts by phone number prefixes for targeted messaging
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
        >
          <Plus className="h-4 w-4" />
          Create Group
        </button>
      </div>

      <div className="relative flex-1 max-w-md">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
        <input
          type="text"
          placeholder="Search prefix groups..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full pl-10 pr-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
        />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {filteredGroups.map((group) => (
          <div
            key={group.id}
            className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6 hover:border-white/20 transition-colors"
          >
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="flex items-center justify-center w-12 h-12 bg-white/10 text-white rounded-xl">
                  <Users className="h-6 w-6" />
                </div>
                <div>
                  <h3 className="font-semibold text-white">{group.name}</h3>
                  <p className="text-sm text-zinc-500">{group.description}</p>
                </div>
              </div>
              <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${
                group.status === 'active'
                  ? 'bg-emerald-500/20 text-emerald-400'
                  : 'bg-zinc-500/20 text-zinc-400'
              }`}>
                {group.status}
              </span>
            </div>

            <div className="mb-4">
              <h4 className="text-sm font-medium text-zinc-400 mb-2">Prefixes</h4>
              <div className="flex flex-wrap gap-1">
                {group.prefixes.slice(0, 8).map((prefix) => (
                  <span
                    key={prefix}
                    className="inline-flex items-center gap-1 px-2 py-0.5 bg-white/10 text-zinc-300 rounded text-xs font-mono"
                  >
                    <Phone className="h-3 w-3" />
                    {prefix}
                    <span className="text-zinc-500">({getMNOForPrefix(prefix)})</span>
                  </span>
                ))}
                {group.prefixes.length > 8 && (
                  <span className="px-2 py-0.5 bg-white/10 text-zinc-500 rounded text-xs">
                    +{group.prefixes.length - 8} more
                  </span>
                )}
              </div>
            </div>

            <div className="flex items-center justify-between pt-4 border-t border-white/10">
              <div className="flex items-center gap-4 text-sm text-zinc-500">
                <span className="flex items-center gap-1">
                  <Phone className="h-4 w-4" />
                  {group.contactCount.toLocaleString()} contacts
                </span>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setViewingGroup(group)}
                  className="p-1.5 text-zinc-400 hover:text-white hover:bg-white/5 rounded transition-colors"
                  title="View Details"
                >
                  <Eye className="h-4 w-4" />
                </button>
                <button
                  onClick={() => setEditingGroup(group)}
                  className="p-1.5 text-zinc-400 hover:text-white hover:bg-white/5 rounded transition-colors"
                  title="Edit"
                >
                  <Edit2 className="h-4 w-4" />
                </button>
                <button
                  className="p-1.5 text-red-400 hover:text-red-300 hover:bg-red-500/10 rounded transition-colors"
                  title="Delete"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {showCreateModal && (
        <PrefixGroupModal
          onClose={() => setShowCreateModal(false)}
          onSave={(group) => {
            console.log('Saving group:', group)
            setShowCreateModal(false)
          }}
        />
      )}

      {editingGroup && (
        <PrefixGroupModal
          group={editingGroup}
          onClose={() => setEditingGroup(null)}
          onSave={(group) => {
            console.log('Updating group:', group)
            setEditingGroup(null)
          }}
        />
      )}

      {viewingGroup && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-zinc-900 border border-white/10 rounded-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-white/10">
              <h2 className="text-xl font-semibold text-white">{viewingGroup.name}</h2>
              <p className="text-sm text-zinc-500 mt-1">{viewingGroup.description}</p>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <h3 className="text-sm font-medium text-zinc-400 mb-3">All Prefixes</h3>
                <div className="flex flex-wrap gap-2">
                  {viewingGroup.prefixes.map((prefix) => {
                    const mno = getMNOForPrefix(prefix)
                    const mnoColor = mno === 'MTN' ? 'bg-emerald-500/20 text-emerald-400' :
                                    mno === 'Airtel' ? 'bg-red-500/20 text-red-400' :
                                    mno === 'Glo' ? 'bg-blue-500/20 text-blue-400' :
                                    mno === '9mobile' ? 'bg-yellow-500/20 text-yellow-400' :
                                    'bg-purple-500/20 text-purple-400'
                    return (
                      <span
                        key={prefix}
                        className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-sm font-mono ${mnoColor}`}
                      >
                        <Phone className="h-3 w-3" />
                        {prefix}
                      </span>
                    )
                  })}
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h3 className="text-sm font-medium text-zinc-400 mb-1">Total Contacts</h3>
                  <p className="text-white">{viewingGroup.contactCount.toLocaleString()}</p>
                </div>
                <div>
                  <h3 className="text-sm font-medium text-zinc-400 mb-1">Created At</h3>
                  <p className="text-white">
                    {new Date(viewingGroup.createdAt).toLocaleString()}
                  </p>
                </div>
              </div>
              <div>
                <h3 className="text-sm font-medium text-zinc-400 mb-2">MNOs Covered</h3>
                <div className="flex flex-wrap gap-2">
                  {Array.from(new Set(viewingGroup.prefixes.map(getMNOForPrefix))).map((mno) => {
                    const count = viewingGroup.prefixes.filter(p => getMNOForPrefix(p) === mno).length
                    const mnoColor = mno === 'MTN' ? 'bg-emerald-500/20 text-emerald-400' :
                                    mno === 'Airtel' ? 'bg-red-500/20 text-red-400' :
                                    mno === 'Glo' ? 'bg-blue-500/20 text-blue-400' :
                                    mno === '9mobile' ? 'bg-yellow-500/20 text-yellow-400' :
                                    'bg-purple-500/20 text-purple-400'
                    return (
                      <span
                        key={mno}
                        className={`px-3 py-1 rounded-full text-sm font-medium ${mnoColor}`}
                      >
                        {mno} ({count} prefixes)
                      </span>
                    )
                  })}
                </div>
              </div>
            </div>
            <div className="p-6 border-t border-white/10 flex justify-end">
              <button
                onClick={() => setViewingGroup(null)}
                className="px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function PrefixGroupModal({
  group,
  onClose,
  onSave,
}: {
  group?: typeof mockPrefixGroups[0]
  onClose: () => void
  onSave: (group: any) => void
}) {
  const [name, setName] = useState(group?.name || '')
  const [description, setDescription] = useState(group?.description || '')
  const [selectedPrefixes, setSelectedPrefixes] = useState<string[]>(group?.prefixes || [])

  const togglePrefix = (prefix: string) => {
    setSelectedPrefixes((prev) =>
      prev.includes(prefix) ? prev.filter((p) => p !== prefix) : [...prev, prefix]
    )
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-zinc-900 border border-white/10 rounded-2xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6 border-b border-white/10">
          <h2 className="text-xl font-semibold text-white">
            {group ? 'Edit Prefix Group' : 'Create New Prefix Group'}
          </h2>
        </div>
        <div className="p-6 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">
                Group Name
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Corporate Contacts"
                className="w-full px-3 py-2 bg-black/50 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">
                Description
              </label>
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="e.g., Corporate clients and partners"
                className="w-full px-3 py-2 bg-black/50 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-2">
              Select Prefixes ({selectedPrefixes.length} selected)
            </label>
            <div className="bg-black/50 rounded-xl p-4 max-h-80 overflow-y-auto">
              <div className="grid grid-cols-4 gap-2">
                {allPrefixes.map((item) => {
                  const isSelected = selectedPrefixes.includes(item.prefix)
                  return (
                    <button
                      key={item.prefix}
                      onClick={() => togglePrefix(item.prefix)}
                      className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-mono transition-colors ${
                        isSelected
                          ? 'bg-white text-black'
                          : 'bg-white/10 text-zinc-300 hover:bg-white/20'
                      }`}
                    >
                      <Phone className="h-3 w-3" />
                      <span>{item.prefix}</span>
                      <span className={`text-xs ml-auto ${isSelected ? 'text-black/70' : 'text-zinc-500'}`}>
                        {item.mno}
                      </span>
                    </button>
                  )
                })}
              </div>
            </div>
          </div>

          {selectedPrefixes.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-2">
                Selected Prefixes
              </label>
              <div className="flex flex-wrap gap-2">
                {selectedPrefixes.map((prefix) => {
                  const item = allPrefixes.find(p => p.prefix === prefix)
                  const mno = item?.mno || 'Unknown'
                  return (
                    <span
                      key={prefix}
                      className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-sm font-mono bg-white/10 text-zinc-300"
                    >
                      <Phone className="h-3 w-3" />
                      {prefix}
                      <button
                        onClick={() => togglePrefix(prefix)}
                        className="ml-1 hover:opacity-70"
                      >
                        ×
                      </button>
                    </span>
                  )
                })}
              </div>
            </div>
          )}
        </div>
        <div className="p-6 border-t border-white/10 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => onSave({ name, description, prefixes: selectedPrefixes })}
            className="px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
          >
            {group ? 'Save Changes' : 'Create Group'}
          </button>
        </div>
      </div>
    </div>
  )
}
