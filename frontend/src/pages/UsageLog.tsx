import { useEffect, useState } from 'react'
import { Search, Filter, Clock, FileText, X } from 'lucide-react'
import { api } from '../api'
import { formatUSD, formatNum } from '../lib/utils'
import { PageHeader, Button, Card, Badge, Skeleton } from '../components/ui'

interface UsageRecord {
  id: string
  provider_id: string
  model: string
  input_tokens: number
  output_tokens: number
  cache_read: number
  cache_create: number
  cost_usd: number
  source: string
  timestamp: string
}

const sourceOptions = [
  { value: '', label: '全部来源' },
  { value: 'ccswitch', label: 'cc-switch' },
  { value: 'jsonl', label: 'JSONL' },
  { value: 'syncer', label: 'Syncer' },
  { value: 'manual', label: '手动' },
]

const sourceBadgeVariant: Record<string, 'default' | 'success' | 'warning' | 'info' | 'muted'> = {
  ccswitch: 'info',
  jsonl: 'default',
  syncer: 'warning',
  manual: 'muted',
}

export function UsageLog() {
  const [records, setRecords] = useState<UsageRecord[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(50)
  const [loading, setLoading] = useState(true)
  const [sourceFilter, setSourceFilter] = useState('')
  const [modelFilter, setModelFilter] = useState('')

  useEffect(() => {
    let cancelled = false
    const params: Record<string, string | number> = { page, page_size: pageSize }
    if (sourceFilter) params.source = sourceFilter
    if (modelFilter) params.model = modelFilter
    api.usage.list(params)
      .then((data) => {
        if (cancelled) return
        setRecords(data.records)
        setTotal(data.total)
      })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [sourceFilter, modelFilter, page, pageSize])

  const totalPages = Math.ceil(total / pageSize)

  function refresh() {
    setLoading(true)
    const params: Record<string, string | number> = { page, page_size: pageSize }
    if (sourceFilter) params.source = sourceFilter
    if (modelFilter) params.model = modelFilter
    api.usage.list(params)
      .then((data) => { setRecords(data.records); setTotal(data.total) })
      .finally(() => setLoading(false))
  }

  if (loading) return <LoadingSkeleton />

  return (
    <div className="space-y-6">
      <PageHeader
        title="用量日志"
        description={`共 ${total} 条记录`}
        actions={
          <Button variant="secondary" onClick={refresh}>
            <Clock className="w-4 h-4" /> 刷新
          </Button>
        }
      />

      <Card>
        <div className="flex items-center gap-3 flex-wrap">
          <div className="flex items-center gap-2 px-3 py-2 bg-slate-50 dark:bg-slate-800 rounded-xl border border-slate-100 dark:border-slate-700">
            <Filter className="w-4 h-4 text-slate-400" />
            <select
              className="bg-transparent text-sm text-slate-700 dark:text-slate-300 focus:outline-none min-w-[100px]"
              value={sourceFilter}
              onChange={(e) => { setSourceFilter(e.target.value); setPage(1) }}
              title="来源过滤"
            >
              {sourceOptions.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
            </select>
          </div>
          <div className="flex items-center gap-2 px-3 py-2 bg-slate-50 dark:bg-slate-800 rounded-xl border border-slate-100 dark:border-slate-700">
            <Search className="w-4 h-4 text-slate-400" />
            <input
              className="bg-transparent text-sm text-slate-700 dark:text-slate-300 focus:outline-none w-40"
              placeholder="过滤模型..."
              value={modelFilter}
              onChange={(e) => { setModelFilter(e.target.value); setPage(1) }}
            />
            {modelFilter && (
              <button type="button" title="清除过滤" onClick={() => setModelFilter('')} className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-300">
                <X className="w-3 h-3" />
              </button>
            )}
          </div>
        </div>
      </Card>

      <Card padding={false}>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/50">
                <th className="text-left px-4 py-3.5 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">时间</th>
                <th className="text-left px-4 py-3.5 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">模型</th>
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
                  <td className="px-4 py-3 font-mono text-xs text-slate-700 dark:text-slate-300 max-w-[200px] truncate">{r.model}</td>
                  <td className="px-4 py-3 text-right tabular-nums text-slate-600 dark:text-slate-400">{formatNum(r.input_tokens)}</td>
                  <td className="px-4 py-3 text-right tabular-nums text-slate-600 dark:text-slate-400">{formatNum(r.output_tokens)}</td>
                  <td className="px-4 py-3 text-right tabular-nums text-blue-600 dark:text-blue-400">{formatNum(r.cache_read)}</td>
                  <td className="px-4 py-3 text-right tabular-nums text-violet-600 dark:text-violet-400">{formatNum(r.cache_create)}</td>
                  <td className="px-4 py-3 text-right tabular-nums font-semibold text-emerald-700 dark:text-emerald-400">{formatUSD(r.cost_usd)}</td>
                  <td className="px-4 py-3">
                    <Badge variant={sourceBadgeVariant[r.source] || 'muted'}>{r.source}</Badge>
                  </td>
                </tr>
              ))}
              {records.length === 0 && (
                <tr>
                  <td colSpan={8} className="px-4 py-12 text-center">
                    <FileText className="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
                    <p className="text-sm text-slate-400">暂无用量记录</p>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {totalPages > 1 && (
        <Card>
          <div className="flex items-center justify-between">
            <p className="text-sm text-slate-500 dark:text-slate-400">
              共 {total} 条，第 {page}/{totalPages} 页
            </p>
            <div className="flex gap-2">
              <Button variant="secondary" size="sm" disabled={page <= 1} onClick={() => setPage(p => Math.max(1, p - 1))}>上一页</Button>
              <Button variant="secondary" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>下一页</Button>
            </div>
          </div>
        </Card>
      )}
    </div>
  )
}

function LoadingSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-16" />
      <Skeleton className="h-96" />
    </div>
  )
}
