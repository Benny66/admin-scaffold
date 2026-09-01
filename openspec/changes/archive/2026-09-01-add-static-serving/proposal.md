# Proposal: add-static-serving

## Why

脚手架当前是纯 API 服务：`go run main.go` 只 serve `/api/*` 与 `/static/*`（品牌图），前端靠 `vite dev`、移动端靠另一个 dev server 各自跑。这导致脚手架无法独立交付——即便后续能编译出多平台二进制，打出的包也没有任何页面可访问。本 change 让后端在「前端产物存在」时自动托管桌面端（`/`）与移动端（`/m/`），使脚手架具备可部署的一体化能力；产物不存在时优雅降级为纯 API（保持零配置开发体验）。

## What Changes

- 后端新增前端产物托管：存在 `dist/` 时 `r.Static("/", "./dist")` + `NoRoute` 回退 `index.html`（SPA 路由）。
- 后端新增移动端托管：存在 `dist-mobile/` 时 `r.Static("/m/", "./dist-mobile")`。
- 存在性检查：用 `os.Stat` 判断目录，不存在则跳过（纯 API 降级，不破坏现有 `go run` 开发态）。
- 新增 `router/serve.go`（或等价文件）承载托管逻辑，保持 `router.go` 职责单一。
- 更新 `docs/目录结构约定.md` 与 `docs/配置体系.md`，说明部署目录布局与降级行为。

## Capabilities

### New Capabilities

- `static-serving`: 后端对桌面端/移动端前端产物的条件托管与 SPA 回退，支持无产物时优雅降级为纯 API。

### Modified Capabilities

（无。）

## Impact

- 修改文件：`backend/router/router.go`（调用新托管逻辑）、新增 `backend/router/serve.go`、`docs/目录结构约定.md`、`docs/配置体系.md`。
- 无新第三方依赖（`os.Stat` + Gin 原生 `Static`/`NoRoute`）。
- 无破坏：`dist/`、`dist-mobile/` 不存在时行为与现状完全一致。
- 为后续 `add-multi-platform-packaging`（多平台打包）提供前置：打包产物落盘后即可被本 change 托管。
