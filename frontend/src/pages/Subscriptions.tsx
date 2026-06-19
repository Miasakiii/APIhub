import { useEffect, useState } from 'react'
import { Plus, Calendar, Trash2, CreditCard, Package } from 'lucide-react'
import { api, type Subscription } from '../api'
import { cn, formatUSD } from '../lib/utils'
import { PageHeader, Button, Modal, Input, Select, Badge, Skeleton } from '../components/ui'

export function Subscriptions() {
  const [subs, setSubs] = useState<Subscription[]>([])
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState<Partial<Subscription>>({
    plan_name: '', price: 0, currency: 'USD', billing_cycle: 'monthly', status: 'active'
  })
  const [loading, setLoading] = useState(true)

  useEffect(() => { load() }, [])

  async function load() {
    setLoading(true)
    try { setSubs(await api.subscriptions.list()) } catch (e) { console.error(e) } finally { setLoading(false) }
  }

  async function handleCreate() {
    if (!form.plan_name) return
    await api.subscriptions.create(form)
    setForm({ plan_name: '', price: 0, currency: 'USD', billing_cycle: 'monthly', status: 'active' })
    setShowForm(false)
    load()
  }

  async function handleDelete(id: string) { await api.subscriptions.delete(id); load() }

  const statusMap: Record<string, { label: string; variant: 'success' | 'danger' | 'muted' | 'warning' }> = {
    active: { label: '活跃', variant: 'success' },
    expired: { label: '已过期', variant: 'danger' },
    cancelled: { label: '已取消', variant: 'muted' },
    paused: { label: '已暂停', variant: 'warning' },
  }
  const cycleMap: Record<string, string> = { monthly: '月付', yearly: '年付', 'one-time': '一次性', 'pay-as-go': '按量计费' }

  if (loading) return <LoadingSkeleton />

  return (
    <div className="space-y-6">
      <PageHeader title="订阅管理" description={`共 ${subs.length} 个订阅`} actions={<Button onClick={() => setShowForm(true)}><Plus className="w-4 h-4" /> 添加订阅</Button>} />

      <Modal open={showForm} onClose={() => setShowForm(false)} title="新建订阅">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <Input placeholder="计划名称" value={form.plan_name || ''} onChange={e => setForm({ ...form, plan_name: e.target.value })} />
          <Input type="number" placeholder="价格" value={form.price || 0} onChange={e => setForm({ ...form, price: Number(e.target.value) })} />
          <Select value={form.billing_cycle} onChange={e => setForm({ ...form, billing_cycle: e.target.value })}>
            {Object.entries(cycleMap).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
          </Select>
          <Select value={form.status} onChange={e => setForm({ ...form, status: e.target.value })}>
            <option value="active">活跃</option><option value="paused">暂停</option><option value="expired">已过期</option><option value="cancelled">已取消</option>
          </Select>
          <Input type="date" placeholder="开始日期" onChange={e => setForm({ ...form, start_date: e.target.value })} />
          <Input type="date" placeholder="续费日期" onChange={e => setForm({ ...form, renew_date: e.target.value })} />
        </div>
        <div className="flex gap-2 mt-4">
          <Button onClick={handleCreate}>确认添加</Button>
          <Button variant="ghost" onClick={() => setShowForm(false)}>取消</Button>
        </div>
      </Modal>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {subs.map(s => {
          const status = statusMap[s.status] || { label: s.status, variant: 'muted' as const }
          const daysLeft = s.renew_date ? Math.ceil((new Date(s.renew_date).getTime() - Date.now()) / (1000 * 60 * 60 * 24)) : null

          return (
            <div key={s.id} className="bg-white dark:bg-slate-900/80 rounded-2xl border border-slate-200 dark:border-slate-700/60 p-5 shadow-sm hover:shadow-lg dark:hover:shadow-xl transition-all duration-300 overflow-hidden relative">
              <div className="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-violet-500 to-purple-600" />
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 text-white flex items-center justify-center shrink-0">
                    <Package className="w-5 h-5" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-slate-800 dark:text-slate-200 text-sm">{s.plan_name}</h3>
                    {s.provider && <p className="text-xs text-slate-400 dark:text-slate-500">{s.provider.name}</p>}
                  </div>
                </div>
                <div className="flex items-center gap-1.5">
                  {s.source === 'auto' && <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-indigo-50 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400 font-medium border border-indigo-100 dark:border-indigo-800/50">自动检测</span>}
                  <Badge variant={status.variant}>{status.label}</Badge>
                </div>
              </div>

              <div className="flex items-end gap-1 mb-4">
                <span className="text-2xl font-bold text-slate-900 dark:text-slate-100">{formatUSD(s.price)}</span>
                <span className="text-sm text-slate-400 dark:text-slate-500 mb-1">/{cycleMap[s.billing_cycle] || s.billing_cycle}</span>
              </div>

              {s.quota_total > 0 && (
                <div className="mb-4">
                  <div className="flex justify-between text-xs text-slate-500 dark:text-slate-400 mb-1.5">
                    <span>用量</span>
                    <span className="font-medium">{((s.quota_used / s.quota_total) * 100).toFixed(0)}%</span>
                  </div>
                  <div className="h-2 bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden">
                    <div className="h-full bg-gradient-to-r from-violet-500 to-purple-600 rounded-full transition-all" style={{ width: `${Math.min((s.quota_used / s.quota_total) * 100, 100)}%` }} />
                  </div>
                  <p className="text-xs text-slate-400 dark:text-slate-500 mt-1.5">{s.quota_used.toLocaleString()} / {s.quota_total.toLocaleString()} {s.quota_type}</p>
                </div>
              )}

              {daysLeft !== null && (
                <div className="flex items-center gap-1.5 text-sm mb-4">
                  <Calendar className="w-3.5 h-3.5 text-slate-400" />
                  <span className={cn(daysLeft < 7 ? 'text-red-600 dark:text-red-400 font-medium' : 'text-slate-500 dark:text-slate-400')}>
                    {daysLeft > 0 ? `还剩 ${daysLeft} 天` : daysLeft === 0 ? '今天到期' : `已过期 ${-daysLeft} 天`}
                  </span>
                </div>
              )}

              <div className="flex gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
                <button type="button" onClick={() => handleDelete(s.id)} className="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs text-slate-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/30 rounded-lg transition">
                  <Trash2 className="w-3.5 h-3.5" /> 删除
                </button>
              </div>
            </div>
          )
        })}
        {subs.length === 0 && (
          <div className="col-span-full text-center py-16 bg-white dark:bg-slate-900/80 rounded-2xl border border-slate-200 dark:border-slate-700/60 border-dashed">
            <CreditCard className="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
            <p className="text-sm text-slate-400">暂无订阅</p>
          </div>
        )}
      </div>
    </div>
  )
}

function LoadingSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <Skeleton className="h-8 w-48" />
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3].map(i => <Skeleton key={i} className="h-40" />)}
      </div>
    </div>
  )
}
