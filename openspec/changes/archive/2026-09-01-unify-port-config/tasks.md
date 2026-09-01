# Tasks: unify-port-config

## 1. vite proxy 可配置化

- [x] 1.1 `frontend/vite.config.js` 用 `loadEnv` 读 `VITE_API_BASE`，proxy target 改为 `env.VITE_API_BASE || 'http://localhost:8080'`
- [x] 1.2 `mobile/vite.config.js` 同样改造
- [x] 1.3 新增 `frontend/.env.example`、`mobile/.env.example`（含 `VITE_API_BASE=http://localhost:8080` 注释）
- [x] 1.4 `.gitignore` 忽略 `.env`（保留 `.env.example`）

## 2. init.sh 支持端口

- [x] 2.1 `scripts/init.sh` 新增 `--port` 参数解析
- [x] 2.2 传 --port 时生成 `frontend/.env` 与 `mobile/.env`（`VITE_API_BASE=http://localhost:<port>`）
- [x] 2.3 更新 init.sh 头部注释与 README 说明

## 3. 文档与验证

- [x] 3.1 更新 `docs/配置体系.md`：说明端口改动的正确姿势（config.yaml + .env）
- [x] 3.2 前端/移动端 build 通过，dev 模式下 proxy 指向 .env 指定的端口
- [x] 3.3 验证缺省（无 .env）时回退 8080，行为与现状一致
