import { useEffect, useState } from 'react'
import { api, type UsageSession, type SessionStats, type ActivityBucket } from '../api'
import { Card, StatCard, Button } from '../components/ui'
import { Clock, Zap, DollarSign, BarChart3 } from 'lucide-react'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts'
import { formatUSD, formatNum } from '../lib/utils'

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`
  const min = Math.floor(ms / 60_000)
  const sec = Math.floor((ms % 60_000) / 1000)
  return `${min}m ${sec}s`
}

export function Sessions() {
  const [sessions, setSessions] = useState<UsageSession[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [stats, setStats] = useState<SessionStats | null>(null)
  const [hourlyBuckets, setHourlyBuckets] = useState<ActivityBucket[]>([])
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10))
  const [loading, setLoading] = useState(true)
  const [model, setModel] = useState('')
  const [providerID, setProviderID] = useState('')
  const pageSize = 20

  useEffect(() => {
    let cancelled = false

    const params: Record<string, string | number> = { page, page_size: pageSize }
    if (model) params.model = model
    if (providerID) params.provider_id = providerID

    Promise.all([
      api.sessions.list(params).catch(() => ({ sessions: [], total: 0, page: 1, page_size: pageSize })),
      api.sessions.stats().catch(() => ({ total_sessions: 0, avg_duration_ms: 0, avg_cost_usd: 0, total_cost_usd: 0, total_requests: 0 })),
      api.sessions.hourly(date).catch(() => ({ date, buckets: [] })),
    ]).then(([sessRes, statsRes, hourlyRes]) => {
      if (cancelled) return
      setSessions(sessRes.sessions ?? [])
      setTotal(sessRes.total)
      setStats(statsRes)
      setHourlyBuckets(hourlyRes.buckets ?? [])
    }).catch(() => {}).finally(() => {
      if (!cancelled) setLoading(false)
    })

    return () => { cancelled = true }
  }, [page, model, providerID, date])

  // Prepare hourly chart data: merge all models per hour
  const hourlyData = Array.from({ length: 24 }, (_, h) => {
    const hourBuckets = hourlyBuckets.filter(b => new Date(b.bucket_start).getHours() === h)
    return {
      hour: `${h}:00`,
      requests: hourBuckets.reduce((sum, b) => sum + b.request_count, 0),
      cost: hourBuckets.reduce((sum, b) => sum + b.cost_usd, 0),
    }
  })

  const totalPages = Math.ceil(total / pageSize)

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">会话分析</h1>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard icon={Zap} label="总会话数" value={formatNum(stats?.total_sessions ?? 0)} />
        <StatCard icon={Clock} label="平均时长" value={formatDuration(stats?.avg_duration_ms ?? 0)} />
        <StatCard icon={DollarSign} label="平均成本" value={formatUSD(stats?.avg_cost_usd ?? 0)} />
        <StatCard icon={BarChart3} label="总会话成本" value={formatUSD(stats?.total_cost_usd ?? 0)} />
      </div>

      {/* Hourly Activity Chart */}
      <Card>
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-semibold">小时活跃分布</h2>
          <input
            type="date"
            value={date}
            onChange={e => { setDate(e.target.value); setPage(1) }}
            className="px-3 py-1.5 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800"
          />
        </div>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={hourlyData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" opacity={0.3} />
              <XAxis dataKey="hour" tick={{ fontSize: 11 }} interval={2} />
              <YAxis tick={{ fontSize: 11 }} />
              <Tooltip
                formatter={(value, name) => [
                  name === 'requests' ? formatNum(Number(value)) : formatUSD(Number(value)),
                  name === 'requests' ? '请求数' : '成本',
                ]}
                contentStyle={{ borderRadius: 8, border: '1px solid #334155' }}
              />
              <Bar dataKey="requests" fill="#6366f1" radius={[4, 4, 0, 0]} name="requests" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </Card>

      {/* Filters */}
      <Card>
        <div className="flex flex-wrap items-center gap-3 mb-4">
          <h2 className="font-semibold">会话列表</h2>
          <input
            placeholder="模型过滤..."
            value={model}
            onChange={e => { setModel(e.target.value); setPage(1) }}
            className="px-3 py-1.5 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 w-48"
          />
          <input
            placeholder="Provider 过滤..."
            value={providerID}
            onChange={e => { setProviderID(e.target.value); setPage(1) }}
            className="px-3 py-1.5 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 w-48"
          />
          <span className="text-sm text-slate-500 ml-auto">共 {total} 条</span>
        </div>

        {loading ? (
          <div className="flex items-center justify-center h-32">
            <div className="w-6 h-6 rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600 animate-pulse" />
          </div>
        ) : sessions.length === 0 ? (
          <div className="text-center py-12 text-slate-500">暂无会话数据</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-200 dark:border-slate-700 text-left text-slate-500">
                  <th className="pb-2 pr-4">模型</th>
                  <th className="pb-2 pr-4">Provider</th>
                  <th className="pb-2 pr-4">来源</th>
                  <th className="pb-2 pr-4">开始时间</th>
                  <th className="pb-2 pr-4 text-right">时长</th>
                  <th className="pb-2 pr-4 text-right">请求数</th>
                  <th className="pb-2 pr-4 text-right">Input</th>
                  <th className="pb-2 pr-4 text-right">Output</th>
                  <th className="pb-2 text-right">成本</th>
                </tr>
              </thead>
              <tbody>
                {sessions.map(s => (
                  <tr key={s.id} className="border-b border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50">
                    <td className="py-2.5 pr-4 font-mono text-xs">{s.model}</td>
                    <td className="py-2.5 pr-4 text-xs">{s.provider_id.slice(0, 8)}</td>
                    <td className="py-2.5 pr-4">
                      <span className="px-2 py-0.5 text-xs rounded-full bg-slate-100 dark:bg-slate-700">{s.source}</span>
                    </td>
                    <td className="py-2.5 pr-4 text-xs">{new Date(s.started_at).toLocaleString()}</td>
                    <td className="py-2.5 pr-4 text-right text-xs">{formatDuration(s.duration_ms)}</td>
                    <td className="py-2.5 pr-4 text-right">{s.request_count}</td>
                    <td className="py-2.5 pr-4 text-right text-xs">{formatNum(s.input_tokens)}</td>
                    <td className="py-2.5 pr-4 text-right text-xs">{formatNum(s.output_tokens)}</td>
                    <td className="py-2.5 text-right font-medium">{formatUSD(s.cost_usd)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between mt-4">
            <span className="text-sm text-slate-500">
              第 {page} / {totalPages} 页
            </span>
            <div className="flex gap-2">
              <Button variant="secondary" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
                上一页
              </Button>
              <Button variant="secondary" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
                下一页
              </Button>
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}
