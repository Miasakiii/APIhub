import { lazy, Suspense, useEffect, useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { api, isWailsEnv, setUnauthorizedHandler } from './api'
import { clearToken, isAuthed } from './lib/auth'
import { ThemeProvider } from './lib/theme'
import { CompactModeProvider } from './lib/compact-mode'
import { useCompactShortcut } from './lib/use-compact-shortcut'
import { ToastProvider } from './components/ui'
import { useToast } from './lib/use-toast'
import { WebSocketProvider } from './lib/use-ws.tsx'
import { useWSMessage } from './lib/use-ws'
import type { AlertData } from './lib/ws-types'
import { Sidebar } from './components/layout/Sidebar'
import { TopBar } from './components/layout/TopBar'
import { Login } from './pages/Login'

// Route-level code splitting
const Dashboard = lazy(() => import('./pages/Dashboard').then(m => ({ default: m.Dashboard })))
const ModelDetail = lazy(() => import('./pages/ModelDetail').then(m => ({ default: m.ModelDetail })))
const UsageLog = lazy(() => import('./pages/UsageLog').then(m => ({ default: m.UsageLog })))
const Frequency = lazy(() => import('./pages/Frequency').then(m => ({ default: m.Frequency })))
const SettingsPage = lazy(() => import('./pages/Settings').then(m => ({ default: m.Settings })))

function PageLoader() {
  return (
    <div className="flex items-center justify-center h-64">
      <div className="flex flex-col items-center gap-3">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600 animate-pulse shadow-lg shadow-indigo-500/30" />
        <p className="text-sm text-slate-400">加载中...</p>
      </div>
    </div>
  )
}

// Listens for WebSocket alert events and shows toast notifications
function AlertToaster() {
  const { warning, error } = useToast()
  useWSMessage('alert.triggered', (msg) => {
    const data = msg.data as AlertData
    if (data.level === 'critical') {
      error(`🚨 ${data.title}: ${data.message}`)
    } else {
      warning(`⚠️ ${data.title}: ${data.message}`)
    }
  })
  return null
}

function CompactShortcutListener() {
  useCompactShortcut()
  return null
}

function AppLayout({ onLogout, authEnabled }: { onLogout?: () => void; authEnabled: boolean }) {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const location = useLocation()

  return (
      <CompactModeProvider>
        <CompactShortcutListener />
        <WebSocketProvider>
          <AlertToaster />
          <div className="flex h-screen overflow-hidden app-shell-bg">
            <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} />

            <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
              <TopBar
                authEnabled={authEnabled}
                onLogout={onLogout}
                onMenuOpen={() => setSidebarOpen(true)}
              />

              <div className="flex-1 overflow-auto p-4 lg:p-8">
                <div className="max-w-7xl mx-auto page-enter" key={location.pathname}>
                  <Suspense fallback={<PageLoader />}>
                    <Routes>
                      <Route path="/" element={<Dashboard />} />
                      <Route path="/model/:model" element={<ModelDetail />} />
                      <Route path="/usage" element={<UsageLog />} />
                      <Route path="/frequency" element={<Frequency />} />
                      <Route path="/settings" element={<SettingsPage />} />
                      <Route path="*" element={<Navigate to="/" replace />} />
                    </Routes>
                  </Suspense>
                </div>
              </div>
            </main>
          </div>
        </WebSocketProvider>
      </CompactModeProvider>
  )
}

function AppShell() {
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

  // Wails native notification listener
  useEffect(() => {
    if (!isWailsEnv()) return
    window.runtime?.EventsOn('notification', (data: { title: string; message: string }) => {
      if ('Notification' in window && Notification.permission === 'granted') {
        new Notification(data.title, { body: data.message })
      }
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

  const handleLogout = authEnabled
    ? () => { clearToken(); setAuthed(false) }
    : undefined

  return <AppLayout onLogout={handleLogout} authEnabled={!!authEnabled} />
}

export default function App() {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <ToastProvider>
          <AppShell />
        </ToastProvider>
      </ThemeProvider>
    </BrowserRouter>
  )
}
