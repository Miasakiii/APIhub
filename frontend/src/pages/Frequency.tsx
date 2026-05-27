import { useEffect, useState } from 'react'
import { Zap, Clock, Calendar, BarChart3 } from 'lucide-react'
import { api } from '../api'
import { cn } from '../lib/utils'
import { PageHeader, Card, CardHeader, Skeleton } from '../components/ui'

export function Frequency() {
  const [hourly, setHourly] = useState<number[][]>([])
  const [peakQps, setPeakQps] = useState({ peak_qps: 0, peak_minute: '', avg_qps: 0, peak_count: 0 })
  const [today, setToday] = useState<number[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    Promise.all([api.frequency.hourly(), api.frequency.peakQps(), api.frequency.today()])
      .then(([h, p, t]) => {
        if (cancelled) return
        setHourly(h.heatmap)
        setPeakQps(p)
        setToday(t.hourly)
      })
      .catch((e) => console.error(e))
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  const maxHeat = Math.max(...hourly.flat(), 0)
  const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']

  if (loading) return <LoadingSkeleton />

  return (
    <div className="space-y-6">
      <PageHeader title="频率分析" description="API 请求频率与分布" />

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <FreqStatCard icon={Zap} label="峰值 QPS" value={peakQps.peak_qps.toFixed(2)} subtext={`${peakQps.peak_count} 请求`} color="amber" />
        <FreqStatCard icon={Clock} label="峰值时刻" value={peakQps.peak_minute || '--'} subtext="每分钟请求数" color="blue" />
        <FreqStatCard icon={BarChart3} label="平均 QPS" value={peakQps.avg_qps.toFixed(2)} subtext="过去 7 天" color="emerald" />
      </div>

      <Card>
        <CardHeader title="今日每小时分布" description="24 小时请求量" action={<Calendar className="w-4 h-4 text-slate-400" />} />
        <div className="flex items-end gap-1 h-48">
          {today.map((count, hour) => (
            <div key={hour} className="flex-1 flex flex-col items-center gap-1">
              <div
                className="w-full bg-gradient-to-t from-blue-500 to-blue-400 dark:from-blue-600 dark:to-blue-500 rounded-t-md transition-all hover:from-blue-600 hover:to-blue-500 min-h-[2px]"
                style={{ height: `${Math.max((count / Math.max(...today, 1)) * 100, 2)}%` }}
                title={`${hour}:00 - ${count} 次`}
              />
              <span className="text-[10px] text-slate-400 dark:text-slate-500">{hour}</span>
            </div>
          ))}
        </div>
      </Card>

      <Card>
        <CardHeader title="7 天小时热力图" description="一周内每小时请求分布" action={<BarChart3 className="w-4 h-4 text-slate-400" />} />
        <div className="overflow-x-auto">
          <div className="min-w-[800px]">
            <div className="flex mb-1">
              <div className="w-12" />
              {Array.from({ length: 24 }, (_, i) => <div key={i} className="flex-1 text-[10px] text-slate-400 dark:text-slate-500 text-center">{i}</div>)}
            </div>
            {hourly.map((day, dayIdx) => (
              <div key={dayIdx} className="flex items-center mb-1">
                <div className="w-12 text-xs text-slate-500 dark:text-slate-400 font-medium">{dayNames[dayIdx]}</div>
                <div className="flex flex-1 gap-0.5">
                  {day.map((count, hour) => (
                    <div key={hour} className="flex-1 aspect-square rounded-sm transition hover:ring-2 hover:ring-blue-400 hover:z-10"
                      style={{ backgroundColor: getHeatColor(count, maxHeat) }} title={`${dayNames[dayIdx]} ${hour}:00 - ${count} 次`} />
                  ))}
                </div>
              </div>
            ))}
            <div className="flex items-center gap-2 mt-4 justify-end">
              <span className="text-xs text-slate-400">少</span>
              {['#f0fdf4', '#dcfce7', '#bbf7d0', '#86efac', '#4ade80', '#22c55e', '#16a34a', '#15803d'].map((c, i) => <div key={i} className="w-4 h-4 rounded-sm" style={{ backgroundColor: c }} />)}
              <span className="text-xs text-slate-400">多</span>
            </div>
          </div>
        </div>
      </Card>
    </div>
  )
}

function FreqStatCard({ icon: Icon, label, value, subtext, color }: {
  icon: React.ComponentType<{ className?: string }>; label: string; value: string; subtext: string; color: string
}) {
  const colorMap: Record<string, { bg: string; text: string }> = {
    emerald: { bg: 'bg-emerald-50 dark:bg-emerald-950/40', text: 'text-emerald-600 dark:text-emerald-400' },
    blue: { bg: 'bg-blue-50 dark:bg-blue-950/40', text: 'text-blue-600 dark:text-blue-400' },
    amber: { bg: 'bg-amber-50 dark:bg-amber-950/40', text: 'text-amber-600 dark:text-amber-400' },
  }
  const c = colorMap[color] || colorMap.emerald
  return (
    <div className="bg-white dark:bg-slate-900/80 rounded-2xl border border-slate-200 dark:border-slate-700/60 p-5 shadow-sm hover:shadow-md dark:hover:shadow-lg transition-shadow duration-300 overflow-hidden relative">
      <div className={cn('absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r', color === 'amber' ? 'from-amber-500 to-yellow-400' : color === 'blue' ? 'from-blue-500 to-cyan-400' : 'from-emerald-500 to-teal-400')} />
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-1">{label}</p>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">{value}</p>
          <p className="text-xs text-slate-400 dark:text-slate-500 mt-1">{subtext}</p>
        </div>
        <div className={cn('w-10 h-10 rounded-xl flex items-center justify-center', c.bg, c.text)}>
          <Icon className="w-5 h-5" />
        </div>
      </div>
    </div>
  )
}

function getHeatColor(count: number, max: number): string {
  if (max === 0) return '#f0fdf4'
  const ratio = count / max
  const colors = ['#f0fdf4', '#dcfce7', '#bbf7d0', '#86efac', '#4ade80', '#22c55e', '#16a34a', '#15803d']
  return colors[Math.min(Math.floor(ratio * (colors.length - 1)), colors.length - 1)]
}

function LoadingSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <Skeleton className="h-8 w-48" />
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {[1, 2, 3].map(i => <Skeleton key={i} className="h-28" />)}
      </div>
      <Skeleton className="h-64" />
    </div>
  )
}
