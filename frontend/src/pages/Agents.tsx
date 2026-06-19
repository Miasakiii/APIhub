import { useEffect, useState } from 'react'
import { Plus, Trash2, Bot, Terminal, Globe, Code } from 'lucide-react'
import { api, type Agent } from '../api'
import { PageHeader, Button, Modal, Input, Select, Skeleton } from '../components/ui'

const typeMap: Record<string, { label: string; icon: typeof Bot }> = {
  cli: { label: 'CLI 工具', icon: Terminal },
  ide: { label: 'IDE 插件', icon: Code },
  api: { label: 'API 调用', icon: Globe },
  proxy: { label: '代理', icon: Globe },
}

export function Agents() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState<Partial<Agent>>({ name: '', type: 'cli' })
  const [loading, setLoading] = useState(true)

  useEffect(() => { load() }, [])

  async function load() {
    setLoading(true)
    try { setAgents(await api.agents.list()) } catch (e) { console.error(e) } finally { setLoading(false) }
  }

  async function handleCreate() {
    if (!form.name) return
    await api.agents.create(form)
    setForm({ name: '', type: 'cli' })
    setShowForm(false)
    load()
  }

  async function handleDelete(id: string) {
    await api.agents.delete(id)
    load()
  }

  if (loading) return <LoadingSkeleton />

  return (
    <div className="space-y-6">
      <PageHeader
        title="Agent 管理"
        description={`共 ${agents.length} 个 Agent`}
        actions={<Button onClick={() => setShowForm(true)}><Plus className="w-4 h-4" /> 添加 Agent</Button>}
      />

      <Modal open={showForm} onClose={() => setShowForm(false)} title="新建 Agent">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <Input
            placeholder="Agent 名称"
            value={form.name || ''}
            onChange={e => setForm({ ...form, name: e.target.value })}
          />
          <Select value={form.type} onChange={e => setForm({ ...form, type: e.target.value })}>
            <option value="cli">CLI 工具</option>
            <option value="ide">IDE 插件</option>
            <option value="api">API 调用</option>
            <option value="proxy">代理</option>
          </Select>
        </div>
        <div className="flex gap-2 mt-4">
          <Button onClick={handleCreate}>确认添加</Button>
          <Button variant="ghost" onClick={() => setShowForm(false)}>取消</Button>
        </div>
      </Modal>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {agents.map(a => {
          const t = typeMap[a.type] || { label: a.type, icon: Bot }
          const Icon = t.icon
          return (
            <div key={a.id} className="bg-white dark:bg-slate-900/80 rounded-2xl border border-slate-200 dark:border-slate-700/60 p-5 shadow-sm hover:shadow-lg dark:hover:shadow-xl transition-all duration-300 overflow-hidden relative">
              <div className="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-emerald-500 to-teal-600" />
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 text-white flex items-center justify-center shrink-0">
                    <Icon className="w-5 h-5" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-slate-800 dark:text-slate-200 text-sm">{a.name}</h3>
                    <p className="text-xs text-slate-400 dark:text-slate-500">{t.label}</p>
                  </div>
                </div>
              </div>
              <div className="flex gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
                <button type="button" onClick={() => handleDelete(a.id)} className="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs text-slate-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/30 rounded-lg transition">
                  <Trash2 className="w-3.5 h-3.5" /> 删除
                </button>
              </div>
            </div>
          )
        })}
        {agents.length === 0 && (
          <div className="col-span-full text-center py-16 bg-white dark:bg-slate-900/80 rounded-2xl border border-slate-200 dark:border-slate-700/60 border-dashed">
            <Bot className="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
            <p className="text-sm text-slate-400">暂无 Agent</p>
            <p className="text-xs text-slate-300 dark:text-slate-600 mt-1">添加 Agent 后可按工具追踪用量</p>
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
        {[1, 2, 3].map(i => <Skeleton key={i} className="h-36 rounded-2xl" />)}
      </div>
    </div>
  )
}
