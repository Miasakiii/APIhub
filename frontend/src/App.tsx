import { useEffect, useState } from 'react'
import { Settings as SettingsPage } from './pages/Settings'
import { Dashboard } from './pages/Dashboard'
import { Providers } from './pages/Providers'
import { Keys } from './pages/Keys'
import { UsageLog } from './pages/UsageLog'
import { Alerts } from './pages/Alerts'
import { Subscriptions } from './pages/Subscriptions'
import { Frequency } from './pages/Frequency'
import { Playground } from './pages/Playground'
import { Login } from './pages/Login'
import { ModelDetail } from './pages/ModelDetail'
import { api, setUnauthorizedHandler } from './api'
import { clearToken, isAuthed } from './lib/auth'
import { ThemeProvider } from './lib/theme'
import { ToastProvider } from './components/ui'
import { Sidebar } from './components/layout/Sidebar'
import { TopBar } from './components/layout/TopBar'
import { allNav } from './lib/nav'

function AppShell() {
  const [page, setPage] = useState('dashboard')
  const [selectedModel, setSelectedModel] = useState('')
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [authEnabled, setAuthEnabled] = useState<boolean | null>(null)
  const [allowRegister, setAllowRegister] = useState(true)
  const [authed, setAuthed] = useState(isAuthed())

  useEffect(() => {
    setUnauthorizedHandler(() => setAuthed(false))
    api.auth.config()
      .then((cfg) => {
        setAuthEnabled(cfg.enabled)
        setAllowRegister(cfg.allow_register)
        if (!cfg.enabled) setAuthed(true)
      })
      .catch(() => {
        setAuthEnabled(false)
        setAuthed(true)
      })
  }, [])

  if (authEnabled === null) {
    return (
      <div className="min-h-screen flex items-center justify-center app-shell-bg">
        <div className="flex flex-col items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500 to-violet-600 animate-pulse shadow-lg shadow-indigo-500/30" />
          <p className="text-sm text-slate-500 dark:text-slate-400">加载中...</p>
        </div>
      </div>
    )
  }

  if (authEnabled && !authed) {
    return <Login onSuccess={() => setAuthed(true)} allowRegister={allowRegister} />
  }

  const currentLabel = page === 'model-detail' ? selectedModel : allNav.find((n) => n.id === page)?.label ?? ''

  return (
    <div className="flex h-screen overflow-hidden app-shell-bg">
      <Sidebar
        page={page}
        setPage={setPage}
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
      />

      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <TopBar
          currentLabel={currentLabel}
          authEnabled={!!authEnabled}
          onLogout={
            authEnabled
              ? () => {
                  clearToken()
                  setAuthed(false)
                }
              : undefined
          }
          onMenuOpen={() => setSidebarOpen(true)}
        />

        <div className="flex-1 overflow-auto p-4 lg:p-8">
          <div className="max-w-7xl mx-auto page-enter" key={page}>
            {page === 'dashboard' && <Dashboard onModelClick={(m) => { setSelectedModel(m); setPage('model-detail') }} />}
            {page === 'model-detail' && <ModelDetail model={selectedModel} onBack={() => setPage('dashboard')} />}
            {page === 'providers' && <Providers />}
            {page === 'keys' && <Keys />}
            {page === 'usage' && <UsageLog />}
            {page === 'alerts' && <Alerts />}
            {page === 'subscriptions' && <Subscriptions />}
            {page === 'frequency' && <Frequency />}
            {page === 'playground' && <Playground />}
            {page === 'settings' && (
              <SettingsPage
                onLogout={
                  authEnabled
                    ? () => {
                        clearToken()
                        setAuthed(false)
                      }
                    : undefined
                }
              />
            )}
          </div>
        </div>
      </main>
    </div>
  )
}

export default function App() {
  return (
    <ThemeProvider>
      <ToastProvider>
        <AppShell />
      </ToastProvider>
    </ThemeProvider>
  )
}
