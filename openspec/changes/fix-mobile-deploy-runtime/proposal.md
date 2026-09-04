# fix-mobile-deploy-runtime

## Why

打包（`make package`）部署后，后端将移动端 H5 产物托管在 `/m/` 子路径，但移动端构建产物（vite 默认 `base: '/'`）引用的 JS/CSS 是**根路径绝对引用**（`/assets/*`）。访问 `/m/` 只会拿到 `index.html`，其静态资源请求落到服务器根路径，被桌面端 SPA 回退（返回 `dist/index.html` 文本）或直接 404 —— 页面白屏，**部署后访问不了移动端**。

同时 `deploy/` 目录未包含 `backend/static/`（logo/favicon/登录背景图），部署后四端品牌图全部 404。

根因是 `static-serving` / `multi-platform-packaging` 两个 spec 只验收了「`/m/` 能返回 index.html」「部署目录结构完整」，没有把「产物能在 `/m/` 子路径**完整渲染**」「品牌资源随包部署」作为契约，缺陷因此没有任何护栏拦截。

## What Changes

- **移动端构建按子路径发布**：`mobile/vite.config.js` 在 `production` 时设 `base: '/m/'`（`dev` 保持 `/`，不影响开发体验）；`mobile/src/router/index.js` 的 `createWebHistory` 使用 `base: '/m/'`，使移动端 history 路由（`/m/login`、`/m/mine`）与构建 base 一致。
- **后端 `/m/` 子路径 SPA 回退**：`backend/router/serve.go` 将 `/m/*filepath` 从 `r.Static` 改为「文件存在则服务文件，否则回退 `dist-mobile/index.html`」——与桌面端 NoRoute 回退对称，保证 `/m/` 子路由直达/刷新不 404。
- **部署包补齐品牌资源**：`scripts/package.sh` 组装 `deploy/` 时拷贝 `backend/static/` → `deploy/static/`，使 `config.yaml` 的 `logo`/`favicon`/登录背景引用可用。
- **spec 契约补齐**（护栏）：
  - `static-serving` 新增「移动端 H5 在 `/m/` 子路径必须完整可渲染（HTML + 静态资源 + 子路由回退）」；
  - `multi-platform-packaging` 新增「部署包必须包含 `static/`；移动端产物必须以 `/m/` 子路径可访问」。
- **验收覆盖**：给两个 spec 补上「打包产物 → 启动 → `/m/` 可完整渲染并登录、刷新子路由不 404」的场景。

## Capabilities

### New Capabilities

（无——移动端子路径发布不是新能力，而是对既有「静态托管」与「打包交付」契约的收紧。）

### Modified Capabilities

- `static-serving`: 移动端 H5 托管从「`/m/` 返回 index.html」收紧为「`/m/` 子路径下**完整可渲染**（HTML 与资源均可达）并提供移动端 SPA 子路径回退」。
- `multi-platform-packaging`: 部署包从「包含二进制 + dist + dist-mobile + config.yaml + 启动脚本」扩展为「还包含 `backend/static/` 品牌资源目录；且 dist-mobile 产物满足 `/m/` 子路径托管约定」。

## Impact

- `mobile/vite.config.js`、`mobile/src/router/index.js` —— 产物与路由 base 变为 `/m/`（仅影响生产构建，dev 不变）。
- `backend/router/serve.go` —— `/m/` 前缀的路由处理逻辑（文件服务 + SPA 回退）。
- `scripts/package.sh` —— `deploy/` 组装新增拷贝 `static/`。
- 部署产物（`deploy/` 目录与压缩包）新增 `static/` 目录；`/m/` 下静态资源 URL 形态变化。
- Spec 变更：`openspec/specs/static-serving/spec.md`、`openspec/specs/multi-platform-packaging/spec.md`。
