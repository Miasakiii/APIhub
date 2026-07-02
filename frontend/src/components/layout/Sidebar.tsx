import { NavLink } from 'react-router-dom'
import { X, Sparkles, Moon, Sun, ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '../../lib/utils'
import { useTheme } from '../../lib/use-theme'
import { useCompactMode } from '../../lib/compact-mode'
import { navMain, navBottom } from '../../lib/nav'

interface SidebarProps {
  open: boolean
  onClose: () => void
}

export function Sidebar({ open, onClose }: SidebarProps) {
  const { theme, toggle } = useTheme()
  const { mode, setMode } = useCompactMode()
  const isCollapsed = mode === 'icons'

  function NavItem({ path, label, icon: Icon }: { path: string; label: string; icon: React.ComponentType<{ className?: string }> }) {
    return (
      <NavLink
        to={path}
        end={path === '/'}
        onClick={onClose}
        className={({ isActive }) => cn(
          'w-full flex items-center gap-3 rounded-xl text-sm font-medium transition-all duration-200 relative group',
          isCollapsed ? 'justify-center px-2 py-2.5' : 'px-3 py-2.5',
          isActive ? 'bg-indigo-50 dark:bg-indigo-950/30 text-indigo-600 dark:text-indigo-400' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-white/5',
        )}
        title={isCollapsed ? label : undefined}
      >
        {({ isActive }) => (
          <>
            {isActive && !isCollapsed && <span className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-r-full bg-indigo-500 dark:bg-indigo-400" />}
            {isActive && isCollapsed && <span className="absolute left-1/2 -translate-x-1/2 bottom-0 w-5 h-[3px] rounded-t-full bg-indigo-500 dark:bg-indigo-400" />}
            <Icon className={cn('w-5 h-5 shrink-0', isActive ? 'text-indigo-600 dark:text-indigo-400' : 'text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200')} />
            {!isCollapsed && <span>{label}</span>}
          </>
        )}
      </NavLink>
    )
  }

  return (
    <>
      {open && <div className="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-30 lg:hidden" onClick={onClose} />}

      <aside className={cn(
        'fixed lg:static inset-y-0 left-0 z-40 flex flex-col bg-slate-50 dark:bg-slate-900/50 border-r border-slate-200 dark:border-slate-700/50',
        'transition-all duration-300 ease-out',
        isCollapsed ? 'w-16' : 'w-64',
        open ? 'translate-x-0' : '-translate-x-full lg:translate-x-0',
      )}>
        <div className={cn(
          'h-16 flex items-center border-b border-slate-200 dark:border-slate-700/50 shrink-0',
          isCollapsed ? 'justify-center px-2' : 'gap-3 px-5',
        )}>
          <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center shadow-lg shadow-indigo-500/30 shrink-0">
            <Sparkles className="w-4 h-4 text-white" />
          </div>
          {!isCollapsed && (
            <div className="flex-1 min-w-0">
              <span className="font-bold text-slate-900 dark:text-white text-base tracking-tight">APIHub</span>
              <p className="text-[10px] text-slate-500 dark:text-slate-500 uppercase tracking-widest">Monitor</p>
            </div>
          )}
          <button
            type="button"
            className={cn(
              'text-slate-500 hover:text-slate-900 dark:hover:text-white p-1 rounded-lg hover:bg-slate-200 dark:hover:bg-white/10 transition-colors',
              isCollapsed ? 'hidden' : 'lg:hidden',
            )}
            onClick={onClose}
            aria-label="关闭"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <nav className={cn('flex-1 overflow-y-auto', isCollapsed ? 'p-2' : 'p-3 space-y-6')}>
          {isCollapsed ? (
            <div className="space-y-1">{navMain.map((n) => <NavItem key={n.id} {...n} />)}</div>
          ) : (
            <div>
              <div className="space-y-0.5">{navMain.map((n) => <NavItem key={n.id} {...n} />)}</div>
            </div>
          )}
        </nav>

        <div className={cn('p-3 border-t border-slate-200 dark:border-slate-700/50 space-y-1', isCollapsed && 'p-2')}>
          {isCollapsed ? (
            <>
              {navBottom.map((n) => <NavItem key={n.id} {...n} />)}
              <button
                type="button"
                onClick={toggle}
                className="w-full flex items-center justify-center px-2 py-2.5 rounded-xl text-sm font-medium text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-white/5 transition-all duration-200"
                title={theme === 'dark' ? '亮色模式' : '暗色模式'}
              >
                {theme === 'dark' ? <Sun className="w-5 h-5 text-slate-500" /> : <Moon className="w-5 h-5 text-slate-500" />}
              </button>
            </>
          ) : (
            <>
              {navBottom.map((n) => <NavItem key={n.id} {...n} />)}
              <button type="button" onClick={toggle} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-white/5 transition-all duration-200">
                {theme === 'dark' ? <Sun className="w-[18px] h-[18px] text-slate-500" /> : <Moon className="w-[18px] h-[18px] text-slate-500" />}
                <span>{theme === 'dark' ? '亮色模式' : '暗色模式'}</span>
              </button>
                <div className="px-3 pt-2"><p className="text-[10px] text-slate-400 dark:text-slate-600">APIHub v0.4</p></div>
            </>
          )}
        </div>

        <button
          type="button"
          onClick={() => setMode(isCollapsed ? 'full' : 'icons')}
          className={cn(
            'absolute top-1/2 -translate-y-1/2 w-6 h-12 flex items-center justify-center',
            'bg-slate-200 dark:bg-slate-800 hover:bg-slate-300 dark:hover:bg-slate-700 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white',
            'rounded-r-lg transition-all duration-200',
            'right-0',
          )}
          aria-label={isCollapsed ? '展开侧边栏' : '折叠侧边栏'}
        >
          {isCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
        </button>
      </aside>
    </>
  )
}
