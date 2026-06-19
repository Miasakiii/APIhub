import { useLocation, useParams } from 'react-router-dom'
import { Menu, LogOut } from 'lucide-react'
import { cn } from '../../lib/utils'
import { allNav } from '../../lib/nav'

interface TopBarProps {
  authEnabled: boolean
  onLogout?: () => void
  onMenuOpen: () => void
}

export function TopBar({ authEnabled, onLogout, onMenuOpen }: TopBarProps) {
  const location = useLocation()
  const params = useParams()

  const currentLabel = location.pathname.startsWith('/model/')
    ? (params.model ?? '')
    : allNav.find((n) => n.path === location.pathname)?.label ?? ''

  return (
    <header className="h-16 shrink-0 flex items-center gap-4 px-4 lg:px-8 border-b border-slate-200/60 dark:border-slate-700/60 bg-white/70 dark:bg-slate-900/70 backdrop-blur-xl">
      <button
        type="button"
        className="lg:hidden p-2 -ml-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400"
        onClick={onMenuOpen}
        aria-label="菜单"
      >
        <Menu className="w-5 h-5" />
      </button>
      <div className="flex-1 min-w-0">
        <h1 className="text-lg font-semibold text-slate-900 dark:text-slate-100 truncate">{currentLabel}</h1>
      </div>
      {authEnabled && onLogout && (
        <button
          type="button"
          onClick={onLogout}
          className={cn(
            'flex items-center gap-2 px-3 py-2 rounded-xl text-sm transition',
            'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800',
          )}
        >
          <LogOut className="w-4 h-4" />
          <span className="hidden sm:inline">退出</span>
        </button>
      )}
    </header>
  )
}
