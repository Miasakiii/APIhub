import { useEffect, useState } from 'react'
import { Plus, Trash2, Server, Link2, CreditCard, BookOpen, Globe } from 'lucide-react'
import { api, type Provider } from '../api'
import { PageHeader, Button, Modal, Input, Skeleton } from '../components/ui'
import { cn } from '../lib/utils'

export function Providers() {
  const [providers, setProviders] = useState<Provider[]>([])
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', type: '', base_url: '' })
  const [loading, setLoading] = useState(true)

  useEffect(() => { load() }, [])

  async function load() {
    const list = await api.providers.list()
    setProviders(list)
    setLoading(false)
  }

  async function handleCreate() {
    if (!form.name || !form.type) return
    await api.providers.create(form)
    setForm({ name: '', type: '', base_url: '' })
    setShowForm(false)
    load()
  }

  async function handleDelete(id: string) {
    await api.providers.delete(id)
    load()
  }

  if (loading) return <LoadingSkeleton />

  return (
    <div className="space-y-6">
      <PageHeader
        title="Providers"
        description="管理你的 LLM 服务商"
        actions={
          <Button onClick={() => setShowForm(true)}>
            <Plus className="w-4 h-4" /> 添加 Provider
          </Button>
        }
      />

      <Modal open={showForm} onClose={() => setShowForm(false)} title="新建 Provider">
        <div className="grid grid-cols-1 gap-3">
          <Input placeholder="名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          <Input placeholder="类型 (claude/openai/codex...)" value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })} />
          <Input placeholder="Base URL" value={form.base_url} onChange={(e) => setForm({ ...form, base_url: e.target.value })} />
        </div>
        <div className="flex gap-2 mt-4">
          <Button onClick={handleCreate}>确认添加</Button>
          <Button variant="ghost" onClick={() => setShowForm(false)}>取消</Button>
        </div>
      </Modal>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {providers.map((p) => (
          <ProviderCard key={p.id} provider={p} onDelete={handleDelete} />
        ))}
        {providers.length === 0 && (
          <div className="col-span-full text-center py-16 bg-white dark:bg-slate-900/80 rounded-2xl border border-slate-200 dark:border-slate-700/60 border-dashed">
            <Server className="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
            <p className="text-sm text-slate-400">暂无 Provider</p>
            <p className="text-xs text-slate-300 dark:text-slate-600 mt-1">点击上方按钮添加你的第一个服务商</p>
          </div>
        )}
      </div>
    </div>
  )
}

const typeColors: Record<string, string> = {
  claude: 'from-amber-500 to-orange-600',
  openai: 'from-emerald-500 to-teal-600',
  gemini: 'from-blue-500 to-indigo-600',
  codex: 'from-violet-500 to-purple-600',
}

function ProviderCard({ provider: p, onDelete }: { provider: Provider; onDelete: (id: string) => void }) {
  const gradient = typeColors[p.type] || 'from-blue-500 to-indigo-600'

  return (
    <div className="bg-white dark:bg-slate-900/80 rounded-2xl border border-slate-200 dark:border-slate-700/60 p-5 shadow-sm hover:shadow-lg dark:hover:shadow-xl transition-all duration-300 group overflow-hidden relative">
      {/* Top accent bar */}
      <div className={cn('absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r', gradient)} />

      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          <div className={cn('w-10 h-10 rounded-xl bg-gradient-to-br text-white flex items-center justify-center font-bold text-sm shrink-0', gradient)}>
            {p.name.slice(0, 2).toUpperCase()}
          </div>
          <div>
            <h3 className="font-semibold text-slate-800 dark:text-slate-200">{p.name}</h3>
            <span className="inline-block mt-0.5 px-2 py-0.5 bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 text-xs rounded-md font-mono">{p.type}</span>
          </div>
        </div>
        <button
          type="button"
          aria-label="删除 Provider"
          className="text-slate-300 dark:text-slate-600 hover:text-red-500 transition p-1 rounded-lg hover:bg-red-50 dark:hover:bg-red-950/30"
          onClick={() => onDelete(p.id)}
        >
          <Trash2 className="w-4 h-4" />
        </button>
      </div>

      {p.base_url && (
        <div className="flex items-center gap-1.5 mb-3">
          <Globe className="w-3.5 h-3.5 text-slate-400" />
          <p className="text-xs text-slate-400 dark:text-slate-500 font-mono truncate">{p.base_url}</p>
        </div>
      )}

      <div className="flex flex-wrap gap-2 mt-3 pt-3 border-t border-slate-100 dark:border-slate-800">
        {p.console_url && (
          <a href={p.console_url} target="_blank" rel="noreferrer" className="flex items-center gap-1.5 px-3 py-1.5 bg-slate-50 dark:bg-slate-800 text-slate-600 dark:text-slate-400 rounded-lg text-xs hover:bg-blue-50 dark:hover:bg-blue-950/30 hover:text-blue-600 transition">
            <Link2 className="w-3 h-3" /> 控制台
          </a>
        )}
        {p.topup_url && (
          <a href={p.topup_url} target="_blank" rel="noreferrer" className="flex items-center gap-1.5 px-3 py-1.5 bg-emerald-50 dark:bg-emerald-950/30 text-emerald-600 dark:text-emerald-400 rounded-lg text-xs hover:bg-emerald-100 dark:hover:bg-emerald-950/50 transition">
            <CreditCard className="w-3 h-3" /> 充值
          </a>
        )}
        {p.docs_url && (
          <a href={p.docs_url} target="_blank" rel="noreferrer" className="flex items-center gap-1.5 px-3 py-1.5 bg-violet-50 dark:bg-violet-950/30 text-violet-600 dark:text-violet-400 rounded-lg text-xs hover:bg-violet-100 dark:hover:bg-violet-950/50 transition">
            <BookOpen className="w-3 h-3" /> 文档
          </a>
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
        {[1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-40" />
        ))}
      </div>
    </div>
  )
}
