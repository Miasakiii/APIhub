import { useState, useEffect } from 'react'
import { Settings as SettingsIcon, User, Shield, Database, Bell, Sun, Moon, Languages, DollarSign, Trash2, Download, Scan, CheckCircle, XCircle, AlertCircle, Monitor } from 'lucide-react'
import { cn } from '../lib/utils'
import { api, isWailsEnv } from '../api'
import type { ScanFinding, ScanImportResult } from '../api'
import { useTheme } from '../lib/use-theme'
import { useCompactMode } from '../lib/compact-mode'
import type { CompactMode } from '../lib/compact-mode'
import { Card, Button, Select } from '../components/ui'

interface SettingsProps { onLogout?: () => void }

export function Settings({ onLogout }: SettingsProps) {
  const [activeTab, setActiveTab] = useState('general')

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">设置</h1>
        <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">管理你的应用偏好</p>
      </div>
      <Card padding={false}>
        <div className="flex flex-col lg:flex-row">
          <div className="lg:w-56 border-b lg:border-b-0 lg:border-r border-slate-100 dark:border-slate-800 p-3 space-y-1">
            <TabButton id="general" label="通用" icon={SettingsIcon} active={activeTab} onClick={setActiveTab} />
            <TabButton id="account" label="账号" icon={User} active={activeTab} onClick={setActiveTab} />
            <TabButton id="security" label="安全" icon={Shield} active={activeTab} onClick={setActiveTab} />
            <TabButton id="notifications" label="通知" icon={Bell} active={activeTab} onClick={setActiveTab} />
            <TabButton id="data" label="数据" icon={Database} active={activeTab} onClick={setActiveTab} />
            {isWailsEnv() && <TabButton id="desktop" label="桌面端" icon={Monitor} active={activeTab} onClick={setActiveTab} />}
          </div>
          <div className="flex-1 p-6 lg:p-8">
            {activeTab === 'general' && <GeneralSettings />}
            {activeTab === 'account' && <AccountSettings onLogout={onLogout} />}
            {activeTab === 'security' && <SecuritySettings />}
            {activeTab === 'notifications' && <NotificationSettings />}
            {activeTab === 'data' && <DataSettings />}
            {activeTab === 'desktop' && <DesktopSettings />}
          </div>
        </div>
      </Card>
    </div>
  )
}

function TabButton({ id, label, icon: Icon, active, onClick }: {
  id: string; label: string; icon: React.ComponentType<{ className?: string }>; active: string; onClick: (id: string) => void
}) {
  return (
    <button type="button" onClick={() => onClick(id)}
      className={cn('w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition',
        active === id ? 'bg-indigo-50 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400 font-medium' : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800'
      )}>
      <Icon className={cn('w-4 h-4', active === id ? 'text-indigo-500 dark:text-indigo-400' : 'text-slate-400 dark:text-slate-500')} />
      {label}
    </button>
  )
}

function GeneralSettings() {
  const { theme, setTheme } = useTheme()
  const { mode: compactMode, setMode: setCompactMode } = useCompactMode()
  return (
    <div className="space-y-6">
      <h2 className="text-lg font-bold text-slate-900 dark:text-slate-100">通用设置</h2>
      <div className="space-y-4">
        <SettingRow icon={Languages} label="语言" description="当前仅支持中文">
          <Select className="min-w-[120px]" title="语言"><option value="zh">中文</option><option value="en" disabled>English</option></Select>
        </SettingRow>
        <SettingRow icon={DollarSign} label="货币" description="费用显示货币">
          <Select className="min-w-[120px]" title="货币"><option value="USD">USD ($)</option><option value="CNY">CNY (¥)</option></Select>
        </SettingRow>
        <SettingRow icon={theme === 'dark' ? Moon : Sun} label="主题" description="切换亮色/暗色主题">
          <Select className="min-w-[120px]" value={theme} onChange={(e) => setTheme(e.target.value as 'light' | 'dark')} title="主题">
            <option value="light">亮色</option><option value="dark">暗色</option>
          </Select>
        </SettingRow>
        <SettingRow icon={Monitor} label="侧边栏模式" description="切换侧边栏显示方式">
          <Select
            className="min-w-[120px]"
            value={compactMode}
            onChange={(e) => setCompactMode(e.target.value as CompactMode)}
            title="侧边栏模式"
          >
            <option value="full">完整</option>
            <option value="icons">仅图标</option>
            <option value="hidden">隐藏</option>
          </Select>
        </SettingRow>
      </div>
    </div>
  )
}

function AccountSettings({ onLogout }: { onLogout?: () => void }) {
  return (
    <div className="space-y-6">
      <h2 className="text-lg font-bold text-slate-900 dark:text-slate-100">账号设置</h2>
      <div className="space-y-4">
        {onLogout && (
          <SettingRow icon={User} label="退出登录" description="清除本地登录状态">
            <Button variant="secondary" onClick={onLogout}>退出</Button>
          </SettingRow>
        )}
        <SettingRow icon={Shield} label="密码" description="v0.2 暂不支持在线改密">
          <span className="text-xs text-slate-400">—</span>
        </SettingRow>
      </div>
    </div>
  )
}

function SecuritySettings() {
  return (
    <div className="space-y-6">
      <h2 className="text-lg font-bold text-slate-900 dark:text-slate-100">安全设置</h2>
      <div className="space-y-4">
        <SettingRow icon={Shield} label="主密钥" description="用于加密存储的 API Key">
          <Button variant="secondary" size="sm">查看</Button>
        </SettingRow>
        <SettingRow icon={Shield} label="API Key 加密" description="所有 Key 使用 AES-256-GCM 加密">
          <span className="inline-flex items-center px-3 py-1 bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-400 text-xs rounded-full border border-emerald-100 dark:border-emerald-800/50 font-medium">已启用</span>
        </SettingRow>
      </div>
    </div>
  )
}

function NotificationSettings() {
  return (
    <div className="space-y-6">
      <h2 className="text-lg font-bold text-slate-900 dark:text-slate-100">通知设置</h2>
      <div className="space-y-4">
        <SettingRow icon={Bell} label="余额告警" description="余额低于阈值时发送通知"><ToggleSwitch defaultChecked /></SettingRow>
        <SettingRow icon={Calendar} label="订阅到期" description="订阅即将到期时发送通知"><ToggleSwitch defaultChecked /></SettingRow>
      </div>
    </div>
  )
}

function DataSettings() {
  const [scanState, setScanState] = useState<'idle' | 'scanning' | 'confirm' | 'importing' | 'done'>('idle')
  const [findings, setFindings] = useState<(ScanFinding & { selected?: boolean })[]>([])
  const [importResults, setImportResults] = useState<ScanImportResult[]>([])

  const handleScan = () => {
    setScanState('scanning')
    api.scan.run()
      .then(res => {
        setFindings(res.findings.map(f => ({ ...f, selected: true })))
        setScanState('confirm')
      })
      .catch(() => setScanState('idle'))
  }

  const handleImport = () => {
    const selectedIndices = findings
      .map((f, i) => f.selected ? i : -1)
      .filter(i => i >= 0)
    if (selectedIndices.length === 0) return
    setScanState('importing')
    api.scan.import(selectedIndices)
      .then(res => {
        setImportResults(res.results)
        setScanState('done')
      })
      .catch(() => setScanState('confirm'))
  }

  const toggleFinding = (index: number) => {
    setFindings(prev => prev.map((f, i) => i === index ? { ...f, selected: !f.selected } : f))
  }

  return (
    <div className="space-y-6">
      <h2 className="text-lg font-bold text-slate-900 dark:text-slate-100">数据管理</h2>
      <div className="space-y-4">
        {/* Scan Local Config */}
        <div className="border border-slate-100 dark:border-slate-800 rounded-xl p-4">
          <div className="flex items-center gap-3 mb-3">
            <Scan className="w-4 h-4 text-slate-400 dark:text-slate-500" />
            <div>
              <p className="font-medium text-sm text-slate-800 dark:text-slate-200">扫描本地配置</p>
              <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">自动检测 Claude Code、DeepSeek、Kimi 等工具的 API Key</p>
            </div>
          </div>

          {scanState === 'idle' && (
            <Button onClick={handleScan}><Scan className="w-4 h-4" /> 开始扫描</Button>
          )}

          {scanState === 'scanning' && (
            <div className="flex items-center gap-2 text-sm text-slate-500">
              <div className="w-4 h-4 border-2 border-indigo-300 border-t-indigo-600 rounded-full animate-spin" />
              扫描中...
            </div>
          )}

          {scanState === 'confirm' && findings.length > 0 && (
            <div className="space-y-3">
              <p className="text-sm text-slate-600 dark:text-slate-400">发现 {findings.length} 个配置：</p>
              <div className="space-y-2 max-h-48 overflow-y-auto">
                {findings.map((f, i) => (
                  <label key={i} className="flex items-center gap-3 p-2 rounded-lg bg-slate-50 dark:bg-slate-800/50 cursor-pointer">
                    <input type="checkbox" checked={f.selected} onChange={() => toggleFinding(i)}
                      className="w-4 h-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-slate-800 dark:text-slate-200">{f.name}</p>
                      <p className="text-xs text-slate-400 dark:text-slate-500 truncate">{f.masked_key}</p>
                    </div>
                    <span className="text-xs px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400">{f.source}</span>
                  </label>
                ))}
              </div>
              <div className="flex gap-2">
                <Button onClick={handleImport} size="sm">导入选中项</Button>
                <Button variant="secondary" size="sm" onClick={() => setScanState('idle')}>取消</Button>
              </div>
            </div>
          )}

          {scanState === 'confirm' && findings.length === 0 && (
            <div className="text-sm text-slate-500 dark:text-slate-400">
              <AlertCircle className="w-4 h-4 inline mr-1" />
              未发现本地 API Key 配置
              <Button variant="secondary" size="sm" className="ml-2" onClick={() => setScanState('idle')}>返回</Button>
            </div>
          )}

          {scanState === 'importing' && (
            <div className="flex items-center gap-2 text-sm text-slate-500">
              <div className="w-4 h-4 border-2 border-indigo-300 border-t-indigo-600 rounded-full animate-spin" />
              导入中...
            </div>
          )}

          {scanState === 'done' && (
            <div className="space-y-3">
              <p className="text-sm font-medium text-slate-800 dark:text-slate-200">导入完成：</p>
              <div className="space-y-1.5">
                {importResults.map((r, i) => (
                  <div key={i} className="flex items-center gap-2 text-sm">
                    {r.status === 'created' && <CheckCircle className="w-4 h-4 text-emerald-500" />}
                    {r.status === 'skipped' && <AlertCircle className="w-4 h-4 text-amber-500" />}
                    {r.status === 'error' && <XCircle className="w-4 h-4 text-red-500" />}
                    <span className="text-slate-700 dark:text-slate-300">{r.name}</span>
                    <span className="text-xs text-slate-400 dark:text-slate-500">
                      {r.status === 'created' ? '已导入' : r.status === 'skipped' ? '已跳过' : r.message}
                    </span>
                  </div>
                ))}
              </div>
              <Button variant="secondary" size="sm" onClick={() => { setScanState('idle'); setFindings([]); setImportResults([]) }}>完成</Button>
            </div>
          )}
        </div>

        <SettingRow icon={Download} label="导出数据" description="导出所有用量记录为 CSV">
          <Button onClick={() => api.export.csv().catch(console.error)}><Download className="w-4 h-4" /> 导出</Button>
        </SettingRow>
        <SettingRow icon={Trash2} label="清除缓存" description="清除本地缓存数据">
          <Button variant="secondary">清除</Button>
        </SettingRow>
        <SettingRow icon={Trash2} label="删除所有数据" description="此操作不可撤销" danger>
          <Button variant="danger">删除</Button>
        </SettingRow>
      </div>
    </div>
  )
}

function SettingRow({ icon: Icon, label, description, children, danger }: {
  icon?: React.ComponentType<{ className?: string }>; label: string; description: string; children: React.ReactNode; danger?: boolean
}) {
  return (
    <div className="flex items-center justify-between py-4 border-b border-slate-100 dark:border-slate-800 last:border-0">
      <div className="flex items-center gap-3">
        {Icon && <Icon className={cn('w-4 h-4', danger ? 'text-red-400' : 'text-slate-400 dark:text-slate-500')} />}
        <div>
          <p className={cn('font-medium text-sm', danger ? 'text-red-600 dark:text-red-400' : 'text-slate-800 dark:text-slate-200')}>{label}</p>
          <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">{description}</p>
        </div>
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  )
}

function DesktopSettings() {
  const wails = window.go?.main?.WailsApp
  const [minimizeToTray, setMinimizeToTray] = useState(false)
  const [autoStart, setAutoStart] = useState(false)
  const [version, setVersion] = useState('')
  const [dataDir, setDataDir] = useState('')

  useEffect(() => {
    if (!wails) return
    wails.GetMinimizeToTray().then(setMinimizeToTray)
    wails.IsAutoStartEnabled().then(setAutoStart)
    wails.GetVersion().then(setVersion)
    wails.GetDataDir().then(setDataDir)
  }, [wails])

  const handleMinimizeToTray = async (enable: boolean) => {
    await wails?.SetMinimizeToTray(enable)
    setMinimizeToTray(enable)
  }

  const handleAutoStart = async (enable: boolean) => {
    await wails?.SetAutoStart(enable)
    setAutoStart(enable)
  }

  if (!wails) return null

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4">桌面端设置</h3>

        <div className="space-y-4">
          {/* Minimize to tray */}
          <div className="flex items-center justify-between p-4 rounded-xl bg-slate-50 dark:bg-slate-800/50">
            <div>
              <p className="text-sm font-medium text-slate-900 dark:text-slate-100">关闭时最小化</p>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">关闭窗口时最小化到任务栏而非退出应用</p>
            </div>
            <ControlledToggle checked={minimizeToTray} onChange={handleMinimizeToTray} />
          </div>

          {/* Auto-start */}
          <div className="flex items-center justify-between p-4 rounded-xl bg-slate-50 dark:bg-slate-800/50">
            <div>
              <p className="text-sm font-medium text-slate-900 dark:text-slate-100">开机自启</p>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">系统启动时自动运行 APIHub</p>
            </div>
            <ControlledToggle checked={autoStart} onChange={handleAutoStart} />
          </div>

          {/* Version */}
          <div className="flex items-center justify-between p-4 rounded-xl bg-slate-50 dark:bg-slate-800/50">
            <div>
              <p className="text-sm font-medium text-slate-900 dark:text-slate-100">版本</p>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">当前应用版本号</p>
            </div>
            <span className="text-sm font-mono text-slate-600 dark:text-slate-300">{version}</span>
          </div>

          {/* Data directory */}
          <div className="flex items-center justify-between p-4 rounded-xl bg-slate-50 dark:bg-slate-800/50">
            <div>
              <p className="text-sm font-medium text-slate-900 dark:text-slate-100">数据目录</p>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">数据库和配置文件存储位置</p>
            </div>
            <span className="text-xs font-mono text-slate-500 dark:text-slate-400 max-w-[200px] truncate" title={dataDir}>{dataDir}</span>
          </div>
        </div>
      </div>
    </div>
  )
}

function ControlledToggle({ checked, onChange }: { checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <button type="button" role="switch" aria-checked={checked} onClick={() => onChange(!checked)}
      className={cn('relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
        checked ? 'bg-indigo-600' : 'bg-slate-200 dark:bg-slate-700'
      )}>
      <span className={cn('inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
        checked ? 'translate-x-6' : 'translate-x-1'
      )} />
    </button>
  )
}

function ToggleSwitch({ defaultChecked = false }: { defaultChecked?: boolean }) {
  return (
    <label className="relative inline-flex items-center cursor-pointer">
      <input type="checkbox" aria-label="开关" className="sr-only peer" defaultChecked={defaultChecked} />
      <div className="w-11 h-6 bg-slate-200 dark:bg-slate-700 peer-focus:ring-4 peer-focus:ring-indigo-300 dark:peer-focus:ring-indigo-800 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-600" />
    </label>
  )
}

function Calendar({ className }: { className?: string }) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
      <rect width="18" height="18" x="3" y="4" rx="2" ry="2" /><line x1="16" x2="16.01" y1="2" y2="2" /><line x1="8" x2="8.01" y1="2" y2="2" /><line x1="3" x2="21" y1="10" y2="10" />
    </svg>
  )
}
