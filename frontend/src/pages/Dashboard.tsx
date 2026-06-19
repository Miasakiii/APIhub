import { useEffect, useState, useCallback, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  DollarSign, BarChart3, Activity, Bell, Zap, Download, RefreshCw,
  Layers, ArrowUpRight, Cpu
} from 'lucide-react'
import {
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  BarChart, Bar, AreaChart, Area
} from 'recharts'
import { api } from '../api'
import { formatUSD, formatNum } from '../lib/utils'
import { useWSMessage } from '../lib/use-ws'
import { Button, Card, CardHeader, PageHeader, StatCard as MetricCard, Skeleton } from '../components/ui'

interface Summary {
  total_cost_usd: number
  total_tokens: number
  total_requests: number
  unique_models: number
  unique_keys: number
}

interface TrendItem {
  date: string
  cost_usd: number
  tokens: number
  request_count: number
}

interface BreakdownItem {
  model: string
  total_cost_usd: number
  total_tokens: number
  request_count: number
}

interface Alert {
  id: string
  name: string
  type: string
  enabled: boolean
}

export function Dashboard() {
  const navigate = useNavigate()
  const [summary, setSummary] = useState<Summary | null>(null)
  const [trend, setTrend] = useState<TrendItem[]>([])
  const [breakdown, setBreakdown] = useState<BreakdownItem[]>([])
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const refreshTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  const fetchData = useCallback(() => {
    Promise.all([
      api.usage.summary().catch(() => null),
      api.stats.costTrend().catch(() => []),
      api.stats.modelBreakdown().catch(() => []),
      api.alerts.list().catch(() => []),
    ]).then(([s, t, b, a]) => {
      setSummary(s)
      setTrend(t.reverse())
      setBreakdown(b)
      setAlerts(a.filter((x: Alert) => x.enabled).slice(0, 3))
    }).catch((e) => console.error(e)).finally(() => setLoading(false))
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  // Real-time: refresh data when usage updates arrive (debounced 500ms)
  useWSMessage('usage.update', () => {
    clearTimeout(refreshTimer.current)
    refreshTimer.current = setTimeout(fetchData, 500)
  })

  if (loading) return <LoadingSkeleton />

  const totalCost = summary?.total_cost_usd ?? 0
  const totalTokens = summary?.total_tokens ?? 0
  const totalRequests = summary?.total_requests ?? 0

  const recentTrend = trend.slice(-7)
  const prevTrend = trend.slice(-14, -7)
  const recentCost = recentTrend.reduce((s, i) => s + i.cost_usd, 0)
  const prevCost = prevTrend.reduce((s, i) => s + i.cost_usd, 0)
  const costChange = prevCost > 0 ? ((recentCost - prevCost) / prevCost) * 100 : 0

  return (
    <div className="space-y-6">
      <PageHeader
        title="数据总览"
        description="实时监控 API 用量、费用与模型分布"
        actions={
          <>
            <Button variant="secondary" onClick={() => api.export.csv().catch(console.error)}>
              <Download className="w-4 h-4" /> 导出
            </Button>
            <Button variant="secondary" onClick={() => { setLoading(true); fetchData() }}>
              <RefreshCw className="w-4 h-4" /> 刷新
            </Button>
          </>
        }
      />

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard icon={DollarSign} label="总费用" value={formatUSD(totalCost)} change={costChange} accent="emerald" />
        <MetricCard icon={BarChart3} label="总 Tokens" value={formatNum(totalTokens)} accent="indigo" />
        <MetricCard icon={Activity} label="总请求数" value={formatNum(totalRequests)} accent="violet" />
        <MetricCard icon={Cpu} label="模型数" value={String(summary?.unique_models ?? 0)} accent="amber" />
      </div>

      <Card>
        <CardHeader title="费用趋势" description="近 30 天费用变化" />
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={trend}>
            <defs>
              <linearGradient id="costGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#6366f1" stopOpacity={0.2} />
                <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
            <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#94a3b8' }} tickFormatter={(d: string) => d.slice(5)} axisLine={false} tickLine={false} />
            <YAxis tick={{ fontSize: 11, fill: '#94a3b8' }} tickFormatter={(v: number) => `$${v.toFixed(0)}`} axisLine={false} tickLine={false} />
            <Tooltip
              contentStyle={{ borderRadius: '12px', border: '1px solid #e2e8f0', boxShadow: '0 4px 12px rgba(0,0,0,0.08)', padding: '12px', backgroundColor: 'rgba(255,255,255,0.95)', backdropFilter: 'blur(8px)' }}
              formatter={(v) => [`$${Number(v).toFixed(2)}`, '费用']}
              labelFormatter={(l) => String(l)}
            />
            <Area type="monotone" dataKey="cost_usd" stroke="#6366f1" strokeWidth={2} fill="url(#costGradient)" dot={false} activeDot={{ r: 4, fill: '#6366f1', strokeWidth: 2, stroke: '#fff' }} />
          </AreaChart>
        </ResponsiveContainer>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">模型费用排行</h2>
              <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">点击模型查看详情</p>
            </div>
            <Layers className="w-4 h-4 text-slate-400" />
          </div>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={breakdown.slice(0, 8).reverse()} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" horizontal={false} />
              <XAxis type="number" tick={{ fontSize: 11, fill: '#94a3b8' }} tickFormatter={(v: number) => `$${v.toFixed(0)}`} axisLine={false} tickLine={false} />
              <YAxis type="category" dataKey="model" tick={{ fontSize: 11, fill: '#64748b' }} width={100} axisLine={false} tickLine={false} />
              <Tooltip
                contentStyle={{ borderRadius: '12px', border: '1px solid #e2e8f0', boxShadow: '0 4px 12px rgba(0,0,0,0.08)', padding: '12px', backgroundColor: 'rgba(255,255,255,0.95)', backdropFilter: 'blur(8px)' }}
                formatter={(v) => [`$${Number(v).toFixed(2)}`, '费用']}
              />
              <Bar dataKey="total_cost_usd" fill="#6366f1" radius={[0, 6, 6, 0]} barSize={20} cursor="pointer" />
            </BarChart>
          </ResponsiveContainer>

          <div className="mt-4 space-y-1">
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

        <Card>
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">活跃告警</h2>
              <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">最近触发的告警规则</p>
            </div>
            <Bell className="w-4 h-4 text-slate-400" />
          </div>
          {alerts.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="w-12 h-12 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center mb-3">
                <Bell className="w-5 h-5 text-slate-400" />
              </div>
              <p className="text-sm text-slate-400">暂无活跃告警</p>
              <p className="text-xs text-slate-300 dark:text-slate-600 mt-1">所有系统运行正常</p>
            </div>
          ) : (
            <div className="space-y-3">
              {alerts.map(a => (
                <div key={a.id} className="flex items-center gap-3 p-3 bg-amber-50 dark:bg-amber-950/30 rounded-xl border border-amber-100 dark:border-amber-900/50">
                  <Zap className="w-4 h-4 text-amber-600 dark:text-amber-400 shrink-0" />
                  <div>
                    <p className="text-sm font-medium text-slate-800 dark:text-slate-200">{a.name}</p>
                    <p className="text-xs text-slate-500 dark:text-slate-400">{a.type}</p>
                  </div>
                </div>
              ))}
            </div>
          )}

          <div className="mt-6">
            <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-3">用量详情</h3>
            <div className="border border-slate-200 dark:border-slate-700 rounded-xl overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/50">
                    <th className="text-left px-4 py-2.5 text-xs font-medium text-slate-500 dark:text-slate-400">模型</th>
                    <th className="text-right px-4 py-2.5 text-xs font-medium text-slate-500 dark:text-slate-400">费用</th>
                    <th className="text-right px-4 py-2.5 text-xs font-medium text-slate-500 dark:text-slate-400">Tokens</th>
                    <th className="text-right px-4 py-2.5 text-xs font-medium text-slate-500 dark:text-slate-400">请求</th>
                  </tr>
                </thead>
                <tbody>
                  {breakdown.slice(0, 5).map((m) => (
                    <tr key={m.model} className="border-b border-slate-50 dark:border-slate-800/50 last:border-0 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition">
                      <td className="px-4 py-2.5 font-medium text-slate-700 dark:text-slate-300 text-xs truncate max-w-[120px]">{m.model}</td>
                      <td className="px-4 py-2.5 text-right tabular-nums text-emerald-700 dark:text-emerald-400 text-xs font-medium">{formatUSD(m.total_cost_usd)}</td>
                      <td className="px-4 py-2.5 text-right tabular-nums text-slate-500 dark:text-slate-400 text-xs">{formatNum(m.total_tokens)}</td>
                      <td className="px-4 py-2.5 text-right tabular-nums text-slate-500 dark:text-slate-400 text-xs">{formatNum(m.request_count)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </Card>
      </div>

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
      <Skeleton className="h-80" />
    </div>
  )
}
