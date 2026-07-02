# APIHub

> 大模型 API 用量监控工具 - 专注、简洁、实时

借鉴 token-monitor 的设计哲学，专注于一件事：**监控你的 AI 工具消耗了多少 Token 和费用**。

[中文](#apihub-cn)

---

## What is APIHub?

APIHub is a local-first dashboard for monitoring your LLM API usage and costs across multiple providers (OpenAI, Anthropic, OpenRouter, etc.).

It aggregates data from three sources:
- **Source A**: cc-switch proxy request logs (if you use cc-switch)
- **Source B**: Claude Code JSONL usage files (from `~/.claude/projects/`)
- **Source C**: Built-in syncers that fetch usage directly from provider APIs

## Current Status

> Snapshot: **v0.13 stable**, updated 2026-06-19.

APIHub has a real backend, database schema, auth flow, scheduler, sync paths, and a polished React dashboard with dark mode support. Docker serves the frontend in a single container — `docker-compose up --build` is sufficient to run the full UI.

Two ways to run:

- **Docker** (recommended): `docker-compose up --build` → open `http://localhost:8080`
- **Dev mode**: Go backend on `http://localhost:8080` + Vite frontend on `http://localhost:5173`

Latest local verification:

| Check | Result |
|---|---|
| `npm.cmd run build` in `frontend/` | ✅ Pass (code-split, no chunk warning) |
| `npm.cmd run lint` in `frontend/` | ✅ Pass (0 errors, 0 warnings) |
| `npm.cmd run test` in `frontend/` | ✅ Pass (42 tests) |
| `go test ./internal/...` | ✅ Pass |
| `go vet ./internal/...` | ✅ Pass |

Test coverage (v0.14):

| Package | Coverage |
|---|---|
| util | 100% |
| scanner | 83.1% |
| crypto | 76.0% |
| db | 67.8% |
| repository | 41.1% |
| service | 40.0% |
| api | 28.8% |
| alert | 22.0% |
| ws | 17.2% |
| aggregator | 11.1% |

v0.14 adds desktop settings panel, notification integration, and comprehensive unit tests. See [ROADMAP.md](ROADMAP.md) for P3+ plans.

## Features

### 核心功能

- **实时用量扫描** - 自动扫描 Claude Code、Codex、OpenCode 等工具的本地日志
- **成本统计** - 按模型、按天汇总费用，内置 35+ 模型定价
- **简单仪表盘** - 今日费用、总 Token 数、模型排行、最近用量
- **历史记录** - 按天/周/月查看用量趋势

### 界面特性

- **暗色模式** - 完整 dark/light 主题切换，localStorage 持久化
- **紧凑模式** - 侧边栏可折叠为图标模式，Ctrl+Shift+C 快捷切换
- **实时更新** - WebSocket 推送，秒级刷新
- **简洁导航** - 只保留核心页面，一目了然

### 数据源

- **Claude Code JSONL** - 增量扫描 `~/.claude/projects/` 目录
- **本地配置扫描** - 自动检测 API Key（Claude Code、DeepSeek、Kimi Code、Codex）
- **MCP 配置扫描** - 识别 MCP 服务器配置中的 API Key

### 技术特性

- **SQLite-backed** - 本地优先，无需外部数据库
- **Docker 支持** - 单容器部署
- **Wails 桌面端** - 系统托盘模式，后台持续监控

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26, Gin, SQLite |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS v4, Recharts |
| UI | 简洁的组件库，暗色模式，紧凑模式 |
| 实时 | WebSocket 推送 |

## Quick Start

### Prerequisites

- Go 1.26+
- Node.js 18+

### Run Backend

```bash
cd /f/su/APIhub
go run ./cmd/apihub
```

Server starts at `http://localhost:8080`.

### Run Frontend (Dev)

```bash
cd /f/su/APIhub/frontend
npm install
npm run dev
```

Vite dev server starts at `http://localhost:5173` (proxies API to `http://localhost:8080`).

### Build

```bash
# Backend
cd /f/su/APIhub
go build -o apihub ./cmd/apihub

# Frontend
cd /f/su/APIhub/frontend
npm run build
```

### Docker

Docker builds the backend and serves the frontend assets in a single container. The Go server now mounts `frontend/dist` and handles SPA fallback for client-side routing.

```bash
docker-compose up --build
```

Then open `http://localhost:8080` — no separate Vite process needed.

```bash
# Or use Docker directly
docker build -t apihub .
docker run -p 8080:8080 -v apihub-data:/app/data apihub
```

Docker Compose enables auth by default. See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).

### Authentication (optional)

Local development runs **without** login. To enable JWT auth:

```bash
export APIHUB_AUTH_ENABLED=true
export APIHUB_CORS_ORIGIN=http://localhost:5173
go run ./cmd/apihub
```

Open the frontend, register the first user, then optionally set `APIHUB_ALLOW_REGISTER=false`. See [docs/SECURITY.md](docs/SECURITY.md).

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APIHUB_PORT` | `8080` | HTTP listen port |
| `APIHUB_DATA_DIR` | `~/.apihub` | Database and master key directory |
| `APIHUB_AUTH_ENABLED` | `false` | Require JWT for API access |
| `APIHUB_JWT_EXPIRY` | `168h` | JWT lifetime |
| `APIHUB_ALLOW_REGISTER` | `true` | Allow new user registration |
| `APIHUB_CORS_ORIGIN` | `http://localhost:5173` | Allowed CORS origin |
| `APIHUB_BIND_ADDR` | `127.0.0.1` (no auth) / `0.0.0.0` (auth) | Listen address (host) |
| `APIHUB_SYNC_INTERVAL` | `5m` | cc-switch + JSONL sync interval |
| `APIHUB_SYNCER_INTERVAL` | `30m` | Source C syncer interval |
| `APIHUB_CC_SWITCH_PATH` | (auto) | Override cc-switch.db path |

Copy [.env.example](.env.example) for a full template.

## Security

- **Local mode** (`APIHUB_AUTH_ENABLED=false`): intended for trusted local development only. The server defaults to listening on `127.0.0.1:8080`; enable auth and set `APIHUB_BIND_ADDR=0.0.0.0` before any LAN/VPS exposure.
- **Secured mode**: enable auth for Docker/LAN; never expose `GET /keys/:id/decrypt` to the public internet without auth.
- **Master key**: current implementation auto-generates `$APIHUB_DATA_DIR/.master_key`. Back up this file together with `apihub.db`; losing it makes stored API keys undecryptable.

Details: [docs/SECURITY.md](docs/SECURITY.md) · Roadmap: [ROADMAP.md](ROADMAP.md)

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Dashboard  │────▶│   React +   │────▶│  Go + Gin   │
│  (简化的)   │     │  Recharts   │     │  REST API   │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                          ┌────────────────────┼────────────────────┐
                          ▼                    ▼                    ▼
                   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
                   │ Scanner     │    │ JSONL       │    │ Aggregator  │
                   │ (自动检测)  │    │ (增量扫描)  │    │ (用量聚合)  │
                   └─────────────┘    └─────────────┘    └─────────────┘
```

### Backend Structure

```
cmd/apihub/              # Entry point
internal/
  api/                   # HTTP handlers (route registration + request/response)
  repository/            # Data access layer (SQL queries, scanning)
  service/               # Business logic layer (validation, orchestration)
  model/                 # Domain models (Provider, APIKey, UsageRecord, etc.)
  aggregator/            # Single-goroutine daily_stats updater
  alert/                 # Alert engine (balance_low, key_expired, abnormal_frequency)
  crypto/                # HKDF key derivation + AES-256-GCM
  db/                    # SQLite + WAL + schema migrations
  scanner/               # Local config scanner (env vars + config files)
  scheduler/             # Background job scheduler
  sync/                  # cc-switch sync logic
  syncer/                # Syncer interface + Manager
    providers/           # OpenRouter, OpenAI, Anthropic, one-api, new-api implementations
sources/
  ccswitch/              # cc-switch.db reader
  jsonl/                 # Incremental JSONL parser + sync
```

### Frontend Structure

```
frontend/src/
  App.tsx                # BrowserRouter + Routes with React.lazy code splitting
  api.ts                 # API client
  lib/
    theme.tsx            # ThemeProvider (dark/light toggle)
    use-theme.ts         # useTheme hook
    use-toast.ts         # useToast hook
    nav.ts               # Navigation items with route paths
    utils.ts             # cn(), formatUSD(), formatNum()
    auth.ts              # Token management
  components/
    layout/
      Sidebar.tsx        # Sidebar with NavLink navigation
      TopBar.tsx         # Top header bar (label from useLocation)
    ui/
      index.tsx          # Card, Button, Input, Badge, StatCard, etc.
      Modal.tsx          # Unified modal dialog
      Toast.tsx          # Toast notification provider
      Tabs.tsx           # Tab switcher
  pages/                 # Route-level lazy-loaded pages
    Dashboard.tsx        # / — Overview with charts
    ModelDetail.tsx      # /model/:model — Model cost & usage detail
    Providers.tsx        # /providers — Provider management
    Keys.tsx             # /keys — API key management
    UsageLog.tsx         # /usage — Paginated usage table
    Alerts.tsx           # /alerts — Alert rules + history
    Subscriptions.tsx    # /subscriptions — Subscription tracking
    Frequency.tsx        # /frequency — Hourly heatmap
    Playground.tsx       # /playground — API testing
    Sessions.tsx         # /sessions — Session analysis with hourly chart
    Settings.tsx         # /settings — App settings
    Login.tsx            # Auth login/register (not routed)
```

## API Endpoints (v1)

### Auth
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login, returns token |
| GET | `/api/v1/auth/me` | Current user info |

### Providers
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/providers` | List all providers |
| POST | `/api/v1/providers` | Create provider |
| GET | `/api/v1/providers/:id` | Provider detail (keys, usage, stats) |
| DELETE | `/api/v1/providers/:id` | Delete provider |

### Keys
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/keys` | List keys |
| POST | `/api/v1/keys` | Add key (AES encrypted) |
| DELETE | `/api/v1/keys/:id` | Delete key |
| POST | `/api/v1/keys/:id/revoke` | Revoke key |

### Usage & Stats
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/usage` | Usage records (filterable) |
| GET | `/api/v1/usage/summary` | Aggregated summary |
| GET | `/api/v1/stats/daily` | Daily stats |
| GET | `/api/v1/stats/cost-trend` | Cost trend (30 days) |
| GET | `/api/v1/stats/model-breakdown` | Model cost breakdown |

### Sync
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/sync/status` | Sync status |
| POST | `/api/v1/sync/:provider_id` | Trigger manual sync |
| GET | `/api/v1/syncers` | List available syncers |

### Alerts
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/alerts` | List alert rules |
| POST | `/api/v1/alerts` | Create alert rule |
| PUT | `/api/v1/alerts/:id` | Update alert rule |
| DELETE | `/api/v1/alerts/:id` | Delete alert rule |
| GET | `/api/v1/alerts/history` | Alert history |

### Subscriptions
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/subscriptions` | List subscriptions |
| POST | `/api/v1/subscriptions` | Create subscription |
| GET | `/api/v1/subscriptions/:id` | Subscription detail |
| PUT | `/api/v1/subscriptions/:id` | Update subscription |
| DELETE | `/api/v1/subscriptions/:id` | Delete subscription |

### Frequency & Export
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/frequency/hourly` | Hourly heatmap |
| GET | `/api/v1/frequency/peak-qps` | Peak QPS |
| GET | `/api/v1/frequency/today` | Today's hourly distribution |
| GET | `/api/v1/export/csv` | Export usage as CSV |

### Sessions
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/sessions` | Paginated session list (filterable by provider, model, source, date range) |
| GET | `/api/v1/sessions/stats` | Aggregate session statistics |
| GET | `/api/v1/sessions/buckets` | Activity bucket list |
| GET | `/api/v1/sessions/hourly` | 24-hour bucket distribution for a date |

### Scan
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/scan` | Scan local configs for API keys (returns masked results) |
| POST | `/api/v1/scan/import` | Import selected findings by index (re-scans internally) |

### Agents
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/agents` | List all agents |
| POST | `/api/v1/agents` | Create agent |
| GET | `/api/v1/agents/:id` | Get agent by ID |
| PUT | `/api/v1/agents/:id` | Update agent |
| DELETE | `/api/v1/agents/:id` | Delete agent |

### Playground
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/playground/chat` | Send test chat request |
| POST | `/api/v1/playground/validate` | Validate API Key |

### Webhooks
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/webhooks` | List webhooks |
| POST | `/api/v1/webhooks` | Create webhook |
| DELETE | `/api/v1/webhooks/:id` | Delete webhook |

---

## Contributing

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed design, [ANALYSIS.md](ANALYSIS.md) for codebase analysis, and [ROADMAP.md](ROADMAP.md) for release planning.

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/xxx`)
3. Commit your changes (`git commit -m 'feat: add xxx'`)
4. Push to the branch (`git push origin feature/xxx`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file.

---

## APIHub 中文简介

**大模型 API 用量监控工具** - 专注、简洁、实时

借鉴 token-monitor 的设计哲学，专注于一件事：**监控你的 AI 工具消耗了多少 Token 和费用**。

### 核心功能

- **实时用量扫描** - 自动扫描 Claude Code、Codex 等工具的本地日志
- **成本统计** - 按模型、按天汇总费用
- **简单仪表盘** - 今日费用、总 Token 数、模型排行
- **暗色模式** - 完整 dark/light 主题切换
- **紧凑模式** - 侧边栏可折叠为图标模式

### 快速开始

```bash
# 后端
cd /f/su/APIhub
go run ./cmd/apihub

# 前端
cd /f/su/APIhub/frontend
npm install
npm run dev
```

### 技术栈

| 层级 | 技术 |
|---|---|
| 后端 | Go 1.26, Gin, SQLite |
| 前端 | React 19, TypeScript, Vite, Tailwind CSS v4, Recharts |
| 实时 | WebSocket 推送 |

### 许可证

MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。
