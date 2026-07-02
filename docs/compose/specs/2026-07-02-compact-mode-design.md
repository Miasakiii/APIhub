# 紧凑模式设计文档

## [S1] 问题

当前 APIHub 的侧边栏固定宽度（264px），在小屏幕或需要专注工作时占用过多空间。用户希望能够：
1. 折叠侧边栏，只显示图标
2. 在系统托盘显示关键指标
3. 有浮动小窗口显示核心数据

## [S2] 解决方案

采用状态驱动 + CSS 类切换的方案，在 `AppLayout` 中添加紧凑模式状态，通过 CSS 类控制布局。

**核心组件**：
1. **CompactModeProvider** - 管理紧凑模式状态（localStorage 持久化）
2. **Sidebar** - 修改支持仅图标模式
3. **SystemTray** - 系统托盘集成（Wails）
4. **FloatingWidget** - 浮动小窗口组件

## [S3] 侧边栏折叠设计

**状态**：
```typescript
type CompactMode = 'full' | 'icons' | 'hidden'
```

**布局变化**：
- `full` 模式：侧边栏宽度 264px（当前）
- `icons` 模式：侧边栏宽度 64px，只显示图标
- `hidden` 模式：侧边栏完全隐藏（通过汉堡菜单唤出）

**交互**：
- 侧边栏顶部添加折叠按钮（`ChevronLeft` / `ChevronRight`）
- `icons` 模式下 hover 图标显示 tooltip（文字标签）
- 切换动画：`transition-width 300ms ease-out`

**图标模式下的导航项**：
```tsx
// 仅图标，无文字
<Icon className="w-5 h-5" />
// hover 时显示 tooltip
<div className="tooltip">{label}</div>
```

## [S4] 系统托盘设计

**显示内容**：
- 今日费用（如 `$12.34`）
- 总 Token 数（如 `1.2M`）
- 有告警时图标变色（红色）

**交互**：
- 点击托盘图标：显示/隐藏主窗口
- 右键菜单：显示/隐藏、退出

**实现**：
- 使用 Wails 的系统托盘 API
- 定时更新（每 30 秒）

## [S5] 浮动小窗口设计

**尺寸**：200x120px

**内容**：
- 今日费用（大字体）
- 总 Token 数
- 最近更新时间

**交互**：
- 可拖拽
- 可关闭（隐藏到托盘）
- 点击展开主窗口

**实现**：
- 使用 Wails 的无边框窗口 API
- 独立窗口，置于顶层

## [S6] 状态管理

**存储**：
```typescript
// localStorage
apihub-compact-mode: 'full' | 'icons' | 'hidden'
apihub-floating-visible: boolean
apihub-tray-enabled: boolean
```

**Context**：
```typescript
interface CompactModeContext {
  mode: CompactMode
  setMode: (mode: CompactMode) => void
  toggle: () => void
}
```

## [S7] 设置页面扩展

在设置页面添加"显示"标签页：
- 紧凑模式：下拉选择（完整/仅图标/隐藏）
- 浮动小窗口：开关
- 系统托盘：开关

## [S8] 实现优先级

1. **Phase 1**: 侧边栏折叠（仅图标模式）
2. **Phase 2**: 系统托盘集成
3. **Phase 3**: 浮动小窗口

## [S9] 技术约束

- 使用 React Context 管理状态
- 使用 localStorage 持久化
- 使用 Wails API 实现系统托盘和浮动窗口
- 保持现有主题系统兼容

## [S10] 测试计划

1. 侧边栏折叠/展开动画
2. 图标模式下的 tooltip 显示
3. 系统托盘图标更新
4. 浮动小窗口拖拽和关闭
5. 设置页面切换生效
6. 状态持久化（刷新后保持）
