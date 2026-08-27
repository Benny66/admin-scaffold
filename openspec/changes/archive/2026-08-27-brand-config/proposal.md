# Proposal: brand-config

## Why

基座目前的品牌配置只有 `app.name` 一条链路是通的（config.yaml → GetSystemInfo → store → Layout 文字渲染），品牌 logo、浏览器标签图标（favicon）、页脚（footer）三者都缺失或断裂：后端 `AppConfig` 无 logo/favicon 字段、无静态文件服务；前端 `Layout.vue` 只渲染文字无 `<img>`；`index.html` 引用的 favicon 文件实际不存在（404）；`app.footer` 是定义了却从未被前端消费的死配置；移动端甚至没有 `fetchSystemInfo` 整条链路。本 change 补齐品牌配置的完整能力，让品牌名称、logo、favicon、页脚都可由 `config.yaml` 设置。

## What Changes

- 后端 `config.go` 的 `AppConfig` 新增 `logo`、`favicon`、`footer` 三个字段（默认值 + YAML 覆盖，沿用既有三层配置范式）。
- 后端新增静态文件服务：`r.Static("/static", "./static")`，品牌图片放 `backend/static/`。
- `GetSystemInfo` 扩展返回 `logo`、`favicon`（文件名拼成 `/static/<file>` 完整路径）、`footer`、`name`、`subtitle`。
- 前端 `Layout.vue` 侧边栏 logo 由文字改为 `<img :src="logo">`（无 logo 时回退文字）；底部渲染 footer。
- 前端 `stores/app.js` 的 `fetchSystemInfo` 拉取 logo/favicon/footer，并动态设置浏览器 `document` 的 favicon link。
- 移动端补齐 `fetchSystemInfo` 链路，Home 页 navbar 与 Login 页 header 支持 logo 展示。
- 前端 `vite.config.js` 与移动端 `vite.config.js` 的 dev proxy 新增 `/static` 转发到后端。
- 更新 `config.example.yaml` 与 `docs/配置体系.md` 说明新字段。

## Capabilities

### New Capabilities

- `brand-config`: 品牌名称、logo、favicon、页脚四要素的可配置化，含后端静态文件托管、前端/移动端渲染、运行时 favicon 动态设置。

### Modified Capabilities

（无。本 change 是纯增量能力，不改动既有 spec 级行为。）

## Impact

- 修改文件：`backend/config/config.go`、`backend/controllers/system.go`、`backend/router/router.go`（静态服务）、`frontend/src/layout/Layout.vue`、`frontend/src/stores/app.js`、`mobile/src/stores/app.js`、`mobile/src/views/Home.vue`、`mobile/src/views/Login.vue`、`frontend/vite.config.js`、`mobile/vite.config.js`、`backend/config/config.example.yaml`、`docs/配置体系.md`。
- 新增目录：`backend/static/`（含占位 `.gitkeep`）。
- 无新第三方依赖（`r.Static` 是 Gin 原生；favicon 动态设置是浏览器原生 DOM 操作）。
- 部署影响：生产环境前后端分离时，nginx 需将 `/static` 转发到后端（写入 docs 说明）。
