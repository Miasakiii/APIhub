import {
  LayoutDashboard, Server, Key, FileText, Bell, CalendarDays,
  BarChart3, Settings, Terminal, type LucideIcon,
} from 'lucide-react'

export interface NavItem {
  id: string
  label: string
  icon: LucideIcon
}

export const navMain: NavItem[] = [
  { id: 'dashboard', label: '总览', icon: LayoutDashboard },
  { id: 'providers', label: 'Providers', icon: Server },
  { id: 'keys', label: 'API Keys', icon: Key },
  { id: 'usage', label: '用量日志', icon: FileText },
]

export const navMore: NavItem[] = [
  { id: 'alerts', label: '告警', icon: Bell },
  { id: 'subscriptions', label: '订阅', icon: CalendarDays },
  { id: 'frequency', label: '频率', icon: BarChart3 },
  { id: 'playground', label: 'Playground', icon: Terminal },
]

export const navBottom: NavItem[] = [{ id: 'settings', label: '设置', icon: Settings }]

export const allNav = [...navMain, ...navMore, ...navBottom]
