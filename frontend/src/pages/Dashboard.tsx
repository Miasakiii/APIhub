import { useEffect, useState, useCallback, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { DollarSign, BarChart3, Activity, Layers, ArrowUpRight, Cpu } from 'lucide-react'
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

interface BreakdownItem {
  model: string
  total_cost_usd: number
  total_tokens: number
  request_count: number
}

export function Dashboard() {
  const navigate = useNavigate()
  const [summary, setSummary] = useState<Summary | null>(null)
  const [breakdown, setBreakdown] = useState<BreakdownItem[]>([])
  const [loading, setLoading] = useState(true)
  const refreshTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  const fetchData = useCallback(() => {
    Promise.all([
      api.usage.summary().catch(() => null),
      api.stats.modelBreakdown().catch(() => []),
    ]).then(([s, b]) => {
      setSummary(s)
      setBreakdown(b)
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

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard icon={DollarSign} label="今日费用" value={formatUSD(totalCost)} accent="emerald" />
        <MetricCard icon={BarChart3} label="总 Tokens" value={formatNum(totalTokens)} accent="indigo" />
        <MetricCard icon={Activity} label="总请求数" value={formatNum(totalRequests)} accent="violet" />
        <MetricCard icon={Cpu} label="模型数" value={String(summary?.unique_models ?? 0)} accent="amber" />
      </div>

      <Card>
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">模型费用排行</h2>
            <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">点击模型查看详情</p>
          </div>
          <Layers className="w-4 h-4 text-slate-400" />
        </div>
        <div className="space-y-1">
          {breakdown.slice(0, 5).map((m, idx) => (
            <button
              type="button"
              key={m.model}
              onClick={() => navigate(`/model/${encodeURIComponent(m.model)}`)}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-800/50 transition text-left group"
            >
              <div className="w-6 h-6 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 text-white text-xs flex items-center justify-center font-bold shrink-0">
                {idx + 1}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-slate-700 dark:text-slate-300 truncate">{m.model}</p>
                <div className="w-full h-1.5 bg-slate-100 dark:bg-slate-800 rounded-full mt-1 overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-blue-500 to-indigo-500 rounded-full transition-all"
                    style={{ width: `${breakdown[0].total_cost_usd > 0 ? (m.total_cost_usd / breakdown[0].total_cost_usd) * 100 : 0}%` }}
                  />
                </div>
              </div>
              <div className="text-right shrink-0">
                <p className="text-sm font-semibold text-slate-800 dark:text-slate-200">{formatUSD(m.total_cost_usd)}</p>
                <p className="text-xs text-slate-400 dark:text-slate-500">{formatNum(m.request_count)} 次</p>
              </div>
              <ArrowUpRight className="w-4 h-4 text-slate-300 dark:text-slate-600 group-hover:text-blue-500 transition shrink-0" />
            </button>
          ))}
        </div>
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
      <Skeleton className="h-40" />
    </div>
  )
}
