import {
  LayoutDashboard, FileText, BarChart3, Settings, type LucideIcon,
} from 'lucide-react'

export interface NavItem {
  id: string
  label: string
  icon: LucideIcon
  path: string
}

export const navMain: NavItem[] = [
  { id: 'dashboard', label: '总览', icon: LayoutDashboard, path: '/' },
  { id: 'usage', label: '用量日志', icon: FileText, path: '/usage' },
  { id: 'frequency', label: '频率', icon: BarChart3, path: '/frequency' },
]

export const navBottom: NavItem[] = [
  { id: 'settings', label: '设置', icon: Settings, path: '/settings' },
]

export const allNav = [...navMain, ...navBottom]
