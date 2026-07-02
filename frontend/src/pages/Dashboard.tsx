import { useEffect, useState, useCallback, useRef } from 'react'

import { DollarSign, BarChart3, Activity, Cpu, ArrowUp, ArrowDown, Minus } from 'lucide-react'
import { AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { api } from '../api'
import { formatUSD, formatNum } from '../lib/utils'
import { useWSMessage } from '../lib/use-ws'
import { Button, Card, PageHeader, StatCard as MetricCard, Skeleton } from '../components/ui'

interface Summary {
  total_cost_usd: number
  total_tokens: number
  total_requests: number
  unique_models: number
  unique_keys: number
}

interface TrendPoint {
  date: string
  cost_usd: number
  tokens: number
  request_count: number
}

interface ModelBreakdown {
  model: string
  total_cost_usd: number
  total_tokens: number
  request_count: number
}

interface UsageRecord {
  id: string
  model: string
  cost_usd: number
  input_tokens: number
  output_tokens: number
  timestamp: string
}

export function Dashboard() {
  const [summary, setSummary] = useState<Summary | null>(null)
  const [breakdown, setBreakdown] = useState<ModelBreakdown[]>([])
  const [trend, setTrend] = useState<TrendPoint[]>([])
  const [recentUsage, setRecentUsage] = useState<UsageRecord[]>([])
  const [loading, setLoading] = useState(true)
  const refreshTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  const fetchData = useCallback(() => {
    Promise.all([
      api.usage.summary().catch(() => null),
      api.stats.modelBreakdown().catch(() => []),
      api.stats.costTrend().catch(() => []),
      api.usage.list({ page_size: 5 }).catch(() => ({ records: [], total: 0, page: 1, page_size: 5 })),
    ]).then(([s, b, t, u]) => {
      setSummary(s)
      setBreakdown(b)
      setTrend(t)
      setRecentUsage(u.records)
    }).catch((e) => console.error(e)).finally(() => setLoading(false))
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  useWSMessage('usage.update', () => {
    clearTimeout(refreshTimer.current)
    refreshTimer.current = setTimeout(fetchData, 500)
  })

  if (loading) return <LoadingSkeleton />

  const totalCost = summary?.total_cost_usd ?? 0
  const totalTokens = summary?.total_tokens ?? 0
  const totalRequests = summary?.total_requests ?? 0

  const todayCost = trend.length >= 1 ? trend[trend.length - 1]?.cost_usd ?? 0 : 0
  const yesterdayCost = trend.length >= 2 ? trend[trend.length - 2]?.cost_usd ?? 0 : 0
  const costDelta = yesterdayCost > 0 ? ((todayCost - yesterdayCost) / yesterdayCost) * 100 : 0

  return (
    <div className="space-y-6">
      <PageHeader
        title="数据总览"
        description="实时监控 API 用量、费用与模型分布"
        actions={
          <Button variant="secondary" onClick={() => { setLoading(true); fetchData() }}>
            刷新
          </Button>
        }
      />

      {/* Row 1: Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="relative">
          <MetricCard icon={DollarSign} label="今日费用" value={formatUSD(totalCost)} accent="emerald" />
          {trend.length >= 2 && (
            <div className="absolute top-3 right-3 flex items-center gap-1 text-xs font-medium">
              {costDelta > 0 ? (
                <span className="flex items-center gap-0.5 text-emerald-600">
                  <ArrowUp className="w-3 h-3" />{costDelta.toFixed(0)}%
                </span>
              ) : costDelta < 0 ? (
                <span className="flex items-center gap-0.5 text-rose-500">
                  <ArrowDown className="w-3 h-3" />{Math.abs(costDelta).toFixed(0)}%
                </span>
              ) : (
                <span className="flex items-center gap-0.5 text-slate-400">
                  <Minus className="w-3 h-3" />0%
                </span>
              )}
            </div>
          )}
        </div>
        <MetricCard icon={BarChart3} label="总 Tokens" value={formatNum(totalTokens)} accent="indigo" />
        <MetricCard icon={Activity} label="总请求数" value={formatNum(totalRequests)} accent="violet" />
        <MetricCard icon={Cpu} label="模型数" value={String(summary?.unique_models ?? 0)} accent="amber" />
      </div>

      {/* Row 2: Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Cost Trend */}
        <Card>
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">费用趋势</h2>
              <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">最近 7 天费用变化</p>
            </div>
          </div>
          {trend.length > 0 ? (
            <ResponsiveContainer width="100%" height={200}>
              <AreaChart data={trend}>
                <defs>
                  <linearGradient id="costGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#6366f1" stopOpacity={0.2} />
                    <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#94a3b8' }} tickFormatter={(d) => d.slice(5)} />
                <YAxis tick={{ fontSize: 11, fill: '#94a3b8' }} tickFormatter={(v) => `$${v.toFixed(0)}`} />
                <Tooltip formatter={(v) => [`$${Number(v).toFixed(2)}`, '费用']} />
                <Area type="monotone" dataKey="cost_usd" stroke="#6366f1" strokeWidth={2} fill="url(#costGradient)" />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex items-center justify-center h-[200px] text-sm text-slate-400">暂无数据</div>
          )}
        </Card>

        {/* Model Distribution Bar Chart */}
        <Card>
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">模型费用分布</h2>
              <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">按费用排序的模型分布</p>
            </div>
          </div>
          {breakdown.length > 0 ? (
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={breakdown.slice(0, 6)} layout="vertical">
                <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" horizontal={false} />
                <XAxis type="number" tick={{ fontSize: 11, fill: '#94a3b8' }} tickFormatter={(v) => `$${v}`} />
                <YAxis type="category" dataKey="model" tick={{ fontSize: 10, fill: '#64748b' }} width={120} />
                <Tooltip formatter={(v) => [`$${Number(v).toFixed(2)}`, '费用']} />
                <Bar dataKey="total_cost_usd" fill="#6366f1" radius={[0, 4, 4, 0]} barSize={16} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex items-center justify-center h-[200px] text-sm text-slate-400">暂无数据</div>
          )}
        </Card>
      </div>

      {/* Row 3: Recent Usage */}
      <Card>
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">最近用量</h2>
            <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">最新的 API 调用记录</p>
          </div>
        </div>
        {recentUsage.length > 0 ? (
          <div className="space-y-0">
            {recentUsage.map((u) => (
              <div
                key={u.id}
                className="flex items-center justify-between py-2.5 border-b border-slate-100 dark:border-slate-800 last:border-0"
              >
                <div className="flex items-center gap-2.5 min-w-0">
                  <div className="w-2 h-2 rounded-full bg-indigo-500 shrink-0" />
                  <span className="text-sm text-slate-700 dark:text-slate-300 truncate max-w-[200px]">{u.model}</span>
                </div>
                <div className="flex items-center gap-4 shrink-0 ml-4">
                  <span className="text-sm font-medium text-slate-800 dark:text-slate-200">{formatUSD(u.cost_usd)}</span>
                  <span className="text-xs text-slate-400">{formatNum(u.input_tokens + u.output_tokens)} tokens</span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-8 text-sm text-slate-400">暂无用量记录</div>
        )}
      </Card>
    </div>
  )
}

function LoadingSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <div className="h-8 w-48 skeleton-wave rounded-lg" />
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <Skeleton key={i} className="h-28" />
        ))}
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Skeleton className="h-[260px]" />
        <Skeleton className="h-[260px]" />
      </div>
      <Skeleton className="h-48" />
    </div>
  )
}
