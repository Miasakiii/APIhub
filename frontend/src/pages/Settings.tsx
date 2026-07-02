import { useState } from 'react'
import { Settings as SettingsIcon, Database, Sun, Moon, Languages, DollarSign, Download, Scan, CheckCircle, XCircle, AlertCircle, Monitor } from 'lucide-react'
import { cn } from '../lib/utils'
import { api } from '../api'
import type { ScanFinding, ScanImportResult } from '../api'
import { useTheme } from '../lib/use-theme'
import { useCompactMode } from '../lib/compact-mode'
import type { CompactMode } from '../lib/compact-mode'
import { Card, Button, Select } from '../components/ui'

export function Settings() {
  const { theme, setTheme } = useTheme()
  const { mode: compactMode, setMode: setCompactMode } = useCompactMode()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">设置</h1>
        <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">管理你的应用偏好</p>
      </div>

      <Card padding={false}>
        <div className="p-6 lg:p-8">
          <div className="space-y-8">
            <GeneralSettings theme={theme} setTheme={setTheme} compactMode={compactMode} setCompactMode={setCompactMode} />
            <DataSettings />
          </div>
        </div>
      </Card>
    </div>
  )
}

function GeneralSettings({ theme, setTheme, compactMode, setCompactMode }: {
  theme: 'light' | 'dark'; setTheme: (t: 'light' | 'dark') => void;
  compactMode: CompactMode; setCompactMode: (m: CompactMode) => void;
}) {
  return (
    <div className="space-y-4">
      <h2 className="text-lg font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
        <SettingsIcon className="w-5 h-5" /> 通用设置
      </h2>
      <div className="space-y-0 divide-y divide-slate-100 dark:divide-slate-800">
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
    <div className="space-y-4">
      <h2 className="text-lg font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
        <Database className="w-5 h-5" /> 数据管理
      </h2>
      <div className="space-y-0 divide-y divide-slate-100 dark:divide-slate-800">
        <div className="py-4">
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
      </div>
    </div>
  )
}

function SettingRow({ icon: Icon, label, description, children, danger }: {
  icon?: React.ComponentType<{ className?: string }>; label: string; description: string; children: React.ReactNode; danger?: boolean
}) {
  return (
    <div className="flex items-center justify-between py-4">
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
