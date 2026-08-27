# Design: brand-config

## Context

基座品牌配置现状（已审计）：

- **名称链路已通**：`config.yaml: app.name` → `config.go` 解析 → `GetSystemInfo` 返回 → `frontend/src/stores/app.js` 的 `fetchSystemInfo` → `Layout.vue` 文字渲染。
- **logo 断裂**：`AppConfig` 无 `logo` 字段；后端无静态文件服务（`backend/` 无 static 目录、router 无 `r.Static`）；前端 `Layout.vue` 只渲染文字无 `<img>`。
- **favicon 断裂**：`frontend/index.html` 的 `<link rel="icon" href="/favicon.ico">` 引用的文件**实际不存在**（现在就是 404）。
- **footer 是死配置**：`config.go` 定义了 `AppConfig.Footer`、`config.example.yaml` 有示例，但前端从未消费。
- **移动端无系统信息链路**：`mobile/src/stores/app.js` 只有 `setSystemName`，没有 `fetchSystemInfo`，也不 import `getSystemInfo`；`systemName` 永远空、显示兜底值。

## Goals / Non-Goals

**Goals:**

1. `config.yaml` 可配置品牌名称、logo、favicon、页脚四要素。
2. 后端托管品牌图片（`r.Static("/static", "./static")`），零新依赖。
3. 前端侧边栏、移动端 navbar 渲染 logo；浏览器标签图标（favicon）运行时跟随 config。
4. 页脚（footer）在前端 Layout 底部渲染，接上死配置。
5. 移动端补齐 `fetchSystemInfo` 链路，与前端能力对齐。

**Non-Goals:**

- 不做运行时上传 logo（管理员后台上传替换）——logo 通过 config + 文件部署设置，非数据库存储。
- 不做多租户/按用户定制品牌。
- 不做 logo 图片的尺寸/格式校验（由部署者保证文件合法）。
- 不引入 CDN 方案（方案 B 的完整 URL 形态）——本 change 只做方案 A（本地静态文件）。

## Decisions

### D1：logo/favicon 是两个独立字段，favicon 缺省回退 logo；logo 默认值指向内置图

`AppConfig` 新增 `logo`、`favicon` 两个独立字段，`favicon` 为空时回退到 `logo`。

**默认值（模式一：开箱即有默认图）**：`logo` 代码内默认值设为 `"logo.png"`，脚手架自带一张默认 logo 放在 `backend/static/logo.png` 并**提交进版本库**。零配置启动即有默认品牌图；用户改品牌只需替换该文件（甚至不用动 config），或改 `config.yaml` 指向别的文件名。

**Why：** 浏览器图标通常是 16×16 的 `.ico`，侧边栏 logo 通常是更大的 png/svg，尺寸不同，需独立配置。回退逻辑让「只想用一个图」的用户只填 `logo` 即可；`logo` 默认指向内置图，让脚手架「clone 即可见完整品牌」而非空荡荡的文字。

**Alternatives considered：**
- 单字段共用 → 拒绝，无法满足不同尺寸需求。
- favicon 强制必填 → 拒绝，增加无谓配置负担。
- logo 默认空（模式二，回退文字）→ 拒绝，用户明确要「开箱有默认图」，模式一更贴合。

### D2：静态服务路径约定 —— `r.Static("/static", "./static")`

后端注册 `r.Static("/static", "./static")`，品牌图片放 `backend/static/`。`GetSystemInfo` 把文件名拼成 `/static/<file>` 完整路径返回。

**Why：** `r.Static` 是 Gin 原生，零依赖。文件名→路径的拼接收敛在后端，前端拿到完整 URL 直接用，不感知"文件名 vs 路径"的差别。

**Alternatives considered：**
- 前端拼路径（后端只返回文件名）→ 拒绝，路径规则泄漏到前端，且移动端/前端要重复拼。
- base64 内联 → 拒绝（探索阶段已否），config 变丑难维护。

### D3：favicon 运行时动态设置（前端 store 内聚）

前端 `fetchSystemInfo` 拿到 favicon 后，动态改写 `document` 的 `link[rel="icon"]`（无则创建）。

**Why：** `index.html` 的 icon 是编译期写死，无法被运行时 config 驱动；动态 DOM 是唯一不改构建流程的解法，且收敛在 store 的 `fetchSystemInfo` 一处。

**Alternatives considered：**
- 编译期模板变量（vite env）→ 拒绝，env 也是编译期，改品牌仍要重新 build，违背「改 config 即改品牌」目标。
- 服务端渲染注入 → 拒绝，本项目无 SSR。

### D4：移动端补齐 `fetchSystemInfo`

给 `mobile/src/stores/app.js` 补上 `fetchSystemInfo`（与前端同构），Home/Login 页在 `onMounted` 调用，navbar/header 渲染 logo。

**Why：** 移动端无系统信息链路是上一轮「消除硬编码」遗留的 gap，本 change 的「移动端支持 logo」依赖它。

**Alternatives considered：**
- 移动端只做 logo 不做 name/footer → 拒绝，半套链路会再次漂移，应一次补齐。

### D5：dev proxy 加 `/static`

`frontend/vite.config.js` 与 `mobile/vite.config.js` 的 `server.proxy` 各加 `/static` → `http://localhost:8080`。

**Why：** dev 模式下前端端口（5173/5174）直接请求 `/static/logo.png` 会 404，必须代理到后端 8080。这是方案 A 的固有成本，遗漏会导致 dev 环境 logo 不显示。

### D6：页脚渲染位置

前端 `Layout.vue` 底部（`el-footer` 或主容器底部）渲染 `appStore.footer`，为空则不显示。移动端 Home 页底部渲染 footer。

**Why：** 接上死配置，且空值不显示避免空白。

## Risks / Trade-offs

- [方案 A 引入前后端分离部署耦合：生产环境 `/static` 需 nginx 转发] → 写入 `docs/配置体系.md` 的部署说明；dev 环境由 vite proxy 解决。
- [静态文件目录 `backend/static/` 若被 `.gitignore` 误杀，占位文件丢失] → 加 `static/.gitkeep`，确认 `.gitignore` 不排除 static；默认占位 logo 提交进版本库（脚手架自带资产），用户替换的品牌图若含敏感信息可自行 gitignore。
- [favicon 动态设置与浏览器缓存：改了 favicon 但浏览器仍显旧图标] → 不影响功能正确性，接受；可在 docs 提示强刷。
- [移动端 `van-nav-bar` 的 `:title` 当前是纯文字，插 logo 需改结构] → 用 `left` 插槽放 `<van-image>`，标题文字保留或改用 logo，落地时按视觉定。
- [logo 未配置时前端应优雅回退文字] → 所有渲染点统一「无 logo 则显示文字」的兜底，避免破图。

## Migration Plan

1. 后端：`config.go` 加三字段 → `router.go` 加静态服务 → `system.go` 扩展响应 → 建 `backend/static/.gitkeep`。
2. 前端：`stores/app.js` 拉取并设 favicon → `Layout.vue` 渲染 logo/footer → `vite.config.js` 加 proxy。
3. 移动端：`stores/app.js` 补 `fetchSystemInfo` → Home/Login 渲染 logo → `vite.config.js` 加 proxy。
4. 收口：`config.example.yaml` + `docs/配置体系.md` 补文档，跑 smoke + build 验证。

**Rollback：** 纯增量，移除静态服务与字段即回滚；前端 logo 未配置时回退文字，无破坏。

## Open Questions

- **移动端 logo 的具体视觉**：`van-nav-bar` 的 title 是文字居中，logo 放 `left` 插槽还是替换 title 文字？落地时按实际观感定，功能上先保证「有 logo 则显示」。
- **`app.footer` 是否也进移动端**：移动端页面通常无页脚概念，暂定仅前端 Web 端渲染 footer，移动端不做（若用户需要再加）。
