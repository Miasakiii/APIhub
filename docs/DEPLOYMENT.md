# APIHub 部署指南

> 当前状态：v0.2 仍建议以本地开发方式运行。Docker 能构建并启动后端，但 Go 服务尚未提供 `frontend/dist` 静态资源服务，所以还不是完整的单端口 UI 部署。

## 本地开发（推荐）

### 后端

```bash
cd /path/to/APIhub
go run ./cmd/apihub
```

默认行为：

- 监听 `:8080`
- 数据目录为 `~/.apihub`
- 自动生成 `$APIHUB_DATA_DIR/.master_key`
- `APIHUB_AUTH_ENABLED=false`，本地开发无需登录

访问健康检查：

```bash
curl http://localhost:8080/health
# {"auth_enabled":false,"status":"ok"}
```

### 前端

```bash
cd frontend
npm install
npm run dev
```

Windows PowerShell 如果因为 execution policy 拒绝 `npm.ps1`，使用：

```powershell
npm.cmd run dev
```

访问 `http://localhost:5173`。Vite 会代理 `/api` 到 `http://localhost:8080`。

### 开启认证（模拟生产）

```bash
export APIHUB_AUTH_ENABLED=true
export APIHUB_CORS_ORIGIN=http://localhost:5173
go run ./cmd/apihub
```

首次访问前端会显示注册页。注册第一个用户后，建议重启服务并设置：

```bash
export APIHUB_ALLOW_REGISTER=false
```

## Docker

```bash
docker-compose up --build
```

当前 compose 文件默认：

- `APIHUB_AUTH_ENABLED=true`
- `APIHUB_ALLOW_REGISTER=true`
- 数据持久化到 volume `apihub-data`

注意事项：

- 镜像会构建 `frontend/dist`，但后端尚未挂载静态文件；直接打开 `http://localhost:8080` 不会进入 React UI。
- 如果使用本地 Vite 前端连接 Docker 后端，请把 `APIHUB_CORS_ORIGIN` 改为 `http://localhost:5173`。
- 如果只在本机使用，建议把端口映射改为 `127.0.0.1:8080:8080`，避免无意暴露到局域网。
- 首次注册后建议设置 `APIHUB_ALLOW_REGISTER=false`。

## 数据源路径

| 源 | 路径 | 说明 |
|----|------|------|
| cc-switch | `%USERPROFILE%\.cc-switch\cc-switch.db`（Windows） | 可用 `APIHUB_CC_SWITCH_PATH` 覆盖 |
| JSONL | `~/.claude/projects/**/*.jsonl` | 自动扫描，增量 byte offset |
| Syncer | 当前依赖 provider id 或类型与 syncer 名匹配 | 如 `openrouter`、`one-api`；映射字段仍在 Roadmap 中 |

## 手动触发同步

无认证模式：

```bash
# cc-switch 增量
curl -X POST http://localhost:8080/api/v1/sync/ccswitch

# 指定 syncer；当前 provider_id 需要能匹配 syncer 名称或现有 Provider id
curl -X POST "http://localhost:8080/api/v1/sync/openrouter"
```

认证开启时需加 Header：

```bash
curl -X POST http://localhost:8080/api/v1/sync/ccswitch \
  -H "Authorization: Bearer <token>"
```

## 构建与验证

当前已知结果：

```bash
cd frontend
npm.cmd run build   # 通过，仍有 bundle size warning
npm.cmd run lint    # 当前失败，详见 ROADMAP.md

cd ..
go test ./...       # 当前失败，详见 ROADMAP.md
```

把以上命令恢复全绿后，再把 Docker 单端口 UI 部署作为稳定部署路径。
