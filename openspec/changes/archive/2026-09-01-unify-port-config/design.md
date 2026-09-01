# Design: unify-port-config

## Context

后端端口 `8080` 三处硬编码：`backend/config/config.go:89`（默认值）、`frontend/vite.config.js:28/32`（dev proxy target）、`mobile/vite.config.js:23/27`（dev proxy target）。用户在 config.yaml 改端口后，dev 模式 vite proxy 仍指向死端口 8080，静默连不上。

生产部署无此问题（后端直接托管前端产物，不走 vite proxy），本 change 聚焦 **dev 开发体验**：让前端/移动端 dev proxy 的 target 可配置。

## Goals / Non-Goals

**Goals:**

1. 前端/移动端 vite proxy 的 target 从硬编码改为从 `.env` 读取（`VITE_API_BASE`），缺省回退 `http://localhost:8080`。
2. 提供 `.env.example` 模板，说明 `VITE_API_BASE` 用途。
3. `init.sh` 支持 `--port` 参数，初始化时生成对应 `.env`。

**Non-Goals:**

- 不改生产部署路径（打包产物由后端托管，无需 vite proxy）。
- 不做「config.yaml 端口自动同步到 .env」的运行时联动（分属不同运行阶段，靠 init.sh 与文档约定）。
- 不强制用户用 .env（缺省回退 8080，行为与现状一致）。

## Decisions

### D1：vite proxy target 用 `loadEnv` 读取 `VITE_API_BASE`

vite.config.js 里 `import { loadEnv } from 'vite'`，`defineConfig(({ mode }) => { const env = loadEnv(mode, ...); ... })`，proxy target 用 `env.VITE_API_BASE || 'http://localhost:8080'`。

**Why：** `loadEnv` 是 Vite 原生，零依赖；`.env` 是 Vite 标准约定，开发者熟悉。缺省回退保持向后兼容。

**Alternatives considered：**
- 硬编码 + init.sh 文本替换端口 → 拒绝，只覆盖初始化时，运行时改端口仍失效。
- 从后端 config.yaml 动态读端口 → 拒绝，vite.config.js 是 Node 环境，读后端 YAML 需引入解析依赖，且耦合后端目录。

### D2：变量名 `VITE_API_BASE` 承载完整后端地址（含端口）

用 `VITE_API_BASE=http://localhost:8080` 承载「协议 + 主机 + 端口」，而非单独 `VITE_PORT`。

**Why：** proxy target 本质是完整 URL；承载完整地址可同时覆盖「后端跑在别的机器/域名」的场景，比只改端口更通用。

### D3：init.sh `--port` 生成 `.env` 而非替换 vite.config

`init.sh --port <端口>` 时生成 `frontend/.env` 与 `mobile/.env`，内容 `VITE_API_BASE=http://localhost:<端口>`；不直接改 vite.config.js。

**Why：** .env 是运行时配置（不进版本库），vite.config.js 是代码（进版本库）。端口是运行时变量，理应在 .env。

### D4：`.env.example` 提交，`.env` 进 .gitignore

提交 `frontend/.env.example`、`mobile/.env.example`；`.gitignore` 忽略 `.env`。

**Why：** example 提供模板与默认值；实际 `.env` 可能含环境相关配置，不进库。

## Risks / Trade-offs

- [`.env` 被忽略后，开发者 clone 下来没有 .env，需手动复制 example] → `.env.example` 提供 `VITE_API_BASE=http://localhost:8080` 缺省说明，且 loadEnv 缺省回退 8080，不复制也能跑。
- [`loadEnv` 在 vite.config.js 中需要正确的 envDir 与 mode] → 用默认 envDir（项目根）与 `mode`（dev/build），`VITE_` 前缀变量自动暴露。
- [`VITE_API_BASE` 若写成 `/api` 之类的相对路径会破坏 proxy] → 文档明确要求完整 URL（含 `http://` + 端口）。

## Open Questions

- 是否也覆盖 `VITE_API_BASE` 在 build 阶段的语义？生产部署不经过 vite proxy，`VITE_API_BASE` 仅影响 dev proxy，build 时无影响。暂不处理。
