import {
  LayoutDashboard, Server, Key, FileText, Bell, CalendarDays,
  BarChart3, Settings, Terminal, Activity, Bot, type LucideIcon,
} from 'lucide-react'

export interface NavItem {
  id: string
  label: string
  icon: LucideIcon
  path: string
}

export const navMain: NavItem[] = [
  { id: 'dashboard', label: '总览', icon: LayoutDashboard, path: '/' },
  { id: 'providers', label: 'Providers', icon: Server, path: '/providers' },
  { id: 'keys', label: 'API Keys', icon: Key, path: '/keys' },
  { id: 'usage', label: '用量日志', icon: FileText, path: '/usage' },
]

export const navMore: NavItem[] = [
  { id: 'sessions', label: '会话', icon: Activity, path: '/sessions' },
  { id: 'agents', label: 'Agent', icon: Bot, path: '/agents' },
  { id: 'alerts', label: '告警', icon: Bell, path: '/alerts' },
  { id: 'subscriptions', label: '订阅', icon: CalendarDays, path: '/subscriptions' },
  { id: 'frequency', label: '频率', icon: BarChart3, path: '/frequency' },
  { id: 'playground', label: 'Playground', icon: Terminal, path: '/playground' },
]

export const navBottom: NavItem[] = [
  { id: 'settings', label: '设置', icon: Settings, path: '/settings' },
]

export const allNav = [...navMain, ...navMore, ...navBottom]
