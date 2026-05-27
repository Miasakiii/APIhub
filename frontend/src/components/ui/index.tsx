import { cn } from '../../lib/utils'
import type { LucideIcon } from 'lucide-react'

export { Modal } from './Modal'
export { ToastProvider } from './Toast'
export { Tabs } from './Tabs'

/* ────────────────────────── Card ────────────────────────── */

export function Card({
  children,
  className,
  padding = true,
}: {
  children: React.ReactNode
  className?: string
  padding?: boolean
}) {
  return (
    <div
      className={cn(
        'bg-white/90 dark:bg-slate-900/80 backdrop-blur-sm rounded-2xl border border-slate-200/80 dark:border-slate-700/60 shadow-sm shadow-slate-200/50 dark:shadow-black/10',
        'hover:shadow-md dark:hover:shadow-lg transition-shadow duration-300',
        padding && 'p-6',
        className,
      )}
    >
      {children}
    </div>
  )
}

export function CardHeader({
  title,
  description,
  action,
}: {
  title: string
  description?: string
  action?: React.ReactNode
}) {
  return (
    <div className="flex items-start justify-between gap-4 mb-6">
      <div>
        <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">{title}</h2>
        {description && <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{description}</p>}
      </div>
      {action}
    </div>
  )
}

/* ────────────────────────── PageHeader ────────────────────────── */

export function PageHeader({
  title,
  description,
  actions,
}: {
  title: string
  description?: string
  actions?: React.ReactNode
}) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-2">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">{title}</h1>
        {description && <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">{description}</p>}
      </div>
      {actions && <div className="flex flex-wrap gap-2 shrink-0">{actions}</div>}
    </div>
  )
}

/* ────────────────────────── Button ────────────────────────── */

export function Button({
  children,
  variant = 'primary',
  size = 'md',
  loading = false,
  className,
  disabled,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md'
  loading?: boolean
}) {
  const base =
    'inline-flex items-center justify-center gap-2 font-medium rounded-xl transition-all duration-200 disabled:opacity-50 disabled:pointer-events-none cursor-pointer'
  const variants = {
    primary:
      'bg-gradient-to-r from-indigo-600 to-violet-600 text-white shadow-md shadow-indigo-500/25 hover:shadow-lg hover:shadow-indigo-500/30 hover:from-indigo-500 hover:to-violet-500',
    secondary:
      'bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700/80 shadow-sm hover:bg-slate-50 dark:hover:bg-slate-750 hover:border-slate-300 dark:hover:border-slate-600',
    ghost: 'text-slate-600 dark:text-slate-400 hover:bg-slate-100/80 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200',
    danger: 'bg-red-600 text-white hover:bg-red-700 shadow-sm',
  }
  const sizes = {
    sm: 'px-3 py-1.5 text-xs',
    md: 'px-4 py-2.5 text-sm',
  }
  return (
    <button
      type="button"
      className={cn(base, variants[variant], sizes[size], className)}
      disabled={disabled || loading}
      {...props}
    >
      {loading && (
        <svg className="animate-spin -ml-0.5 h-3.5 w-3.5" viewBox="0 0 24 24" fill="none">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      )}
      {children}
    </button>
  )
}

/* ────────────────────────── Input ────────────────────────── */

export function Input({
  className,
  ...props
}: React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        'w-full border border-slate-200 dark:border-slate-700 bg-white/80 dark:bg-slate-800/80 rounded-xl px-3.5 py-2.5 text-sm text-slate-900 dark:text-slate-100',
        'placeholder:text-slate-400 dark:placeholder:text-slate-500',
        'focus:outline-none focus:ring-2 focus:ring-indigo-500/30 dark:focus:ring-indigo-400/30 focus:border-indigo-400 dark:focus:border-indigo-500',
        'transition-shadow',
        className,
      )}
      {...props}
    />
  )
}

export function Select({
  className,
  children,
  ...props
}: React.SelectHTMLAttributes<HTMLSelectElement> & { children: React.ReactNode }) {
  return (
    <select
      className={cn(
        'w-full border border-slate-200 dark:border-slate-700 bg-white/80 dark:bg-slate-800/80 rounded-xl px-3.5 py-2.5 text-sm text-slate-900 dark:text-slate-100',
        'focus:outline-none focus:ring-2 focus:ring-indigo-500/30 dark:focus:ring-indigo-400/30 focus:border-indigo-400 dark:focus:border-indigo-500',
        'transition-shadow appearance-none',
        className,
      )}
      {...props}
    >
      {children}
    </select>
  )
}

/* ────────────────────────── Badge ────────────────────────── */

export function Badge({
  children,
  variant = 'default',
}: {
  children: React.ReactNode
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'muted' | 'info'
}) {
  const styles = {
    default: 'bg-indigo-50 text-indigo-700 border-indigo-100 dark:bg-indigo-950/50 dark:text-indigo-300 dark:border-indigo-800/50',
    success: 'bg-emerald-50 text-emerald-700 border-emerald-100 dark:bg-emerald-950/50 dark:text-emerald-300 dark:border-emerald-800/50',
    warning: 'bg-amber-50 text-amber-700 border-amber-100 dark:bg-amber-950/50 dark:text-amber-300 dark:border-amber-800/50',
    danger: 'bg-red-50 text-red-700 border-red-100 dark:bg-red-950/50 dark:text-red-300 dark:border-red-800/50',
    info: 'bg-blue-50 text-blue-700 border-blue-100 dark:bg-blue-950/50 dark:text-blue-300 dark:border-blue-800/50',
    muted: 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700',
  }
  return (
    <span className={cn('inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium border', styles[variant])}>
      {children}
    </span>
  )
}

/* ────────────────────────── StatCard ────────────────────────── */

export function StatCard({
  icon: Icon,
  label,
  value,
  change,
  accent = 'indigo',
}: {
  icon: LucideIcon
  label: string
  value: string
  change?: number
  accent?: 'indigo' | 'emerald' | 'violet' | 'amber'
}) {
  const accents = {
    indigo: {
      icon: 'bg-indigo-50 text-indigo-600 ring-indigo-100 dark:bg-indigo-950/60 dark:text-indigo-400 dark:ring-indigo-800/50',
      glow: 'from-indigo-500/5 dark:from-indigo-400/10',
      bar: 'from-indigo-500 to-indigo-300',
    },
    emerald: {
      icon: 'bg-emerald-50 text-emerald-600 ring-emerald-100 dark:bg-emerald-950/60 dark:text-emerald-400 dark:ring-emerald-800/50',
      glow: 'from-emerald-500/5 dark:from-emerald-400/10',
      bar: 'from-emerald-500 to-emerald-300',
    },
    violet: {
      icon: 'bg-violet-50 text-violet-600 ring-violet-100 dark:bg-violet-950/60 dark:text-violet-400 dark:ring-violet-800/50',
      glow: 'from-violet-500/5 dark:from-violet-400/10',
      bar: 'from-violet-500 to-violet-300',
    },
    amber: {
      icon: 'bg-amber-50 text-amber-600 ring-amber-100 dark:bg-amber-950/60 dark:text-amber-400 dark:ring-amber-800/50',
      glow: 'from-amber-500/5 dark:from-amber-400/10',
      bar: 'from-amber-500 to-amber-300',
    },
  }
  const a = accents[accent]
  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl border border-slate-200/80 dark:border-slate-700/60 bg-white/90 dark:bg-slate-900/80 p-5 shadow-sm',
        'hover:shadow-md dark:hover:shadow-lg hover:border-slate-200 dark:hover:border-slate-600 transition-all duration-300',
      )}
    >
      {/* Top accent bar */}
      <div className={cn('absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r', a.bar)} />
      <div className={cn('absolute inset-0 bg-gradient-to-br to-transparent pointer-events-none', a.glow)} />
      <div className="relative flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <p className="text-[11px] font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">{label}</p>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 tracking-tight mt-1 tabular-nums">{value}</p>
          {change !== undefined && (
            <p
              className={cn(
                'text-xs font-medium mt-2 flex items-center gap-0.5',
                change >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400',
              )}
            >
              {change >= 0 ? '↑' : '↓'} {Math.abs(change).toFixed(1)}%
              <span className="text-slate-400 dark:text-slate-500 font-normal ml-1">vs 上周</span>
            </p>
          )}
        </div>
        <div className={cn('w-11 h-11 rounded-xl flex items-center justify-center ring-1 shrink-0', a.icon)}>
          <Icon className="w-5 h-5" strokeWidth={2} />
        </div>
      </div>
    </div>
  )
}

/* ────────────────────────── EmptyState ────────────────────────── */

export function EmptyState({
  icon: Icon,
  title,
  description,
}: {
  icon: LucideIcon
  title: string
  description?: string
}) {
  return (
    <div className="flex flex-col items-center justify-center py-14 text-center">
      <div className="w-14 h-14 rounded-2xl bg-slate-100 dark:bg-slate-800 flex items-center justify-center mb-4">
        <Icon className="w-6 h-6 text-slate-400 dark:text-slate-500" />
      </div>
      <p className="text-sm font-medium text-slate-700 dark:text-slate-300">{title}</p>
      {description && <p className="text-xs text-slate-400 dark:text-slate-500 mt-1 max-w-xs">{description}</p>}
    </div>
  )
}

/* ────────────────────────── PageLoader ────────────────────────── */

export function PageLoader() {
  return (
    <div className="flex items-center justify-center py-24">
      <div className="flex flex-col items-center gap-3">
        <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500 to-violet-600 animate-pulse shadow-lg shadow-indigo-500/30" />
        <p className="text-sm text-slate-500 dark:text-slate-400">加载中...</p>
      </div>
    </div>
  )
}

/* ────────────────────────── Skeleton ────────────────────────── */

export function Skeleton({ className }: { className?: string }) {
  return <div className={cn('skeleton-wave rounded-xl', className)} />
}
