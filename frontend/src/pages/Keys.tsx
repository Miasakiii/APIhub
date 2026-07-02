import { useEffect, useState } from 'react'
import { Plus, Eye, EyeOff, KeyRound, Copy, Check, History, Clock } from 'lucide-react'
import { api, type KeyAuditEntry } from '../api'
import { formatUSD, cn } from '../lib/utils'
import { PageHeader, Button, Modal, Input, Select, Badge, Skeleton } from '../components/ui'
import { useToast } from '../lib/use-toast'

interface KeyItem {
  id: string
  provider_id: string
  key_hash: string
  name: string
  source: string
  status: string
  balance_usd: number
  last_checked?: string
}

export function Keys() {
  const [keys, setKeys] = useState<KeyItem[]>([])
  const [providers, setProviders] = useState<Array<{ id: string; name: string }>>([])
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ provider_id: '', key: '', name: '' })
  const [loading, setLoading] = useState(true)
  const [decrypted, setDecrypted] = useState<Record<string, string>>({})
  const [copied, setCopied] = useState<string | null>(null)
  const [auditKey, setAuditKey] = useState<KeyItem | null>(null)
  const [auditLogs, setAuditLogs] = useState<KeyAuditEntry[]>([])
  const [auditLoading, setAuditLoading] = useState(false)
  const toast = useToast()

  useEffect(() => { load() }, [])

  async function load() {
    const [kList, pList] = await Promise.all([api.keys.list(), api.providers.list()])
    setKeys(kList)
    setProviders(pList)
    setLoading(false)
  }

  async function handleCreate() {
    if (!form.provider_id || !form.key) return
    await api.keys.create(form.provider_id, form.key, form.name)
    setForm({ provider_id: '', key: '', name: '' })
    setShowForm(false)
    toast.success('API Key 已添加')
    load()
  }

  async function handleDecrypt(id: string) {
    if (decrypted[id]) {
      const next = { ...decrypted }
      delete next[id]
      setDecrypted(next)
      return
    }
    const { key } = await api.keys.decrypt(id)
    setDecrypted((prev) => ({ ...prev, [id]: key }))
  }

  async function handleCopy(id: string, text: string) {
    await navigator.clipboard.writeText(text)
    setCopied(id)
    toast.success('已复制到剪贴板')
    setTimeout(() => setCopied(null), 2000)
  }

  async function handleAudit(k: KeyItem) {
    setAuditKey(k)
    setAuditLoading(true)
    try {
      const { audit_logs } = await api.keys.audit(k.id)
      setAuditLogs(audit_logs)
    } catch {
      setAuditLogs([])
    } finally {
      setAuditLoading(false)
    }
  }

  function closeAudit() {
    setAuditKey(null)
    setAuditLogs([])
  }

  const providerMap: Record<string, string> = {}
  providers.forEach((p) => { providerMap[p.id] = p.name })

  if (loading) return <LoadingSkeleton />

  return (
    <div className="space-y-6">
      <PageHeader
        title="API Keys"
        description="管理你的 API 密钥"
        actions={
          <Button onClick={() => setShowForm(true)}>
            <Plus className="w-4 h-4" /> 添加 Key
          </Button>
        }
      />

      <Modal open={!!auditKey} onClose={closeAudit} title={`审计日志 — ${auditKey?.name || '未命名'}`}>
        {auditLoading ? (
          <div className="flex items-center gap-2 text-sm text-slate-500 py-4">
            <div className="w-4 h-4 border-2 border-indigo-300 border-t-indigo-600 rounded-full animate-spin" />
            加载中...
          </div>
        ) : auditLogs.length === 0 ? (
          <p className="text-sm text-slate-400 py-4">暂无审计记录</p>
        ) : (
          <div className="space-y-1 max-h-80 overflow-y-auto">
            {auditLogs.map((entry) => (
              <div key={entry.id} className="flex items-start gap-3 p-2.5 rounded-lg bg-slate-50 dark:bg-slate-800/50">
                <div className={cn(
                  'w-8 h-8 rounded-lg flex items-center justify-center shrink-0',
                  actionStyle[entry.action]?.bg || 'bg-slate-100 dark:bg-slate-700'
                )}>
                  <Clock className={cn('w-4 h-4', actionStyle[entry.action]?.color || 'text-slate-400')} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className={cn(
                      'text-xs font-medium px-1.5 py-0.5 rounded',
                      actionStyle[entry.action]?.badge || 'bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400'
                    )}>
                      {actionLabel[entry.action] || entry.action}
                    </span>
                    {entry.detail && (
                      <span className="text-xs text-slate-400 dark:text-slate-500 truncate">{entry.detail}</span>
                    )}
                  </div>
                  <p className="text-xs text-slate-400 dark:text-slate-500 mt-1">{formatTime(entry.created_at)}</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </Modal>

      <Modal open={showForm} onClose={() => setShowForm(false)} title="新建 API Key">
        <div className="grid grid-cols-1 gap-3">
          <Select
            value={form.provider_id}
            onChange={(e) => setForm({ ...form, provider_id: e.target.value })}
          >
            <option value="">选择 Provider</option>
            {providers.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
          </Select>
          <Input className="font-mono" placeholder="API Key" value={form.key} onChange={(e) => setForm({ ...form, key: e.target.value })} />
          <Input placeholder="备注名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
        </div>
        <div className="flex gap-2 mt-4">
          <Button onClick={handleCreate}>确认添加</Button>
          <Button variant="ghost" onClick={() => setShowForm(false)}>取消</Button>
        </div>
      </Modal>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {keys.map((k) => (
          <div key={k.id} className="bg-white dark:bg-slate-900/80 rounded-2xl border border-slate-200 dark:border-slate-700/60 p-5 shadow-sm hover:shadow-lg dark:hover:shadow-xl transition-all duration-300 overflow-hidden relative">
            <div className="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-violet-500 to-purple-600" />

            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 text-white flex items-center justify-center shrink-0">
                  <KeyRound className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-semibold text-slate-800 dark:text-slate-200 text-sm">{k.name || '未命名'}</h3>
                  <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">{providerMap[k.provider_id] || k.provider_id}</p>
                </div>
              </div>
              <StatusBadge status={k.status} />
            </div>

            <div className="bg-slate-50 dark:bg-slate-800/80 rounded-xl p-3 mb-3">
              <div className="flex items-center justify-between gap-2">
                <code className="text-xs font-mono text-slate-600 dark:text-slate-400 flex-1 truncate">
                  {decrypted[k.id] ? decrypted[k.id] : `${k.key_hash.slice(0, 8)}••••••••`}
                </code>
                <div className="flex items-center gap-1 shrink-0">
                  <button type="button" onClick={() => handleDecrypt(k.id)} className="p-1.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-700 transition" title={decrypted[k.id] ? '隐藏密钥' : '查看密钥'}>
                    {decrypted[k.id] ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                  </button>
                  {decrypted[k.id] && (
                    <button type="button" onClick={() => handleCopy(k.id, decrypted[k.id])} className="p-1.5 text-slate-400 hover:text-blue-600 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-950/30 transition" title="复制">
                      {copied === k.id ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
                    </button>
                  )}
                </div>
              </div>
            </div>

            <div className="flex items-center justify-between mb-3">
              <div>
                <p className="text-xs text-slate-400 dark:text-slate-500">余额</p>
                <p className="text-lg font-bold text-slate-800 dark:text-slate-200">{formatUSD(k.balance_usd)}</p>
              </div>
              <Badge variant={sourceBadgeVariant[k.source] || 'muted'}>{k.source}</Badge>
            </div>

            <div className="flex gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
              <button type="button" onClick={() => handleAudit(k)} className="flex-1 px-3 py-1.5 text-xs text-slate-500 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-950/30 rounded-lg transition flex items-center justify-center gap-1">
                <History className="w-3 h-3" /> 审计
              </button>
              <button type="button" onClick={async () => { await api.keys.revoke(k.id); load() }} className="flex-1 px-3 py-1.5 text-xs text-slate-500 hover:text-amber-600 hover:bg-amber-50 dark:hover:bg-amber-950/30 rounded-lg transition">撤销</button>
              <button type="button" onClick={async () => { await api.keys.delete(k.id); load() }} className="flex-1 px-3 py-1.5 text-xs text-slate-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/30 rounded-lg transition">删除</button>
            </div>
          </div>
        ))}
        {keys.length === 0 && (
          <div className="col-span-full text-center py-16 bg-white dark:bg-slate-900/80 rounded-2xl border border-slate-200 dark:border-slate-700/60 border-dashed">
            <KeyRound className="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
            <p className="text-sm text-slate-400">暂无 API Key</p>
            <p className="text-xs text-slate-300 dark:text-slate-600 mt-1">点击上方按钮添加你的第一个密钥</p>
          </div>
        )}
      </div>
    </div>
  )
}

const sourceBadgeVariant: Record<string, 'default' | 'success' | 'warning' | 'info' | 'muted'> = {
  ccswitch: 'info',
  jsonl: 'default',
  syncer: 'warning',
  manual: 'muted',
}

function StatusBadge({ status }: { status: string }) {
  const variantMap: Record<string, 'success' | 'danger' | 'warning' | 'muted'> = {
    active: 'success',
    revoked: 'danger',
    expired: 'warning',
    invalid: 'muted',
  }
  return <Badge variant={variantMap[status] || 'muted'}>{status}</Badge>
}

const actionLabel: Record<string, string> = {
  created: '已创建',
  revoked: '已撤销',
  deleted: '已删除',
  auto_imported: '自动导入',
}

const actionStyle: Record<string, { bg: string; color: string; badge: string }> = {
  created: { bg: 'bg-emerald-100 dark:bg-emerald-950/40', color: 'text-emerald-500', badge: 'bg-emerald-50 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400' },
  revoked: { bg: 'bg-amber-100 dark:bg-amber-950/40', color: 'text-amber-500', badge: 'bg-amber-50 dark:bg-amber-950/40 text-amber-600 dark:text-amber-400' },
  deleted: { bg: 'bg-red-100 dark:bg-red-950/40', color: 'text-red-500', badge: 'bg-red-50 dark:bg-red-950/40 text-red-600 dark:text-red-400' },
  auto_imported: { bg: 'bg-blue-100 dark:bg-blue-950/40', color: 'text-blue-500', badge: 'bg-blue-50 dark:bg-blue-950/40 text-blue-600 dark:text-blue-400' },
}

function formatTime(ts: string): string {
  if (!ts) return ''
  const d = new Date(ts + (ts.endsWith('Z') ? '' : 'Z'))
  return d.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function LoadingSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <Skeleton className="h-8 w-48" />
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3].map((i) => <Skeleton key={i} className="h-40" />)}
      </div>
    </div>
  )
}
