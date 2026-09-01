# Proposal: unify-port-config

## Why

后端端口 `8080` 散落在三处硬编码：`backend/config/config.go:89`（默认值）、`frontend/vite.config.js:28/32`（dev proxy target）、`mobile/vite.config.js:23/27`（dev proxy target）。外部使用者在 `config.yaml` 改了端口（或 `init.sh` 未来支持改端口）后，dev 模式下的 vite proxy 仍指向死端口 8080——没有报错，就是静默连不上后端，对新人极其隐蔽。本 change 让端口有单一真相，改一处即可三端联动。

## What Changes

- 端口真源收敛：以 `backend/config/config.yaml` 的 `server.port`（或环境变量 `BASE_BACKEND_SERVER_PORT`）为唯一真相。
- 前端/移动端 vite proxy 的 target 从硬编码改为可配置：从 `.env` 读取后端地址（`VITE_API_BASE` 之类），缺省回退 `http://localhost:8080`。
- 新增 `frontend/.env.example`、`mobile/.env.example` 模板，说明 `VITE_API_BASE` 用途。
- `scripts/init.sh` 支持 `--port` 参数，把默认端口一并替换（若采用替换方案而非 env 方案）。
- 更新 `docs/配置体系.md`：说明端口改动的正确姿势（改 config.yaml + 改 .env）。

## Capabilities

### New Capabilities

- `unified-port-config`: 端口单一真相，后端端口改动自动联动前端/移动端 dev proxy，消除三处硬编码漂移。

### Modified Capabilities

（无。）

## Impact

- 修改文件：`frontend/vite.config.js`、`mobile/vite.config.js`、新增 `frontend/.env.example`、`mobile/.env.example`、`scripts/init.sh`（可能）、`docs/配置体系.md`。
- 无新第三方依赖（`loadEnv` 是 Vite 原生）。
- 无破坏：`.env` 缺省时回退 8080，行为与现状一致。
