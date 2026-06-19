# APIHub

> Personal API monitoring dashboard. Track costs, tokens, and usage across multiple LLM providers.

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
| `go test ./internal/...` | ✅ Pass |
| `go vet ./internal/...` | ✅ Pass |

Test coverage (v0.14):

| Package | Coverage |
|---|---|
| scanner | 83.1% |
| crypto | 76.0% |
| db | 67.8% |
| repository | 41.1% |
| service | 40.0% |
| alert | 22.0% |
| ws | 17.2% |
| util | 100% |

v0.14 adds desktop settings panel, notification integration, and comprehensive unit tests. See [ROADMAP.md](ROADMAP.md) for P3+ plans.

## Features

- **Client-side routing**: react-router with URL-based navigation, browser history, and deep linking
- **Code splitting**: route-level lazy loading — each page is a separate chunk, only loaded when visited
- **Dark mode**: full dark/light theme toggle with localStorage persistence
- **Polished UI**: Modal dialogs, Toast notifications, Tabs, loading skeletons, page transitions
- **Multi-provider API key management** with AES-256-GCM encryption
- **Three data sources**: cc-switch proxy logs / Claude Code JSONL / Provider API syncers
- **Cost aggregation** with daily statistics (accumulative sync, not overwrite)
- **Incremental JSONL sync** with byte-offset tracking and file rotation detection
- **Source C Syncers**: OpenRouter / OpenAI / Anthropic / one-api / new-api balance & usage hooks
- **Provider-syncer mapping**: configurable `syncer` field per provider
- **Alert system**: balance_low, key_expired, abnormal_frequency (5-min background checker)
- **Subscription management**: CRUD + expiry calendar
- **Frequency view**: hourly heatmap, peak QPS, daily distribution
- **Usage list pagination**: page/page_size support with frontend controls
- **CSV export**: usage records export
- **Optional authentication**: JWT (HS256) when `APIHUB_AUTH_ENABLED=true`; local dev defaults to open access
- **Security defaults**: binds to `127.0.0.1` when auth disabled; configurable via `APIHUB_BIND_ADDR`
- **Master key backup prompt**: first-run warning with file path
- **Playground**: multi-protocol API testing (OpenAI Chat Completions / Anthropic Messages)
- **Model detail page**: click any model in Dashboard ranking → full detail page with cost trend, token usage charts, and paginated records
- **Webhook 通知**: 告警触发时自动发送 Webhook 通知
- **Docker**: single-container deployment with static frontend serving + SPA fallback
- **SQLite-backed**, local-first, low external dependency surface
- **Model pricing table**: 35+ built-in model prices with auto cost backfill when source reports $0
- **Incremental DB migrations**: versioned schema via `PRAGMA user_version`, auto-upgrade on startup
- **Three-layer aggregation**: sessions (30-min window) → hourly buckets → daily rollups
- **Session analysis page**: hourly activity chart, session list with filtering and pagination
- **Local config auto-scan**: detects API keys from Claude Code, DeepSeek, Kimi Code, Codex configs and environment variables
- **Subscription auto-detection**: syncer FetchBalance automatically creates/updates subscription records
- **Subscription expiry alerts**: alerts when subscriptions are expiring within 7 days
- **Agent dimension tracking**: per-agent cost tracking with agents table and agent_id across all usage tables
- **WebSocket real-time updates**: live usage updates, alert toast notifications, sync progress broadcasting

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26, Gin, `database/sql` + `modernc.org/sqlite` |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS v4, Recharts, React Router |
| UI | Custom component library with dark mode, Toast system, Modal dialogs |
| Crypto | HKDF-SHA256 key derivation, AES-256-GCM encryption |

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
│  (Pages)    │     │  Recharts   │     │  REST API   │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                          ┌────────────────────┼────────────────────┐
                          ▼                    ▼                    ▼
                   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
                   │ Source A    │    │ Source B    │    │ Source C    │
                   │ cc-switch   │    │ JSONL       │    │ Syncers     │
                   │ (proxy logs)│    │ (incremental)│    │ (OpenRouter)│
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

APIHub 是一个本地优先的个人 API 用量监控仪表盘，用于追踪多个 LLM 服务商（OpenAI、Anthropic、OpenRouter 等）的 API 用量与费用。

### 当前状态

> 快照：**v0.12 stable**，更新于 2026-06-19。

v0.12 完成 WebSocket 实时推送：后端 Hub 模式连接管理，告警/同步/用量事件实时广播，前端自动重连与 Toast 通知集成。

v0.11 完成 Agent 维度追踪：新增 agents 表，所有用量表添加 agent_id 字段，前端新增 Agent 管理页面。

v0.10 完成订阅自动检测：syncer 的 FetchBalance 成功后自动创建/更新订阅记录，告警引擎实现订阅到期检测，前端显示自动标记。

v0.9 完成本地配置自动扫描：启动时检测 Claude Code、DeepSeek、Kimi Code、Codex 等工具的 API Key，前端提供一键导入 UI。

v0.8 完成三层用量聚合：usage_sessions（会话粒度，30 分钟窗口）+ usage_activity_buckets（小时粒度）+ daily_stats（天粒度），前端新增会话分析页面。

推荐运行方式：

```bash
# Docker（推荐）
docker-compose up --build
# 打开 http://localhost:8080

# 或开发模式
go run ./cmd/apihub
cd frontend && npm.cmd run dev
```

验证状态：前端 build ✅ | 前端 lint ✅ | Go 测试 ✅，详见 [ROADMAP.md](ROADMAP.md)。

### 数据来源
- **Source A**：cc-switch 代理请求日志（如使用 cc-switch）
- **Source B**：Claude Code JSONL 用量文件（`~/.claude/projects/`）
- **Source C**：内置同步器（OpenRouter / OpenAI / Anthropic / one-api / new-api）

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
| 后端 | Go 1.26, Gin, database/sql + modernc.org/sqlite |
| 前端 | React 19, TypeScript, Vite, Tailwind CSS v4, Recharts, React Router |
| UI | 自研组件库，支持暗色主题、Toast 通知、Modal 对话框 |
| 加密 | HKDF-SHA256 密钥派生, AES-256-GCM 加密 |

### 许可证

MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。
