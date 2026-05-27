import { useEffect, useState } from 'react'
import { ArrowLeft, DollarSign, BarChart3, Activity, Calendar, Layers, Clock } from 'lucide-react'
import {
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  AreaChart, Area, BarChart, Bar, Legend
} from 'recharts'
import { api } from '../api'
import type { DailyStats, UsageRecord } from '../api'
import { formatUSD, formatNum } from '../lib/utils'
import { PageHeader, Button, Card, CardHeader, StatCard, Badge, Skeleton } from '../components/ui'

interface Props {
  model: string
  onBack: () => void
}

export function ModelDetail({ model, onBack }: Props) {
  const [daily, setDaily] = useState<DailyStats[]>([])
  const [records, setRecords] = useState<UsageRecord[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(30)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    Promise.all([
      api.stats.daily({ model }).catch(() => []),
      api.usage.list({ model, page: 1, page_size: pageSize }).catch(() => ({ records: [] as UsageRecord[], total: 0, page: 1, page_size: pageSize })),
    ]).then(([d, u]) => {
      if (cancelled) return
      setDaily(d.reverse())
      setRecords(u.records)
      setTotal(u.total)
    }).finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [model, pageSize])

  function loadPage(p: number) {
    setPage(p)
    api.usage.list({ model, page: p, page_size: pageSize })
      .then((data) => { setRecords(data.records); setTotal(data.total) })
      .catch(console.error)
  }

  if (loading) return <LoadingSkeleton />

  const totalCost = daily.reduce((s, d) => s + d.cost_usd, 0)
  const totalTokens = daily.reduce((s, d) => s + d.input_tokens + d.output_tokens + d.cache_read + d.cache_create, 0)
  const totalReqs = daily.reduce((s, d) => s + d.request_count, 0)
  const dayCount = daily.length || 1
  const avgCost = totalCost / dayCount

  const trendData = daily.map(d => ({
    date: d.date,
    cost_usd: d.cost_usd,
    input_tokens: d.input_tokens,
    output_tokens: d.output_tokens,
  }))

  const totalPages = Math.ceil(total / pageSize)

  return (
    <div className="space-y-6">
      <PageHeader
        title={model}
        description="模型用量详情"
        actions={
          <Button variant="secondary" onClick={onBack}>
            <ArrowLeft className="w-4 h-4" /> 返回总览
          </Button>
        }
      />

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard icon={DollarSign} label="总费用" value={formatUSD(totalCost)} accent="emerald" />
        <StatCard icon={BarChart3} label="总 Tokens" value={formatNum(totalTokens)} accent="indigo" />
        <StatCard icon={Activity} label="总请求数" value={formatNum(totalReqs)} accent="violet" />
        <StatCard icon={Calendar} label="日均费用" value={formatUSD(avgCost)} accent="amber" />
      </div>

      <Card>
        <CardHeader title="费用趋势" description={`${daily.length} 天费用变化`} />
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={trendData}>
            <defs>
              <linearGradient id="modelCostGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#6366f1" stopOpacity={0.2} />
                <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
            <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#94a3b8' }} tickFormatter={(d: string) => d.slice(5)} axisLine={false} tickLine={false} />
            <YAxis tick={{ fontSize: 11, fill: '#94a3b8' }} tickFormatter={(v: number) => `$${v.toFixed(2)}`} axisLine={false} tickLine={false} />
            <Tooltip
              contentStyle={{ borderRadius: '12px', border: '1px solid #e2e8f0', boxShadow: '0 4px 12px rgba(0,0,0,0.08)', padding: '12px', backgroundColor: 'rgba(255,255,255,0.95)', backdropFilter: 'blur(8px)' }}
              formatter={(v, name) => {
                if (name === 'cost_usd') return [`$${Number(v).toFixed(4)}`, '费用']
                return [formatNum(Number(v)), name === 'input_tokens' ? '输入 Token' : '输出 Token']
              }}
              labelFormatter={(l) => String(l)}
            />
            <Area type="monotone" dataKey="cost_usd" stroke="#6366f1" strokeWidth={2} fill="url(#modelCostGrad)" dot={false} activeDot={{ r: 4, fill: '#6366f1', strokeWidth: 2, stroke: '#fff' }} />
          </AreaChart>
        </ResponsiveContainer>
      </Card>

      <Card>
        <CardHeader title="Token 用量分布" description="每日输入/输出 Token" />
        <ResponsiveContainer width="100%" height={280}>
          <BarChart data={trendData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
            <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#94a3b8' }} tickFormatter={(d: string) => d.slice(5)} axisLine={false} tickLine={false} />
            <YAxis tick={{ fontSize: 11, fill: '#94a3b8' }} tickFormatter={(v: number) => formatNum(v)} axisLine={false} tickLine={false} />
            <Tooltip
              contentStyle={{ borderRadius: '12px', border: '1px solid #e2e8f0', boxShadow: '0 4px 12px rgba(0,0,0,0.08)', padding: '12px', backgroundColor: 'rgba(255,255,255,0.95)', backdropFilter: 'blur(8px)' }}
              formatter={(v, name) => [formatNum(Number(v)), name === 'input_tokens' ? '输入' : '输出']}
            />
            <Legend formatter={(value: string) => value === 'input_tokens' ? '输入' : '输出'} />
            <Bar dataKey="input_tokens" stackId="tokens" fill="#6366f1" radius={[0, 0, 0, 0]} barSize={20} />
            <Bar dataKey="output_tokens" stackId="tokens" fill="#a78bfa" radius={[4, 4, 0, 0]} barSize={20} />
          </BarChart>
        </ResponsiveContainer>
      </Card>

      <Card>
        <CardHeader title="用量记录" description={`共 ${total} 条`} />
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/50">
                <th className="text-left px-4 py-3.5 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">时间</th>
                <th className="text-right px-4 py-3.5 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">输入</th>
                <th className="text-right px-4 py-3.5 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">输出</th>
                <th className="text-right px-4 py-3.5 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Cache 读</th>
                <th className="text-right px-4 py-3.5 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Cache 写</th>
                <th className="text-right px-4 py-3.5 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">费用</th>
                <th className="text-left px-4 py-3.5 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">来源</th>
              </tr>
            </thead>
            <tbody>
              {records.map((r, idx) => (
                <tr key={r.id} className={`border-b border-slate-50 dark:border-slate-800/50 last:border-0 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition ${idx % 2 === 0 ? '' : 'bg-slate-50/30 dark:bg-slate-800/10'}`}>
                  <td className="px-4 py-3 text-slate-500 dark:text-slate-400 whitespace-nowrap">
                    <div className="flex items-center gap-1.5">
                      <Clock className="w-3.5 h-3.5 text-slate-300 dark:text-slate-600" />
                      {new Date(r.timestamp).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-right tabular-nums text-slate-600 dark:text-slate-400">{formatNum(r.input_tokens)}</td>
                  <td className="px-4 py-3 text-right tabular-nums text-slate-600 dark:text-slate-400">{formatNum(r.output_tokens)}</td>
                  <td className="px-4 py-3 text-right tabular-nums text-blue-600 dark:text-blue-400">{formatNum(r.cache_read)}</td>
                  <td className="px-4 py-3 text-right tabular-nums text-violet-600 dark:text-violet-400">{formatNum(r.cache_create)}</td>
                  <td className="px-4 py-3 text-right tabular-nums font-semibold text-emerald-700 dark:text-emerald-400">{formatUSD(r.cost_usd)}</td>
                  <td className="px-4 py-3">
                    <Badge variant={r.source === 'syncer' ? 'warning' : r.source === 'ccswitch' ? 'info' : 'default'}>{r.source}</Badge>
                  </td>
                </tr>
              ))}
              {records.length === 0 && (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center">
                    <Layers className="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
                    <p className="text-sm text-slate-400">暂无用量记录</p>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {totalPages > 1 && (
          <div className="flex items-center justify-between mt-4 pt-4 border-t border-slate-100 dark:border-slate-800">
            <p className="text-sm text-slate-500 dark:text-slate-400">
              共 {total} 条，第 {page}/{totalPages} 页
            </p>
            <div className="flex gap-2">
              <Button variant="secondary" size="sm" disabled={page <= 1} onClick={() => loadPage(page - 1)}>上一页</Button>
              <Button variant="secondary" size="sm" disabled={page >= totalPages} onClick={() => loadPage(page + 1)}>下一页</Button>
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}

function LoadingSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <Skeleton className="h-8 w-48" />
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map(i => <Skeleton key={i} className="h-28" />)}
      </div>
      <Skeleton className="h-80" />
      <Skeleton className="h-72" />
    </div>
  )
}
