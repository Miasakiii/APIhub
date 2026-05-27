# APIHub 安全说明

## 威胁模型

APIHub 设计为**个人本地或受信局域网**工具，不是面向公网的多租户 SaaS。它会保存可调用上游模型服务的 API Key，因此默认部署边界必须保守。

## 当前安全默认值

| 项 | 当前行为 | 风险/建议 |
|----|----------|-----------|
| 认证 | `APIHUB_AUTH_ENABLED=false` 时所有业务 API 无需登录 | 仅用于可信本机开发；任何 LAN/VPS 暴露都应开启认证 |
| 监听地址 | 认证关闭时默认仅监听 `127.0.0.1`；认证开启时监听所有网卡。可通过 `APIHUB_BIND_ADDR` 覆盖。 | 本机开发安全；LAN 部署需手动设置 bind addr 并开启认证 |
| 注册 | `APIHUB_ALLOW_REGISTER=true` 默认允许注册 | 首次注册后建议改为 `false` |
| Docker compose | 默认开启认证，但仍允许注册 | 首次注册后关闭注册；如配合 Vite 前端使用，更新 CORS origin |

## 认证模式

| 模式 | `APIHUB_AUTH_ENABLED` | 适用场景 |
|------|------------------------|----------|
| 本地模式 | `false`（默认） | 可信本机开发，无登录 |
| 加固模式 | `true` | Docker、局域网、小 VPS |

加固模式下：

- 所有业务 API 需 `Authorization: Bearer <JWT>`
- JWT 使用 master key 派生的 `JWTKey`，HS256 签名，默认 7 天过期
- `/auth/register` 在已有用户且 `APIHUB_ALLOW_REGISTER=false` 时关闭

## 敏感端点

| 端点 | 风险 | 建议 |
|------|------|------|
| `GET /api/v1/keys/:id/decrypt` | 返回明文 API Key | 仅在可信本机或认证开启时使用；不要暴露到公网 |
| `POST /api/v1/playground/*` | 使用存储的 Key 调上游 API | 同上；这会产生真实上游请求和费用 |

## 密钥存储

当前实现：

- Master key：`$APIHUB_DATA_DIR/.master_key`，hex 编码 32 字节，首次启动自动生成
- API Key：AES-256-GCM 加密后存入 `api_keys.key_encrypted`
- JWT signing key：由 master key 通过 HKDF 派生

运维要求：

- 备份 `apihub.db` 时必须同时备份 `.master_key`
- 丢失 `.master_key` 后，数据库中已加密的 API Key 无法恢复
- 不要将 `.master_key`、`apihub.db`、`*.db-wal`、`*.db-shm` 提交到版本库或镜像
- 限制数据目录权限，仅允许当前运行用户读取

注意：`APIHUB_MASTER_KEY` 环境变量加载和 Docker 非交互强制密钥属于架构设计目标，当前代码尚未实现。

## CORS

- 默认 `APIHUB_CORS_ORIGIN=http://localhost:5173`
- 认证开启时不要使用 `Access-Control-Allow-Origin: *`
- 如果 Docker 后端配合 Vite 前端开发，需要将 `APIHUB_CORS_ORIGIN` 设为 `http://localhost:5173`
- 如果未来由 Go 后端直接服务前端静态资源，再把 origin 调整到实际访问地址

## 部署清单

- [ ] 非本机访问时设置 `APIHUB_AUTH_ENABLED=true`
- [ ] 设置强密码并完成首次注册后关闭 `APIHUB_ALLOW_REGISTER`
- [ ] 限制端口暴露范围；本机使用时优先绑定或映射到 `127.0.0.1`
- [ ] 使用 HTTPS 反向代理（Caddy / Nginx）后再考虑公网访问
- [ ] 同时备份 `apihub.db` 和 `.master_key`
- [ ] 不要公开明文 Key 解密端点和 Playground
