import { useState, useEffect } from 'react'
import { Send, Loader2, CheckCircle, XCircle, Zap, ExternalLink } from 'lucide-react'
import { api } from '../api'
import { Button, Card, Select } from '../components/ui'

interface APIKey {
  id: string
  provider_id: string
  name: string
  status: string
  balance_usd: number
}

interface Provider {
  id: string
  name: string
  console_url?: string
  docs_url?: string
  topup_url?: string
}

export function Playground() {
  const [keys, setKeys] = useState<APIKey[]>([])
  const [providers, setProviders] = useState<Provider[]>([])
  const [selectedKey, setSelectedKey] = useState('')
  const [prompt, setPrompt] = useState('')
  const [model, setModel] = useState('gpt-4o')
  const [protocol, setProtocol] = useState<'openai' | 'anthropic'>('openai')
  const [response, setResponse] = useState('')
  const [loading, setLoading] = useState(false)
  const [validating, setValidating] = useState(false)
  const [validationResult, setValidationResult] = useState<{ valid: boolean; status?: number } | null>(null)
  const [history, setHistory] = useState<Array<{ prompt: string; response: string; model: string; keyName: string; timestamp: string }>>([])

  useEffect(() => {
    let cancelled = false
    Promise.all([api.keys.list(), api.providers.list()])
      .then(([kList, pList]) => {
        if (cancelled) return
        setKeys(kList)
        setProviders(pList)
        if (kList.length > 0) setSelectedKey(kList[0].id)
      })
      .catch((e) => console.error(e))
    return () => { cancelled = true }
  }, [])

  async function handleTest() {
    if (!selectedKey) return
    setValidating(true); setValidationResult(null)
    try {
      const res = await fetch('/api/v1/playground/validate', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ key_id: selectedKey }) })
      const data = await res.json()
      setValidationResult({ valid: data.valid, status: data.status })
    } catch { setValidationResult({ valid: false }) } finally { setValidating(false) }
  }

  async function handleSend() {
    if (!prompt.trim() || loading || !selectedKey) return
    setLoading(true)
    try {
      const res = await fetch('/api/v1/playground/chat', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ key_id: selectedKey, model, prompt, protocol }) })
      const data = await res.json()
      if (res.ok && data.content) {
        setHistory(prev => [{ prompt, response: data.content, model, keyName: keys.find(k => k.id === selectedKey)?.name || 'Unknown', timestamp: new Date().toLocaleString('zh-CN') }, ...prev])
        setResponse(data.content)
      } else { setResponse(data.error || '请求失败') }
    } catch (err: unknown) { setResponse('请求失败: ' + ((err as Error).message || 'Unknown error')) } finally { setLoading(false) }
  }

  const selectedProvider = keys.find(k => k.id === selectedKey)
  const provider = providers.find(p => p.id === selectedProvider?.provider_id)

  return (
    <div className="h-full flex flex-col max-w-4xl mx-auto">
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-slate-900 dark:text-slate-100 tracking-tight flex items-center gap-2">
          <Zap className="w-5 h-5 text-amber-500" /> API Playground
        </h2>
        <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">测试你的 API Key 和模型</p>
      </div>

      <Card className="mb-4">
        <div className="flex items-end gap-3 flex-wrap">
          <div className="flex-1 min-w-[200px]">
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">选择 API Key</label>
            <Select value={selectedKey} onChange={(e) => { setSelectedKey(e.target.value); setValidationResult(null) }}>
              <option value="">选择 Key...</option>
              {keys.map(k => <option key={k.id} value={k.id}>{k.name} ({k.status})</option>)}
            </Select>
          </div>
          <div className="flex items-end gap-2">
            <Button onClick={handleTest} disabled={!selectedKey || validating} variant="secondary">
              {validating ? <Loader2 className="w-4 h-4 animate-spin" /> : <CheckCircle className="w-4 h-4" />} 测试
            </Button>
            {validationResult && (
              <div className="flex items-center gap-1">
                {validationResult.valid
                  ? <span className="flex items-center gap-1 text-emerald-600 dark:text-emerald-400 text-sm"><CheckCircle className="w-4 h-4" /> 有效</span>
                  : <span className="flex items-center gap-1 text-red-600 dark:text-red-400 text-sm"><XCircle className="w-4 h-4" /> 无效</span>}
              </div>
            )}
          </div>
        </div>
        {provider && (provider.console_url || provider.docs_url || provider.topup_url) && (
          <div className="flex items-center gap-2 mt-4 pt-4 border-t border-slate-100 dark:border-slate-800">
            <span className="text-xs text-slate-400">快捷导航：</span>
            {provider.console_url && <a href={provider.console_url} target="_blank" rel="noreferrer" className="flex items-center gap-1 px-3 py-1.5 bg-slate-50 dark:bg-slate-800 text-slate-600 dark:text-slate-400 rounded-lg text-xs hover:bg-slate-100 dark:hover:bg-slate-700 transition"><ExternalLink className="w-3 h-3" /> 控制台</a>}
            {provider.docs_url && <a href={provider.docs_url} target="_blank" rel="noreferrer" className="flex items-center gap-1 px-3 py-1.5 bg-slate-50 dark:bg-slate-800 text-slate-600 dark:text-slate-400 rounded-lg text-xs hover:bg-slate-100 dark:hover:bg-slate-700 transition"><ExternalLink className="w-3 h-3" /> 文档</a>}
            {provider.topup_url && <a href={provider.topup_url} target="_blank" rel="noreferrer" className="flex items-center gap-1 px-3 py-1.5 bg-emerald-50 dark:bg-emerald-950/30 text-emerald-600 dark:text-emerald-400 rounded-lg text-xs hover:bg-emerald-100 dark:hover:bg-emerald-950/50 transition"><ExternalLink className="w-3 h-3" /> 充值</a>}
          </div>
        )}
      </Card>

      {response && (
        <div className="bg-slate-900 dark:bg-slate-950 rounded-2xl p-5 mb-4 overflow-auto max-h-64 shadow-sm">
          <div className="flex items-center gap-2 mb-3 text-slate-500 text-sm"><span className="font-mono text-xs">Response</span></div>
          <pre className="whitespace-pre-wrap text-sm text-slate-300 font-mono leading-relaxed">{response}</pre>
        </div>
      )}

      {history.length > 0 && (
        <div className="flex-1 overflow-auto mb-4 space-y-3">
          <h3 className="text-sm font-medium text-slate-600 dark:text-slate-400">历史记录</h3>
          {history.map((h, i) => (
            <Card key={i}>
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-slate-400">{h.timestamp}</span>
                <span className="text-xs px-2 py-0.5 bg-slate-100 dark:bg-slate-800 rounded text-slate-600 dark:text-slate-400">{h.model}</span>
              </div>
              <p className="text-sm text-slate-700 dark:text-slate-300 mb-2 font-medium">{h.prompt}</p>
              <pre className="text-xs text-slate-500 dark:text-slate-400 whitespace-pre-wrap leading-relaxed">{h.response}</pre>
            </Card>
          ))}
        </div>
      )}

      <div className="mt-auto">
        <div className="flex gap-2">
          <div className="flex-1">
            <textarea
              className="w-full border border-slate-200 dark:border-slate-700 bg-white/80 dark:bg-slate-800/80 rounded-2xl px-4 py-3 text-sm text-slate-900 dark:text-slate-100 resize-none focus:outline-none focus:ring-2 focus:ring-indigo-500/30 dark:focus:ring-indigo-400/30 transition placeholder:text-slate-400 dark:placeholder:text-slate-500"
              rows={3} placeholder="输入你的提示词..." value={prompt} onChange={(e) => setPrompt(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) handleSend() }}
            />
          </div>
          <div className="flex flex-col gap-2">
            <Select value={protocol} onChange={(e) => { const p = e.target.value as 'openai' | 'anthropic'; setProtocol(p); setModel(p === 'anthropic' ? 'claude-sonnet-4' : 'gpt-4o') }}>
              <option value="openai">OpenAI</option><option value="anthropic">Anthropic</option>
            </Select>
            <Select value={model} onChange={(e) => setModel(e.target.value)}>
              {protocol === 'anthropic'
                ? <><option value="claude-sonnet-4">Claude Sonnet</option><option value="claude-haiku-4">Claude Haiku</option></>
                : <><option value="gpt-4o">GPT-4o</option><option value="gpt-4o-mini">GPT-4o Mini</option><option value="deepseek-chat">DeepSeek</option></>}
            </Select>
            <Button onClick={handleSend} disabled={loading || !prompt.trim() || !selectedKey} loading={loading}>
              <Send className="w-4 h-4" /> 发送
            </Button>
          </div>
        </div>
        <p className="text-xs text-slate-400 dark:text-slate-500 mt-2">按 Ctrl+Enter 快速发送</p>
      </div>
    </div>
  )
}
