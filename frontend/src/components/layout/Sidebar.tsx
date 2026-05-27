import { X, Sparkles, Moon, Sun } from 'lucide-react'
import { cn } from '../../lib/utils'
import { useTheme } from '../../lib/use-theme'
import { navMain, navMore, navBottom } from '../../lib/nav'

interface SidebarProps {
  page: string
  setPage: (p: string) => void
  open: boolean
  onClose: () => void
}

export function Sidebar({ page, setPage, open, onClose }: SidebarProps) {
  const { theme, toggle } = useTheme()

  function NavItem({ id, label, icon: Icon }: { id: string; label: string; icon: React.ComponentType<{ className?: string }> }) {
    const active = page === id
    return (
      <button
        type="button"
        onClick={() => { setPage(id); onClose() }}
        className={cn(
          'w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 relative group',
          active ? 'bg-white/10 text-white' : 'text-slate-400 hover:text-white hover:bg-white/5',
        )}
      >
        {active && <span className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-r-full bg-gradient-to-b from-indigo-400 to-violet-400" />}
        <Icon className={cn('w-[18px] h-[18px] shrink-0', active ? 'text-indigo-300' : 'text-slate-500 group-hover:text-slate-300')} />
        <span>{label}</span>
      </button>
    )
  }

  return (
    <>
      {open && <div className="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-30 lg:hidden" onClick={onClose} />}

      <aside className={cn(
        'fixed lg:static inset-y-0 left-0 z-40 w-64 flex flex-col sidebar-bg border-r border-white/[0.06]',
        'transition-transform duration-300 ease-out',
        open ? 'translate-x-0' : '-translate-x-full lg:translate-x-0',
      )}>
        <div className="h-16 flex items-center gap-3 px-5 border-b border-white/[0.06] shrink-0">
          <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center shadow-lg shadow-indigo-500/30">
            <Sparkles className="w-4 h-4 text-white" />
          </div>
          <div className="flex-1 min-w-0">
            <span className="font-bold text-white text-base tracking-tight">APIHub</span>
            <p className="text-[10px] text-slate-500 uppercase tracking-widest">Monitor</p>
          </div>
          <button type="button" className="lg:hidden text-slate-400 hover:text-white p-1" onClick={onClose} aria-label="关闭">
            <X className="w-5 h-5" />
          </button>
        </div>

        <nav className="flex-1 overflow-y-auto p-3 space-y-6">
          <div>
            <p className="px-3 mb-2 text-[10px] font-semibold uppercase tracking-wider text-slate-600">数据</p>
            <div className="space-y-0.5">{navMain.map((n) => <NavItem key={n.id} {...n} />)}</div>
          </div>
          <div>
            <p className="px-3 mb-2 text-[10px] font-semibold uppercase tracking-wider text-slate-600">工具</p>
            <div className="space-y-0.5">{navMore.map((n) => <NavItem key={n.id} {...n} />)}</div>
          </div>
        </nav>

        <div className="p-3 border-t border-white/[0.06] space-y-1">
          {navBottom.map((n) => <NavItem key={n.id} {...n} />)}
          <button type="button" onClick={toggle} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium text-slate-400 hover:text-white hover:bg-white/5 transition-all duration-200">
            {theme === 'dark' ? <Sun className="w-[18px] h-[18px] text-slate-500" /> : <Moon className="w-[18px] h-[18px] text-slate-500" />}
            <span>{theme === 'dark' ? '亮色模式' : '暗色模式'}</span>
          </button>
          <div className="px-3 pt-2"><p className="text-[10px] text-slate-600">APIHub v0.2</p></div>
        </div>
      </aside>
    </>
  )
}
