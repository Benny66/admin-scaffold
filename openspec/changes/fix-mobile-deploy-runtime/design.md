# fix-mobile-deploy-runtime — Design

## Context

当前「后端单进程托管前端产物」架构：`backend/router/serve.go` 在 `dist/` 存在时把桌面端托管于根路径（NoRoute + SPA 回退），在 `dist-mobile/` 存在时用 `r.Static("/m/", "./dist-mobile")` 托管移动端。`scripts/package.sh` 打包出 `deploy/`：二进制 + `dist/` + `dist-mobile/` + `config.yaml` + 启动脚本。

缺陷链路：移动端 `vite build` 使用默认 `base: '/'`，产物 HTML 以根路径绝对引用 `/assets/index-xxx.js`。部署后访问 `/m/` 只拿到 `index.html`，资源请求落到服务器根路径——命中桌面端 NoRoute 回退时返回 `dist/index.html`（HTML 被当 JS/CSS 解析 → 白屏），无桌面 `dist/` 时则 404。移动端 H5 实际「不可用」。

次要缺口：`deploy/` 未拷贝 `backend/static/`，而 `deploy/config.yaml` 默认 `logo: "logo.png"` 经 `/static/logo.png` 加载 → 部署后品牌图全挂。

## Goals / Non-Goals

**Goals:**
- 打包部署后 `/m/` 移动端 H5 可完整渲染（HTML + JS/CSS 资源 + 登录跳转 + 子路由刷新）。
- `deploy/` 携带品牌静态资源，`/static/*` 部署后可访问。
- 不改变桌面端托管与 dev 开发体验（dev 仍访问 `http://localhost:5174/`）。
- 用 spec 契约封住回归（static-serving / multi-platform-packaging 补验收场景）。

**Non-Goals:**
- 不改移动端页面逻辑、API 封装、鉴权流程。
- 不做「把 dist-mobile 也独立部署到任意根路径」的通用支持（基座约定由后端在 `/m/` 托管）。
- 不改桌面端 `dist/` 的托管方式。

## Decisions

### D1: 移动端产物与路由统一使用 vite 的 BASE_URL 作为子路径

`mobile/vite.config.js` 按 `command` 条件设置 `base`：

- `build` → `base: '/m/'`（资源以 `/m/assets/*` 输出与引用）
- `serve`（dev）→ `base: '/'`（保持现有开发体验）

`mobile/src/router/index.js` 将 `createWebHistory()` 改为 `createWebHistory(import.meta.env.BASE_URL)`——vite 注入的 `BASE_URL` 恒等于构建时的 `base`（build=`/m/`、dev=`/`），**一处配置两端自动对齐**，无需手写两份路径，也不会出现「router base 与资源 base 漂移」。

**为什么不是其他方案：**
- `base: './'` + hash 路由：URL 带 `#`，且相对路径对 history 深层路径脆弱；移动端已用 `createWebHistory`，改成 hash 是风格倒退。
- 后端在 HTML 响应里重写 `/assets` → `/m/assets`：依赖响应改写，缓存/CDN 语义差、脆弱。
- dev/build 都固定 `/m/`：破坏 `vite dev` 直接访问 5174 根的体验。

### D2: 后端 `/m/` 从 `r.Static` 改为「文件服务 + SPA 回退」

移除 `r.Static("/m/", "./dist-mobile")`，改为注册 `/m/*filepath` 处理器（仅 `hasMobile` 时注册）：

1. 取 `filepath` 参数，与 `"dist-mobile"` 拼接前先做 `path.Clean` + 前缀校验，拒绝目录穿越（手工文件服务必须显式防穿越，替代 `r.Static` 原本的内置防护）。
2. 拼接结果若是目录（如 `/m/`）则尝试追加 `index.html`。
3. 目标文件存在 → `c.File` 返回；不存在 → 返回 `dist-mobile/index.html`（SPA 回退，与桌面 NoRoute 对称）。

效果：
- `/m/assets/index-xxx.js` → 直接命中文件（base `/m/` 后资源都在此前缀下），不再落入根路径 NoRoute。
- `/m/login`、`/m/mine` 等 history 子路由直达/刷新 → 回退 `index.html`，前端接管渲染。
- 桌面 NoRoute 中已有的 `/m/` 排除分支保留（对未注册 API 等情形仍是安全兜底，防止 HTML 泄漏给资源请求）。

### D3: `deploy/` 补齐品牌资源

`scripts/package.sh` 组装阶段新增拷贝 `backend/static/` → `deploy/static/`。后端 `r.Static("/static", "./static")` 以进程工作目录（deploy/）服务，无需其他改动。

### D4: Spec 契约补齐（护栏）

- `static-serving`：把「移动端托管」需求从「返回 index.html」提升为「`/m/` 子路径完整托管 + SPA 回退」，场景覆盖资源可达、子路由刷新、目录缺失跳过。
- `multi-platform-packaging`：「部署包组装」清单加入 `static/`；新增「移动端产物按 `/m/` 构建」与「部署后 `/m/` 可完整渲染」需求，用 WHEN/THEN 固化为可回归的验收。

## Risks / Trade-offs

- **手工文件服务的目录穿越** → 显式 `path.Clean` + 拒绝 `..`/绝对路径前缀，对齐 `r.Static` 原防护语义。
- **`/m/*filepath` 与桌面 NoRoute `/m/` 排除分支并存** → 语义不冲突（有路由先匹配，不进 NoRoute）；保留分支作为未注册移动端接口的 JSON 404 兜底。
- **产物不再能脱离后端独立部署到任意根路径** → 这是基座既定部署约定（spec 固定 `/m/`），在 README 部署说明与 spec 中写明，避免误用。
- **老 `deploy/` 目录缺 `static/`** → 变更后需重新 `make package` 覆盖部署；无平滑升级路径，属打包交付物，重打即得。
- **`import.meta.env.BASE_URL` 对 build base 的强绑定** → 若未来有人在 vite 里改 base，router 自动跟随，不会静默失配；反之若手工改 router base 则会失配，故在 tasks/备注中强调「只允许通过 vite base 一处配置」。

## Migration Plan

1. 改 `mobile/vite.config.js`、`mobile/src/router/index.js`（D1）。
2. 改 `backend/router/serve.go` `/m/` 托管（D2）。
3. 改 `scripts/package.sh` 拷贝 `static/`（D3）。
4. 更新两份 spec delta（D4）。
5. 验证：`make package`（本地平台）→ `cd deploy && ./start.sh` → curl 断言 `/m/` 200、`/m/assets/*` 200 且 content-type 正确、`/m/login` 返回 index.html、`/static/logo.png` 200；桌面 `/` 与 `/api/login` 回归不受影响。旧 `deploy/` 直接删除重打。

## Open Questions

无。所有方向均已在上述决策中收敛。
