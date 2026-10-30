'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useUser, useOrganization, SignOutButton } from '@clerk/nextjs'
import {
  LayoutDashboard,
  Send,
  Plug,
  MessageSquare,
  History,
  Tag,
  Key,
  Wallet,
  BarChart3,
  Settings,
  Users,
  Radio,
  DollarSign,
  FileText,
  Shield,
  TrendingUp,
  Building,
  ChevronDown,
  ChevronRight,
  Menu,
  X,
  Loader2,
  Link2,
  Activity,
  Settings2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { formatMoney } from '@/lib/money'

interface NavItem {
  href: string
  label: string
  icon: React.ElementType
}

interface NavGroup {
  label: string
  items: NavItem[]
}

const ADMIN_NAV_GROUPS: NavGroup[] = [
  {
    label: 'Overview',
    items: [{ href: '/admin/dashboard', label: 'Dashboard', icon: LayoutDashboard }],
  },
  {
    label: 'Messaging',
    items: [
      { href: '/admin/messages', label: 'Messages', icon: MessageSquare },
      { href: '/admin/messages/analytics', label: 'Message Analytics', icon: BarChart3 },
      { href: '/admin/messages/daily-report', label: 'Daily Report', icon: FileText },
      { href: '/admin/messages/pending-dlr', label: 'Pending DLRs', icon: Send },
      { href: '/admin/messages/queues/dlr', label: 'DLR Queue', icon: Loader2 },
      { href: '/admin/templates', label: 'Templates', icon: FileText },
      { href: '/admin/templates/dashboard', label: 'Template Dashboard', icon: FileText },
      { href: '/admin/sender-ids', label: 'Sender IDs', icon: Tag },
      { href: '/admin/otp-alternative-senders', label: 'OTP Alternative Senders', icon: Shield },
      { href: '/admin/otp-templates', label: 'OTP Templates', icon: Shield },
      { href: '/admin/otp-sender-template-assignments', label: 'OTP Assignments', icon: Link2 },
    ],
  },
  {
    label: 'Network',
    items: [
      { href: '/admin/mnos', label: 'MNO Networks', icon: Radio },
      { href: '/admin/routing', label: 'Routing', icon: TrendingUp },
      { href: '/admin/routing/health', label: 'Health Monitoring', icon: Activity },
      { href: '/admin/routing/config', label: 'Routing Config', icon: Settings2 },
    ],
  },
  {
    label: 'Customers',
    items: [
      { href: '/admin/service-providers', label: 'Service Providers', icon: Users },
      { href: '/admin/wallets', label: 'Wallets', icon: Wallet },
      { href: '/admin/smpp-sessions', label: 'SMPP Sessions', icon: Plug },
    ],
  },
  {
    label: 'Pricing',
    items: [
      { href: '/admin/pricing', label: 'Pricing', icon: DollarSign },
      { href: '/admin/exchange-rates', label: 'Exchange Rates', icon: Building },
    ],
  },
  {
    label: 'System',
    items: [{ href: '/admin/settings', label: 'Settings', icon: Settings }],
  },
]

const SP_NAV_GROUPS: NavGroup[] = [
  {
    label: 'Overview',
    items: [{ href: '/sp/dashboard', label: 'Dashboard', icon: LayoutDashboard }],
  },
  {
    label: 'Messaging',
    items: [
      { href: '/sp/messages', label: 'Messages', icon: History },
      { href: '/sp/senders', label: 'Sender IDs', icon: Tag },
      { href: '/sp/templates', label: 'Templates', icon: FileText },
    ],
  },
  {
    label: 'Account',
    items: [
      { href: '/sp/api-keys', label: 'Connection', icon: Key },
      { href: '/sp/wallet', label: 'Wallet', icon: Wallet },
      { href: '/sp/analytics', label: 'Analytics', icon: BarChart3 },
    ],
  },
]

function NavItemLink({ item, isActive }: { item: NavItem; isActive: boolean }) {
  return (
    <Link
      href={item.href}
      className={cn(
        'flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors',
        isActive
          ? 'bg-white/10 text-white'
          : 'text-zinc-400 hover:text-white hover:bg-white/5'
      )}
    >
      <item.icon className="h-4 w-4" />
      {item.label}
    </Link>
  )
}

function NavGroupCollapsible({
  group,
  isExpanded,
  onToggle,
  variant,
}: {
  group: NavGroup
  isExpanded: boolean
  onToggle: () => void
  variant: 'admin' | 'sp'
}) {
  const pathname = usePathname()

  return (
    <div className="mb-1">
      <button
        onClick={onToggle}
        className={cn(
          'w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm font-medium transition-colors',
          isExpanded
            ? 'bg-white/10 text-white'
            : 'text-zinc-400 hover:text-white hover:bg-white/5'
        )}
      >
        <span className="flex items-center gap-2">
          <ChevronDown
            className={cn(
              'h-4 w-4 transition-transform',
              isExpanded ? 'rotate-180' : 'rotate-0'
            )}
          />
          {group.label}
        </span>
      </button>
      {isExpanded && (
        <div className="mt-1 ml-4 space-y-1">
          {group.items.map((item) => (
            <NavItemLink
              key={item.href}
              item={item}
              isActive={pathname === item.href}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function NavGroupSimple({
  group,
  variant,
}: {
  group: NavGroup
  variant: 'admin' | 'sp'
}) {
  const pathname = usePathname()

  return (
    <div className="mb-1">
      <p className="px-3 py-2 text-xs font-medium text-zinc-500 uppercase tracking-wider">
        {group.label}
      </p>
      <div className="space-y-1">
        {group.items.map((item) => (
          <NavItemLink
            key={item.href}
            item={item}
            isActive={pathname === item.href}
          />
        ))}
      </div>
    </div>
  )
}

function MobileOverlay({
  isOpen,
  onClose,
  variant,
  navGroups,
  user,
  wallet,
}: {
  isOpen: boolean
  onClose: () => void
  variant: 'admin' | 'sp'
  navGroups: NavGroup[]
  user?: { name: string; email: string }
  wallet?: { balance: number; currency: string }
}) {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 lg:hidden">
      <div
        className="absolute inset-0 bg-black/80 backdrop-blur-sm"
        onClick={onClose}
      />
      <aside className="absolute left-0 top-0 bottom-0 w-72 bg-zinc-900 border-r border-white/10 flex flex-col">
        <div className="flex items-center justify-between h-16 px-4 border-b border-white/10">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-white text-black">
              <MessageSquare className="h-5 w-5" />
            </div>
            <span className="text-lg font-semibold tracking-tight text-white">
              Aegisbox
            </span>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-zinc-400 hover:text-white hover:bg-white/5 rounded-lg"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <nav className="flex-1 overflow-y-auto p-4">
          {navGroups.map((group) =>
            group.items.length === 1 ? (
              <NavGroupSimple key={group.label} group={group} variant={variant} />
            ) : (
              <NavGroupCollapsible
                key={group.label}
                group={group}
                variant={variant}
                isExpanded={true}
                onToggle={() => {}}
              />
            )
          )}
        </nav>
        <div className="p-4 border-t border-white/10">
          {variant === 'sp' && wallet && (
            <div className="mb-4 p-3 rounded-lg bg-white/5 border border-white/10">
              <p className="text-xs text-zinc-400">Wallet Balance</p>
              <p className="text-lg font-semibold text-white">
                {formatMoney(wallet.balance, wallet.currency)}
              </p>
            </div>
          )}
          {user && (
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-full bg-white/10 flex items-center justify-center">
                <span className="text-sm font-medium text-white">
                  {user.name.charAt(0).toUpperCase()}
                </span>
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-white truncate">
                  {user.name}
                </p>
                <p className="text-xs text-zinc-400 truncate">{user.email}</p>
              </div>
            </div>
          )}
          <SignOutButton>
            <Button
              variant="ghost"
              className="w-full justify-start text-zinc-400 hover:text-white hover:bg-white/5 gap-2"
            >
              <Settings className="h-4 w-4" />
              Sign Out
            </Button>
          </SignOutButton>
        </div>
      </aside>
    </div>
  )
}

function SidebarContent({
  variant,
  isCollapsed,
  navGroups,
  user,
  wallet,
  onToggleCollapse,
}: {
  variant: 'admin' | 'sp'
  isCollapsed: boolean
  navGroups: NavGroup[]
  user?: { name: string; email: string }
  wallet?: { balance: number; currency: string }
  onToggleCollapse: () => void
}) {
  const pathname = usePathname()
  const [expandedGroups, setExpandedGroups] = useState<string[]>([])

  useEffect(() => {
    const saved = localStorage.getItem('sidebar-active-group')
    if (saved) {
      try {
        setExpandedGroups(JSON.parse(saved))
      } catch {
        setExpandedGroups([navGroups.find(g => g.items.length > 1)?.label || ''])
      }
    } else {
      setExpandedGroups([navGroups.find(g => g.items.length > 1)?.label || ''])
    }
  }, [navGroups])

  const toggleGroup = (label: string) => {
    setExpandedGroups((prev) => {
      const newExpanded = prev.includes(label)
        ? prev.filter((l) => l !== label)
        : [label]
      localStorage.setItem('sidebar-active-group', JSON.stringify(newExpanded))
      return newExpanded
    })
  }

  return (
    <>
      <div className="flex items-center justify-between h-16 px-4 border-b border-white/10">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-white text-black">
            <MessageSquare className="h-5 w-5" />
          </div>
          {!isCollapsed && (
            <span className="text-lg font-semibold tracking-tight text-white">
              Aegisbox
            </span>
          )}
        </div>
        <button
          onClick={onToggleCollapse}
          className={cn(
            'p-2 text-zinc-400 hover:text-white hover:bg-white/5 rounded-lg transition-colors',
            isCollapsed && 'mx-auto'
          )}
        >
          <ChevronRight
            className={cn(
              'h-4 w-4 transition-transform',
              isCollapsed ? 'rotate-180' : 'rotate-0'
            )}
          />
        </button>
      </div>

      <nav className="flex-1 overflow-y-auto p-4">
        {navGroups.map((group) =>
          group.items.length === 1 ? (
            <NavGroupSimple key={group.label} group={group} variant={variant} />
          ) : (
            <NavGroupCollapsible
              key={group.label}
              group={group}
              variant={variant}
              isExpanded={expandedGroups.includes(group.label)}
              onToggle={() => toggleGroup(group.label)}
            />
          )
        )}
      </nav>

      {!isCollapsed && variant === 'sp' && wallet && (
        <div className="p-4 border-t border-white/10">
          <div className="p-3 rounded-lg bg-white/5 border border-white/10">
            <p className="text-xs text-zinc-400">Wallet Balance</p>
            <p className="text-lg font-semibold text-white">
              {formatMoney(wallet.balance, wallet.currency)}
            </p>
          </div>
        </div>
      )}

      <div className="p-4 border-t border-white/10">
        {user && (
          <div className={cn('flex items-center gap-3', isCollapsed && 'justify-center')}>
            <div className="w-10 h-10 rounded-full bg-white/10 flex items-center justify-center flex-shrink-0">
              <span className="text-sm font-medium text-white">
                {user.name.charAt(0).toUpperCase()}
              </span>
            </div>
            {!isCollapsed && (
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-white truncate">
                  {user.name}
                </p>
                <p className="text-xs text-zinc-400 truncate">{user.email}</p>
              </div>
            )}
          </div>
        )}
        {!isCollapsed && (
          <SignOutButton>
            <Button
              variant="ghost"
              className="w-full mt-4 justify-start text-zinc-400 hover:text-white hover:bg-white/5 gap-2"
            >
              <Settings className="h-4 w-4" />
              Sign Out
            </Button>
          </SignOutButton>
        )}
      </div>
    </>
  )
}

export function DashboardSidebar({ variant = 'admin' }: { variant?: 'admin' | 'sp' }) {
  const { user, isLoaded: userLoaded } = useUser()
  const { organization, isLoaded: orgLoaded } = useOrganization()
  const [isCollapsed, setIsCollapsed] = useState(false)
  const [isMobileOpen, setIsMobileOpen] = useState(false)

  const navGroups = variant === 'admin' ? ADMIN_NAV_GROUPS : SP_NAV_GROUPS

  const isLoaded = userLoaded && orgLoaded

  const userData = user
    ? {
        name: organization?.name || user.firstName || 'User',
        email: user.emailAddresses[0]?.emailAddress || '',
      }
    : undefined

  // The provider's real balance; nothing is shown until it has loaded.
  const [walletData, setWalletData] = useState<{ balance: number; currency: string } | undefined>(undefined)
  useEffect(() => {
    if (variant !== 'sp') return
    let cancelled = false
    fetch('/api/sp/wallet')
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        const w = data?.wallets?.[0]
        if (!cancelled && w) setWalletData({ balance: Number(w.balance), currency: w.currencyCode })
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [variant])

  useEffect(() => {
    const saved = localStorage.getItem('sidebar-collapsed')
    if (saved === 'true') {
      setIsCollapsed(true)
    }
  }, [])

  const toggleCollapse = () => {
    const newCollapsed = !isCollapsed
    setIsCollapsed(newCollapsed)
    localStorage.setItem('sidebar-collapsed', String(newCollapsed))
  }

  return (
    <>
      <button
        onClick={() => setIsMobileOpen(true)}
        className="fixed top-4 left-4 z-40 p-2 text-zinc-400 hover:text-white hover:bg-white/5 rounded-lg lg:hidden"
      >
        <Menu className="h-5 w-5" />
      </button>

      <MobileOverlay
        isOpen={isMobileOpen}
        onClose={() => setIsMobileOpen(false)}
        variant={variant}
        navGroups={navGroups}
        user={userData}
        wallet={walletData}
      />

      <aside
        className={cn(
          'fixed top-0 left-0 h-full bg-black border-r border-white/10 z-30 transition-all duration-300 hidden lg:flex flex-col',
          isCollapsed ? 'w-16' : 'w-64'
        )}
      >
        {isLoaded ? (
          <SidebarContent
            variant={variant}
            isCollapsed={isCollapsed}
            navGroups={navGroups}
            user={userData}
            wallet={walletData}
            onToggleCollapse={toggleCollapse}
          />
        ) : (
          <div className="flex items-center justify-center h-full">
            <Loader2 className="h-6 w-6 animate-spin text-zinc-400" />
          </div>
        )}
      </aside>
    </>
  )
}
