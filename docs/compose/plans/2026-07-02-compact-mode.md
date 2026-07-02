# 紧凑模式实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现侧边栏折叠功能，支持完整模式和仅图标模式切换

**Architecture:** 在 AppLayout 中添加 CompactModeContext，通过 CSS 类控制侧边栏宽度和内容区布局

**Tech Stack:** React 19, TypeScript, Tailwind CSS v4, localStorage

## Global Constraints

- 使用 React Context 管理紧凑模式状态
- 使用 localStorage 持久化状态
- 保持现有主题系统兼容
- 侧边栏宽度：完整模式 264px，图标模式 64px
- 切换动画：transition-width 300ms ease-out

---

## Phase 1: 侧边栏折叠

### Task 1: 创建 CompactModeContext

**Covers:** [S3, S6]

**Files:**
- Create: `frontend/src/lib/compact-mode.tsx`
- Modify: `frontend/src/App.tsx`

**Interfaces:**
- Consumes: None (new module)
- Produces: `CompactModeContext`, `useCompactMode` hook

- [ ] **Step 1: Create compact-mode.tsx**

```typescript
import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'

type CompactMode = 'full' | 'icons' | 'hidden'

interface CompactModeContextValue {
  mode: CompactMode
  setMode: (mode: CompactMode) => void
  toggle: () => void
}

const CompactModeContext = createContext<CompactModeContextValue>({
  mode: 'full',
  setMode: () => {},
  toggle: () => {},
})

export function CompactModeProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<CompactMode>(() => {
    const saved = localStorage.getItem('apihub-compact-mode')
    if (saved === 'full' || saved === 'icons' || saved === 'hidden') return saved
    return 'full'
  })

  useEffect(() => {
    localStorage.setItem('apihub-compact-mode', mode)
  }, [mode])

  const setMode = (m: CompactMode) => setModeState(m)
  const toggle = () => setModeState((prev) => {
    if (prev === 'full') return 'icons'
    if (prev === 'icons') return 'hidden'
    return 'full'
  })

  return (
    <CompactModeContext.Provider value={{ mode, setMode, toggle }}>
      {children}
    </CompactModeContext.Provider>
  )
}

export function useCompactMode() {
  return useContext(CompactModeContext)
}

export type { CompactMode }
```

- [ ] **Step 2: Run TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/compact-mode.tsx
git commit -m "feat: add CompactModeContext for sidebar collapse"
```

---

### Task 2: 修改 Sidebar 支持图标模式

**Covers:** [S3]

**Files:**
- Modify: `frontend/src/components/layout/Sidebar.tsx`

**Interfaces:**
- Consumes: `useCompactMode` from Task 1
- Produces: Sidebar with icons-only mode

- [ ] **Step 1: Update Sidebar.tsx**

```typescript
import { NavLink } from 'react-router-dom'
import { X, Sparkles, Moon, Sun, ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '../../lib/utils'
import { useTheme } from '../../lib/use-theme'
import { useCompactMode } from '../../lib/compact-mode'
import { navMain, navMore, navBottom } from '../../lib/nav'

interface SidebarProps {
  open: boolean
  onClose: () => void
}

export function Sidebar({ open, onClose }: SidebarProps) {
  const { theme, toggle } = useTheme()
  const { mode, setMode } = useCompactMode()
  const isCollapsed = mode === 'icons'

  function NavItem({ path, label, icon: Icon }: { path: string; label: string; icon: React.ComponentType<{ className?: string }> }) {
    return (
      <NavLink
        to={path}
        end={path === '/'}
        onClick={onClose}
        className={({ isActive }) => cn(
          'w-full flex items-center gap-3 rounded-xl text-sm font-medium transition-all duration-200 relative group',
          isCollapsed ? 'justify-center px-2 py-2.5' : 'px-3 py-2.5',
          isActive ? 'bg-white/10 text-white' : 'text-slate-400 hover:text-white hover:bg-white/5',
        )}
        title={isCollapsed ? label : undefined}
      >
        {({ isActive }) => (
          <>
            {isActive && !isCollapsed && <span className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-r-full bg-gradient-to-b from-indigo-400 to-violet-400" />}
            {isActive && isCollapsed && <span className="absolute left-1/2 -translate-x-1/2 bottom-0 w-5 h-[3px] rounded-t-full bg-gradient-to-r from-indigo-400 to-violet-400" />}
            <Icon className={cn('w-5 h-5 shrink-0', isActive ? 'text-indigo-300' : 'text-slate-500 group-hover:text-slate-300')} />
            {!isCollapsed && <span>{label}</span>}
          </>
        )}
      </NavLink>
    )
  }

  return (
    <>
      {open && <div className="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-30 lg:hidden" onClick={onClose} />}

      <aside className={cn(
        'fixed lg:static inset-y-0 left-0 z-40 flex flex-col sidebar-bg border-r border-white/[0.06]',
        'transition-all duration-300 ease-out',
        isCollapsed ? 'w-16' : 'w-64',
        open ? 'translate-x-0' : '-translate-x-full lg:translate-x-0',
      )}>
        <div className={cn(
          'h-16 flex items-center border-b border-white/[0.06] shrink-0',
          isCollapsed ? 'justify-center px-2' : 'gap-3 px-5',
        )}>
          <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center shadow-lg shadow-indigo-500/30 shrink-0">
            <Sparkles className="w-4 h-4 text-white" />
          </div>
          {!isCollapsed && (
            <div className="flex-1 min-w-0">
              <span className="font-bold text-white text-base tracking-tight">APIHub</span>
              <p className="text-[10px] text-slate-500 uppercase tracking-widest">Monitor</p>
            </div>
          )}
          <button
            type="button"
            className={cn(
              'text-slate-400 hover:text-white p-1 rounded-lg hover:bg-white/10 transition-colors',
              isCollapsed ? 'hidden' : 'lg:hidden',
            )}
            onClick={onClose}
            aria-label="关闭"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <nav className={cn('flex-1 overflow-y-auto', isCollapsed ? 'p-2' : 'p-3 space-y-6')}>
          {isCollapsed ? (
            <>
              <div className="space-y-1">{navMain.map((n) => <NavItem key={n.id} {...n} />)}</div>
              <div className="space-y-1">{navMore.map((n) => <NavItem key={n.id} {...n} />)}</div>
            </>
          ) : (
            <>
              <div>
                <p className="px-3 mb-2 text-[10px] font-semibold uppercase tracking-wider text-slate-600">数据</p>
                <div className="space-y-0.5">{navMain.map((n) => <NavItem key={n.id} {...n} />)}</div>
              </div>
              <div>
                <p className="px-3 mb-2 text-[10px] font-semibold uppercase tracking-wider text-slate-600">工具</p>
                <div className="space-y-0.5">{navMore.map((n) => <NavItem key={n.id} {...n} />)}</div>
              </div>
            </>
          )}
        </nav>

        <div className={cn('p-3 border-t border-white/[0.06] space-y-1', isCollapsed && 'p-2')}>
          {isCollapsed ? (
            <>
              {navBottom.map((n) => <NavItem key={n.id} {...n} />)}
              <button
                type="button"
                onClick={toggle}
                className="w-full flex items-center justify-center px-2 py-2.5 rounded-xl text-sm font-medium text-slate-400 hover:text-white hover:bg-white/5 transition-all duration-200"
                title={theme === 'dark' ? '亮色模式' : '暗色模式'}
              >
                {theme === 'dark' ? <Sun className="w-5 h-5 text-slate-500" /> : <Moon className="w-5 h-5 text-slate-500" />}
              </button>
            </>
          ) : (
            <>
              {navBottom.map((n) => <NavItem key={n.id} {...n} />)}
              <button type="button" onClick={toggle} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium text-slate-400 hover:text-white hover:bg-white/5 transition-all duration-200">
                {theme === 'dark' ? <Sun className="w-[18px] h-[18px] text-slate-500" /> : <Moon className="w-[18px] h-[18px] text-slate-500" />}
                <span>{theme === 'dark' ? '亮色模式' : '暗色模式'}</span>
              </button>
              <div className="px-3 pt-2"><p className="text-[10px] text-slate-600">APIHub v0.4</p></div>
            </>
          )}
        </div>

        {/* Collapse toggle button */}
        <button
          type="button"
          onClick={() => setMode(isCollapsed ? 'full' : 'icons')}
          className={cn(
            'absolute top-1/2 -translate-y-1/2 w-6 h-12 flex items-center justify-center',
            'bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white',
            'rounded-r-lg transition-all duration-200',
            isCollapsed ? 'right-0' : 'right-0',
          )}
          aria-label={isCollapsed ? '展开侧边栏' : '折叠侧边栏'}
        >
          {isCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
        </button>
      </aside>
    </>
  )
}
```

- [ ] **Step 2: Run TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/layout/Sidebar.tsx
git commit -m "feat: add icons-only mode to Sidebar"
```

---

### Task 3: 修改 AppLayout 集成 CompactModeProvider

**Covers:** [S3, S6]

**Files:**
- Modify: `frontend/src/App.tsx`

**Interfaces:**
- Consumes: `CompactModeProvider` from Task 1
- Produces: AppLayout with compact mode support

- [ ] **Step 1: Update App.tsx**

Add import and wrap AppLayout:

```typescript
import { CompactModeProvider } from './lib/compact-mode'

// In AppLayout function, wrap with CompactModeProvider:
function AppLayout({ onLogout, authEnabled }: { onLogout?: () => void; authEnabled: boolean }) {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const location = useLocation()

  return (
    <CompactModeProvider>
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
                    {/* ... routes ... */}
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
```

- [ ] **Step 2: Run TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Run build**

Run: `cd frontend && npm run build`
Expected: Build successful

- [ ] **Step 4: Commit**

```bash
git add frontend/src/App.tsx
git commit -m "feat: integrate CompactModeProvider into AppLayout"
```

---

### Task 4: 修改 TopBar 支持紧凑模式

**Covers:** [S3]

**Files:**
- Modify: `frontend/src/components/layout/TopBar.tsx`

**Interfaces:**
- Consumes: `useCompactMode` from Task 1
- Produces: TopBar with hamburger menu for hidden mode

- [ ] **Step 1: Update TopBar.tsx**

```typescript
import { useLocation, useParams } from 'react-router-dom'
import { Menu, LogOut } from 'lucide-react'
import { cn } from '../../lib/utils'
import { useCompactMode } from '../../lib/compact-mode'
import { allNav } from '../../lib/nav'

interface TopBarProps {
  authEnabled: boolean
  onLogout?: () => void
  onMenuOpen: () => void
}

export function TopBar({ authEnabled, onLogout, onMenuOpen }: TopBarProps) {
  const location = useLocation()
  const params = useParams()
  const { mode } = useCompactMode()

  const currentLabel = location.pathname.startsWith('/model/')
    ? (params.model ?? '')
    : allNav.find((n) => n.path === location.pathname)?.label ?? ''

  return (
    <header className="h-16 shrink-0 flex items-center gap-4 px-4 lg:px-8 border-b border-slate-200/60 dark:border-slate-700/60 bg-white/70 dark:bg-slate-900/70 backdrop-blur-xl">
      <button
        type="button"
        className={cn(
          'p-2 -ml-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400',
          mode === 'hidden' ? '' : 'lg:hidden',
        )}
        onClick={onMenuOpen}
        aria-label="菜单"
      >
        <Menu className="w-5 h-5" />
      </button>
      <div className="flex-1 min-w-0">
        <h1 className="text-lg font-semibold text-slate-900 dark:text-slate-100 truncate">{currentLabel}</h1>
      </div>
      {authEnabled && onLogout && (
        <button
          type="button"
          onClick={onLogout}
          className={cn(
            'flex items-center gap-2 px-3 py-2 rounded-xl text-sm transition',
            'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800',
          )}
        >
          <LogOut className="w-4 h-4" />
          <span className="hidden sm:inline">退出</span>
        </button>
      )}
    </header>
  )
}
```

- [ ] **Step 2: Run TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/layout/TopBar.tsx
git commit -m "feat: add hamburger menu for hidden mode in TopBar"
```

---

### Task 5: 添加紧凑模式切换快捷键

**Covers:** [S3]

**Files:**
- Modify: `frontend/src/App.tsx`

**Interfaces:**
- Consumes: `useCompactMode` from Task 1
- Produces: Keyboard shortcut for toggling compact mode

- [ ] **Step 1: Add keyboard shortcut hook**

Create `frontend/src/lib/use-compact-shortcut.ts`:

```typescript
import { useEffect } from 'react'
import { useCompactMode } from './compact-mode'

export function useCompactShortcut() {
  const { toggle } = useCompactMode()

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      // Ctrl/Cmd + Shift + C to toggle compact mode
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'C') {
        e.preventDefault()
        toggle()
      }
    }

    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [toggle])
}
```

- [ ] **Step 2: Use hook in AppLayout**

Add to AppLayout:

```typescript
import { useCompactShortcut } from './lib/use-compact-shortcut'

function AppLayout({ onLogout, authEnabled }: { onLogout?: () => void; authEnabled: boolean }) {
  useCompactShortcut()
  // ... rest of the component
}
```

- [ ] **Step 3: Run TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/use-compact-shortcut.ts frontend/src/App.tsx
git commit -m "feat: add Ctrl+Shift+C shortcut to toggle compact mode"
```

---

### Task 6: 添加设置页面紧凑模式选项

**Covers:** [S7]

**Files:**
- Modify: `frontend/src/pages/Settings.tsx`

**Interfaces:**
- Consumes: `useCompactMode` from Task 1
- Produces: Settings UI for compact mode

- [ ] **Step 1: Update Settings.tsx**

Add to GeneralSettings:

```typescript
import { useCompactMode } from '../lib/compact-mode'
import type { CompactMode } from '../lib/compact-mode'

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
```

- [ ] **Step 2: Run TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/Settings.tsx
git commit -m "feat: add compact mode setting in Settings page"
```

---

### Task 7: 添加紧凑模式 CSS 动画

**Covers:** [S3]

**Files:**
- Modify: `frontend/src/index.css`

**Interfaces:**
- Consumes: None
- Produces: CSS transitions for sidebar collapse

- [ ] **Step 1: Add CSS animations**

Add to `frontend/src/index.css`:

```css
/* Compact mode transitions */
.sidebar-transition {
  transition: width 300ms ease-out;
}

/* Tooltip for icons-only mode */
.tooltip {
  position: absolute;
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  margin-left: 8px;
  padding: 4px 8px;
  background: rgba(15, 23, 42, 0.9);
  color: white;
  font-size: 12px;
  border-radius: 6px;
  white-space: nowrap;
  pointer-events: none;
  opacity: 0;
  transition: opacity 200ms;
  z-index: 50;
}

.group:hover .tooltip {
  opacity: 1;
}
```

- [ ] **Step 2: Run build**

Run: `cd frontend && npm run build`
Expected: Build successful

- [ ] **Step 3: Commit**

```bash
git add frontend/src/index.css
git commit -m "feat: add CSS transitions for compact mode"
```

---

### Task 8: 验证和测试

**Covers:** [S10]

**Files:**
- None (verification only)

**Interfaces:**
- Consumes: All previous tasks
- Produces: Verified working feature

- [ ] **Step 1: Run all checks**

Run: `cd frontend && npm run lint && npm run test && npm run build`
Expected: All pass

- [ ] **Step 2: Manual testing checklist**

- [ ] 点击折叠按钮，侧边栏变为图标模式
- [ ] 图标模式下 hover 显示 tooltip
- [ ] 再次点击展开按钮，恢复完整模式
- [ ] 刷新页面后模式保持
- [ ] 设置页面可以切换模式
- [ ] Ctrl+Shift+C 快捷键可以切换模式
- [ ] 动画过渡平滑

- [ ] **Step 3: Final commit**

```bash
git add -A
git commit -m "feat: complete sidebar collapse feature"
```

---

## Phase 2: 系统托盘（未来实现）

> 以下为 Phase 2 的占位，需要 Wails 后端支持

### Task 9: 系统托盘集成（Wails）

**Covers:** [S4]

**Files:**
- Modify: `app_wails.go`
- Create: `frontend/src/lib/system-tray.ts`

**Interfaces:**
- Consumes: Wails runtime API
- Produces: System tray with cost/token display

---

## Phase 3: 浮动小窗口（未来实现）

> 以下为 Phase 3 的占位，需要 Wails 后端支持

### Task 10: 浮动小窗口（Wails）

**Covers:** [S5]

**Files:**
- Modify: `app_wails.go`
- Create: `frontend/src/pages/FloatingWidget.tsx`

**Interfaces:**
- Consumes: Wails window API
- Produces: Draggable floating widget
