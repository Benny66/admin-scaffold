# Design: add-static-serving

## Context

脚手架后端当前是纯 API 服务，唯一静态能力是 brand-config 引入的 `r.Static("/static", "./static")`（品牌 logo/favicon）。前端靠 `vite dev`（5173），移动端靠另一 dev server（5174）。无 dist 托管、无 `/m/` 路由、无 SPA 回退。

路由拓扑现状：所有业务接口都在 `/api` 前缀下；`/static` 是品牌静态资源。这为 SPA 回退提供了干净的边界——`NoRoute` 只需排除 `/api/` 与 `/static/` 前缀即可避免误伤。

本 change 是 `add-multi-platform-packaging` 的前置：打包产物（dist / dist-mobile）落盘后，需由后端托管才能"双击即跑"。

## Goals / Non-Goals

**Goals:**

1. dist 存在时，后端托管桌面端（`/`）并提供 SPA 回退。
2. dist-mobile 存在时，后端托管移动端（`/m/`）。
3. 无 dist/dist-mobile 时优雅降级为纯 API（`go run` 开发态行为与现状完全一致）。
4. 不误伤 `/api/` 与 `/static/` 路由。

**Non-Goals:**

- 不做 `go:embed` 单二进制（见 D1）。
- 不做 HTTPS/TLS 证书、移动端扫码等（属打包/业务 change）。
- 不做静态资源缓存策略/压缩（gzip）优化。
- 不做前端 history 路由的 404 状态码精细化（SPA 回退统一返回 index.html）。

## Decisions

### D1：文件路径 + 存在性检查，而非 go:embed；桌面端用 NoRoute 而非 r.Static("/")

托管用 `r.Static` / `r.NoRoute` 指向 `./dist`、`./dist-mobile`，用 `os.Stat` 判断目录存在，不存在则跳过。

**桌面端（根路径）必须用 `r.NoRoute` + 手动 `c.File`，不能用 `r.Static("/", ...)`：** 因为 `r.Static("/")` 内部注册的是 catch-all 通配路由 `GET /*filepath`，而 httprouter 不允许根级 catch-all 与已存在的 `/api` 前缀共存——会直接 panic（`catch-all wildcard '*filepath' conflicts with existing path segment 'api'`）。`r.NoRoute` 只在无匹配路由时触发，天然不与 `/api`、`/static/`、`/m/` 冲突。

**Why：** 保持脚手架「零配置 `go run main.go`」的核心承诺——embed 要求编译时 dist 必须存在，会强制开发者也先 build 前端，破坏开发循环。文件路径 + 存在性检查让「有产物→一体化，无产物→纯 API」自动切换。

**Alternatives considered：**
- `go:embed` → 拒绝，破坏零配置开发，且编译期耦合 dist。
- 强制 dist 必存在 → 拒绝，纯 API 开发态（如只调接口）会被无谓阻塞。
- `r.Static("/", "./dist")` → 拒绝（运行时 panic，路由冲突，见上）。

### D2：SPA 回退边界 —— NoRoute 排除 /api/、/static/、/m/

`r.NoRoute` 回退前，判断请求路径：以 `/api/`、`/static/`、`/m/` 开头则返回 JSON 404（走 `utils.Fail`）；若命中 `dist/` 下的真实文件（如 `/assets/xxx.js`）则直接返回该文件；否则回退 `dist/index.html`。

**Why：** 前端 history 路由（如 `/system/user` 刷新）需要回退 index.html；但 API/品牌图/移动端未命中应返回 JSON 而非 HTML，否则前端 axios 会解析到 HTML 报错。NoRoute 手动分派既能服务 dist 内的静态资源（JS/CSS/图片），又能做 SPA 回退，一举两得。

**Alternatives considered：**
- 无条件 NoRoute 回退 index.html → 拒绝，会破坏 dist 内静态资源（JS/CSS 也被回退成 HTML）。
- 只回退、不服务 dist 静态文件 → 拒绝，前端打包的 assets 无法加载。

### D3：托管逻辑独立成 `router/serve.go`

新增 `router/serve.go` 暴露 `setupStaticServing(r *gin.Engine)`，由 `router.go` 在 API 路由注册后调用。

**Why：** `router.go` 已较长（约百行），静态托管是独立关注点，拆开保持职责单一，也便于未来扩展。

**Alternatives considered：**
- 直接塞进 `router.go` → 简单但让 SetupRouter 变臃肿。

### D4：移动端路径 `/m/`

移动端 H5 挂 `/m/`（与资产系统 build.sh 的部署约定一致），`dist-mobile/` 目录对应 `/m/` 前缀。

**Why：** 与打包脚本后续约定的产物目录 `dist-mobile` 对齐，也与资产系统部署习惯一致。

## Risks / Trade-offs

- [NoRoute 回退可能吞掉未注册 API 路径的 404 语义] → 通过排除 `/api/`、`/static/`、`/m/` 前缀规避（D2），未命中 API 仍返回 JSON 404。
- [SPA 回退在子路径部署（如 nginx 反代到 `/base/`）时会错] → 本项目按根路径部署（打包脚本 start.sh 直接根路径启动），子路径反代不在 scope，写入 Open Questions。
- [dist 目录存在但内容为旧构建，开发者误以为托管了最新前端] → 这是打包流程责任（build 会先 npm build 再 copy），本 change 只做「存在即托管」，不做内容校验。
- [`os.Stat` 相对路径依赖工作目录] → 与现有 `r.Static("/static", "./static")` 及 config 的 DSN 解析一致（都相对可执行文件/工作目录），保持约定统一。

## Open Questions

- **子路径部署**：是否需要支持 `app.base_path` 之类的子路径前缀？当前按根路径部署，若未来需 nginx 反代到子路径，SPA 回退与资源引用需额外处理（前端 build 的 `base` 配置）。暂不支持。
- **是否顺带 gzip 静态压缩**：生产部署常对静态资源 gzip，Gin 需 `gin-contrib/gzip` 依赖。当前零依赖原则下暂不引入，列为后续可选。
