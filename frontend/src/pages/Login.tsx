import { useState } from 'react'
import { KeyRound, LogIn, UserPlus, Sparkles } from 'lucide-react'
import { api } from '../api'
import { setToken } from '../lib/auth'
import { Button, Input } from '../components/ui'

interface LoginProps {
  onSuccess: () => void
  allowRegister?: boolean
}

export function Login({ onSuccess, allowRegister = true }: LoginProps) {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      if (mode === 'login') {
        const res = await api.auth.login(username, password)
        setToken(res.token, res.username)
      } else {
        const res = await api.auth.register(username, password)
        if (res.token) {
          setToken(res.token, res.username)
        } else {
          const loginRes = await api.auth.login(username, password)
          setToken(loginRes.token, loginRes.username)
        }
      }
      onSuccess()
    } catch (err) {
      setError(err instanceof Error ? err.message : '操作失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex">
      <div className="hidden lg:flex lg:w-1/2 relative bg-slate-950 overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-indigo-600/40 via-violet-600/20 to-slate-950" />
        <div className="absolute top-1/4 -left-20 w-72 h-72 bg-indigo-500/30 rounded-full blur-3xl" />
        <div className="absolute bottom-1/4 right-0 w-96 h-96 bg-violet-500/20 rounded-full blur-3xl" />
        <div className="relative z-10 flex flex-col justify-center px-16 text-white">
          <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center shadow-2xl shadow-indigo-500/40 mb-8">
            <Sparkles className="w-7 h-7" />
          </div>
          <h1 className="text-4xl font-bold tracking-tight mb-4">APIHub</h1>
          <p className="text-lg text-slate-300 max-w-md leading-relaxed">
            集中监控 LLM API 用量、费用与告警。三层数据源，一个仪表盘。
          </p>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center p-6 app-shell-bg">
        <div className="w-full max-w-md">
          <div className="lg:hidden flex items-center gap-3 mb-8">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center">
              <KeyRound className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">APIHub</h1>
              <p className="text-sm text-slate-500 dark:text-slate-400">登录以继续</p>
            </div>
          </div>

          <div className="bg-white/90 dark:bg-slate-900/90 backdrop-blur-sm rounded-2xl border border-slate-200/80 dark:border-slate-700/60 shadow-xl shadow-slate-200/50 dark:shadow-black/20 p-8">
            <h2 className="text-xl font-bold text-slate-900 dark:text-slate-100 mb-1 hidden lg:block">
              {mode === 'login' ? '欢迎回来' : '创建账号'}
            </h2>
            <p className="text-sm text-slate-500 dark:text-slate-400 mb-6 hidden lg:block">
              {mode === 'login' ? '登录你的 APIHub 控制台' : '注册后即可使用全部功能'}
            </p>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">用户名</label>
                <Input value={username} onChange={(e) => setUsername(e.target.value)} required autoComplete="username" placeholder="输入用户名" />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">密码</label>
                <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required autoComplete={mode === 'login' ? 'current-password' : 'new-password'} placeholder="输入密码" />
              </div>

              {error && <p className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950/30 border border-red-100 dark:border-red-900/50 rounded-xl px-3 py-2.5">{error}</p>}

              <Button type="submit" disabled={loading} loading={loading} className="w-full">
                {mode === 'login' ? <LogIn className="w-4 h-4" /> : <UserPlus className="w-4 h-4" />}
                {loading ? '处理中...' : mode === 'login' ? '登录' : '注册'}
              </Button>
            </form>

            {allowRegister && (
              <p className="text-center text-sm text-slate-500 dark:text-slate-400 mt-6">
                {mode === 'login' ? (
                  <>还没有账号？{' '}<button type="button" className="text-indigo-600 dark:text-indigo-400 font-medium hover:underline" onClick={() => setMode('register')}>注册</button></>
                ) : (
                  <>已有账号？{' '}<button type="button" className="text-indigo-600 dark:text-indigo-400 font-medium hover:underline" onClick={() => setMode('login')}>登录</button></>
                )}
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
