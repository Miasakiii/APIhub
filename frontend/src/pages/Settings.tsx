import { useState } from 'react'
import { Settings as SettingsIcon, User, Shield, Database, Bell, Sun, Moon, Languages, DollarSign, Trash2, Download } from 'lucide-react'
import { cn } from '../lib/utils'
import { api } from '../api'
import { useTheme } from '../lib/use-theme'
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
          </div>
          <div className="flex-1 p-6 lg:p-8">
            {activeTab === 'general' && <GeneralSettings />}
            {activeTab === 'account' && <AccountSettings onLogout={onLogout} />}
            {activeTab === 'security' && <SecuritySettings />}
            {activeTab === 'notifications' && <NotificationSettings />}
            {activeTab === 'data' && <DataSettings />}
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
  return (
    <div className="space-y-6">
      <h2 className="text-lg font-bold text-slate-900 dark:text-slate-100">数据管理</h2>
      <div className="space-y-4">
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
