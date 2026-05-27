import { useEffect, useState } from 'react'
import { Plus, Bell, Trash2, AlertTriangle, AlertCircle, Info, History, ToggleLeft, ToggleRight } from 'lucide-react'
import { api, type Alert, type AlertHistory } from '../api'
import { cn } from '../lib/utils'
import { PageHeader, Button, Modal, Input, Select, Skeleton } from '../components/ui'

export function Alerts() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [history, setHistory] = useState<AlertHistory[]>([])
  const [showForm, setShowForm] = useState(false)
  const [showHistory, setShowHistory] = useState(false)
  const [form, setForm] = useState({ name: '', type: 'balance_low', threshold: 10, unit: 'usd' })
  const [loading, setLoading] = useState(true)

  useEffect(() => { load() }, [])

  async function load() {
    setLoading(true)
    try { setAlerts(await api.alerts.list()) } catch (e) { console.error(e) } finally { setLoading(false) }
  }

  async function loadHistory() {
    try { setHistory(await api.alerts.history()); setShowHistory(true) } catch (e) { console.error(e) }
  }

  async function handleCreate() {
    if (!form.name) return
    await api.alerts.create(form)
    setForm({ name: '', type: 'balance_low', threshold: 10, unit: 'usd' })
    setShowForm(false)
    load()
  }

  async function handleDelete(id: string) { await api.alerts.delete(id); load() }
  async function handleToggle(alert: Alert) { await api.alerts.update(alert.id, { ...alert, enabled: !alert.enabled }); load() }

  const alertTypes: Record<string, string> = { balance_low: '余额不足', key_expired: 'Key 过期', abnormal_frequency: '异常频率', subscription_expiring: '订阅到期' }

  if (loading) return <LoadingSkeleton />

  return (
    <div className="space-y-6">
      <PageHeader
        title="告警规则"
        description={`共 ${alerts.length} 个告警规则`}
        actions={
          <>
            <Button variant="secondary" onClick={loadHistory}><History className="w-4 h-4" /> 历史记录</Button>
            <Button onClick={() => setShowForm(true)}><Plus className="w-4 h-4" /> 添加告警</Button>
          </>
        }
      />

      <Modal open={showForm} onClose={() => setShowForm(false)} title="新建告警规则">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <Input placeholder="规则名称" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} />
          <Select value={form.type} onChange={e => setForm({ ...form, type: e.target.value })}>
            {Object.entries(alertTypes).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
          </Select>
          <Input type="number" placeholder="阈值" value={form.threshold} onChange={e => setForm({ ...form, threshold: Number(e.target.value) })} />
          <Select value={form.unit} onChange={e => setForm({ ...form, unit: e.target.value })}>
            <option value="usd">USD</option><option value="percent">%</option><option value="days">天</option><option value="count">次</option>
          </Select>
        </div>
        <div className="flex gap-2 mt-4">
          <Button onClick={handleCreate}>确认添加</Button>
          <Button variant="ghost" onClick={() => setShowForm(false)}>取消</Button>
        </div>
      </Modal>

      <Modal open={showHistory} onClose={() => setShowHistory(false)} title="告警历史" maxWidth="max-w-xl">
        <div className="space-y-2 max-h-96 overflow-auto">
          {history.length === 0 ? (
            <p className="text-sm text-slate-400 text-center py-8">暂无告警记录</p>
          ) : history.map(h => (
            <div key={h.id} className={cn('flex items-start gap-3 p-3 rounded-xl text-sm',
              h.level === 'critical' ? 'bg-red-50 dark:bg-red-950/30 text-red-700 dark:text-red-300 border border-red-100 dark:border-red-900/50' :
              h.level === 'warning' ? 'bg-amber-50 dark:bg-amber-950/30 text-amber-700 dark:text-amber-300 border border-amber-100 dark:border-amber-900/50' :
              'bg-blue-50 dark:bg-blue-950/30 text-blue-700 dark:text-blue-300 border border-blue-100 dark:border-blue-900/50'
            )}>
              {h.level === 'critical' ? <AlertCircle className="w-4 h-4 mt-0.5 shrink-0" /> :
               h.level === 'warning' ? <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" /> :
               <Info className="w-4 h-4 mt-0.5 shrink-0" />}
              <div>
                <p className="font-medium">{h.message}</p>
                <p className="text-xs opacity-70">{h.created_at}</p>
              </div>
            </div>
          ))}
        </div>
      </Modal>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {alerts.map(a => (
          <div key={a.id} className={cn('bg-white dark:bg-slate-900/80 rounded-2xl border p-5 shadow-sm transition-all duration-300 overflow-hidden relative',
            a.enabled ? 'border-slate-200 dark:border-slate-700/60' : 'border-slate-100 dark:border-slate-800 opacity-60'
          )}>
            <div className={cn('absolute top-0 left-0 right-0 h-0.5', a.enabled ? 'bg-gradient-to-r from-blue-500 to-indigo-500' : 'bg-slate-200 dark:bg-slate-700')} />
            <div className="flex items-start justify-between mb-3">
              <div className="flex items-center gap-3">
                <div className={cn('w-10 h-10 rounded-xl flex items-center justify-center shrink-0',
                  a.enabled ? 'bg-blue-50 dark:bg-blue-950/40 text-blue-600 dark:text-blue-400' : 'bg-slate-100 dark:bg-slate-800 text-slate-400'
                )}>
                  <Bell className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-semibold text-slate-800 dark:text-slate-200 text-sm">{a.name}</h3>
                  <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">{alertTypes[a.type] || a.type}</p>
                </div>
              </div>
              <button type="button" onClick={() => handleToggle(a)} className="text-slate-400 hover:text-blue-600 transition">
                {a.enabled ? <ToggleRight className="w-6 h-6 text-blue-500" /> : <ToggleLeft className="w-6 h-6" />}
              </button>
            </div>

            <div className="flex items-center gap-2 mb-3">
              <span className="px-2 py-0.5 bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 text-xs rounded-md">阈值: {a.threshold} {a.unit}</span>
              {a.last_triggered_at && <span className="text-xs text-amber-600 dark:text-amber-400">上次: {new Date(a.last_triggered_at).toLocaleDateString()}</span>}
            </div>

            <div className="flex gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
              <span className={cn('flex-1 text-center text-xs py-1.5 rounded-lg',
                a.enabled ? 'bg-blue-50 dark:bg-blue-950/30 text-blue-600 dark:text-blue-400' : 'bg-slate-100 dark:bg-slate-800 text-slate-400'
              )}>{a.enabled ? '启用中' : '已禁用'}</span>
              <button type="button" title="删除" onClick={() => handleDelete(a.id)} className="p-1.5 text-slate-400 hover:text-red-500 transition"><Trash2 className="w-4 h-4" /></button>
            </div>
          </div>
        ))}
        {alerts.length === 0 && (
          <div className="col-span-full text-center py-16 bg-white dark:bg-slate-900/80 rounded-2xl border border-slate-200 dark:border-slate-700/60 border-dashed">
            <Bell className="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
            <p className="text-sm text-slate-400">暂无告警规则</p>
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
