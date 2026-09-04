## Why

基座当前是「后端 + Web 前端 + Vant H5 移动端」三端，缺微信小程序入口。企业管理类项目（资产盘点、移动审批、现场巡检）的真实场景在微信小程序侧——它需要一个能直接复用基座 JWT/RBAC/响应协议、又能贴合小程序原生交互（`wx.login`、`pages.json`、`uni.request`）的第四端。现在补这一端，既能让基座真正"多端"，也顺势把"三端统一"铁律推广到多端语境。

## What Changes

- 新增 `miniapp/` 目录（uniapp + vue3 + pinia + uni-mp-weixin），与 `frontend/`、`mobile/` 平级；`mobile/`（Vant H5）保留不动。
- 后端新增微信小程序登录接口 `POST /api/auth/mp-login`：接 `code` → 调微信 `jscode2session` → 拿 `openid` → 查/建 `User` → 签发与账号密码登录**同一种** JWT。
- `User` 模型新增 `OpenID` 字段（可空、唯一索引），与现有 `username/password` 登录方式并存。
- 新增配置段 `wechat.appid` + `wechat.secret`，同步 `config.example.yaml`；后端不引第三方 SDK，直接 `net/http` 调微信接口。
- 三端铁律落地到 miniapp：沿用 `stores/`（复数）与 `@ → src/`；**不强求 `views/`**（uniapp `pages.json` 是硬约定，强扭反而不地道）。
- HTTP 封装：miniapp 不引入 axios，基于 `uni.request` 自包 `utils/request.js`（统一 `Authorization` 头、401 处理、错误提示）。
- eslint flat config 扩展到 `miniapp/`：新增"禁止直接调 `uni.request`，必须走 `@/utils/request`"规则；现有 axios 禁令的 patterns 在 miniapp 下豁免（miniapp 本就没 axios）。
- `deps.yaml` 新增 `miniapp:` 段，登记 `@dcloudio/uni-app`、`@dcloudio/uni-mp-weixin`、`@dcloudio/vite-plugin-uni`、`vue`、`pinia`。
- `Makefile` 新增 `dev-mp`、`build-mp`；`lint` 扩展为后端 vet + 前端 + 移动端 + miniapp ESLint。`make dev` 不并行起 miniapp（小程序 dev 需要微信开发者工具配合，按需起）。
- `scripts/init.sh` 扩展：改 `miniapp/package.json` 的 `name`、`miniapp/src/manifest.json` 的 `appid` 与 `name`、`.env` 前缀。
- `.gitignore` 增加 `miniapp/dist/`、`miniapp/unpackage/`（uniapp 默认产物目录）。
- 叙事改动：`README.md` 标题与目录结构图从「三端」改「多端」；`AGENTS.md §1` 标题与措辞同步。
- miniapp 下提供最小范例：`pages/login/`（接 `mp-login`）+ `pages/index/`（拉取 `/system/info` 显示系统名），对标 `backend/_example/`「黄金路径唯一范例」思路。
- **不扩展** `scripts/package.sh`：小程序发布走微信开发者工具上传，不进 `deploy/` 目录。

## Capabilities

### New Capabilities

- `miniapp-wechat-end`: 微信小程序端骨架——uniapp + vue3 + pinia，复用基座 JWT/RBAC/响应协议；目录约定（`pages/` 而非 `views/`、`stores/` 复数、`@ → src/`）、`uni.request` 封装层、最小范例页面。
- `wechat-mp-login`: 微信小程序登录流程——`code → jscode2session → openid → 签发同样的 JWT`；`User.openid` 字段、`POST /api/auth/mp-login` 接口、`wechat` 配置段。

### Modified Capabilities

- `dependency-registry`: 登记范围从「后端/前端/移动端三端」扩展到「含 miniapp 四端」；deps.yaml 新增 `miniapp:` 段。
- `frontend-guardrails`: ESLint flat config 覆盖范围从 `frontend/` + `mobile/` 扩展到 `miniapp/`；miniapp 下"禁止直接 import axios"替换为"禁止直接调 `uni.request`"（同一规则语义、不同禁用模式）。
- `frontend-store-guard`: 护栏扫描范围从 `frontend/src` + `mobile/src` 扩展到 `miniapp/src`，三端同等覆盖。
- `project-init`: `init.sh` 的改名范围扩展——除现有 backend module 名、frontend/mobile 包名、env 前缀外，新增 miniapp/package.json name 与 manifest.json appid/name 的替换。
- `brand-config`: 品牌信息消费点扩展——`GetSystemInfo` 返回的五字段被 `miniapp/src/stores/app.js` 的 `fetchSystemInfo` 同构消费（小程序 navbar 显示 systemName/logo）。

## Impact

- 新增目录：`miniapp/`（含 `src/pages/login/`、`src/pages/index/`、`src/stores/app.js`、`src/utils/request.js`、`src/api/index.js`、`pages.json`、`manifest.json`、`package.json`、`vite.config.js`）。
- 新增后端文件：`backend/controllers/mp_auth.go`、`backend/services/mp_auth_service.go`、`backend/utils/wechat.go`（jscode2session 调用），并在 `router/router.go` 注册 `POST /api/auth/mp-login`（公开组）。
- 修改文件：
  - `backend/models/user.go`（新增 `OpenID` 字段 + 唯一索引）
  - `backend/config/config.go` + `config.example.yaml`（新增 `wechat` 段）
  - `backend/database/migrations.go` 或对应迁移文件（`openid` 列自动迁移）
  - `deps.yaml`（新增 `miniapp:` 段）
  - `Makefile`（新增 `dev-mp` / `build-mp`，扩展 `lint`）
  - `eslint.config.js`（扩展覆盖 miniapp，新增 uni.request 禁令、给 miniapp/utils/request.js 豁免）
  - `scripts/init.sh`（扩展改 miniapp 包名 + manifest）
  - `.gitignore`（加 `miniapp/dist/`、`miniapp/unpackage/`）
  - `README.md` + `AGENTS.md`（三端 → 多端叙事）
  - `docs/目录结构约定.md`、`docs/鉴权与权限.md`（补充 miniapp 一节）
- 新增依赖（前端侧，登记到 `deps.yaml` 的 `miniapp:` 段）：`@dcloudio/uni-app`、`@dcloudio/uni-mp-weixin`、`@dcloudio/vite-plugin-uni`、`vue`、`pinia`。后端**不引**第三方 SDK，直接 `net/http` 调 jscode2session。
- **破坏性**：无对外接口破坏；`POST /api/auth/mp-login` 是新增公开路由，不影响现有 `/api/auth/login`。
- 行为变化：`User` 表新增 `openid` 列（自动迁移加列，老用户该列为空，登录方式不变）；`AGENTS.md §1` 标题"三端统一"改为"多端统一"。
- **前置依赖**：`auth-token-lifecycle`（in-progress）必须先落地——它正在改 `frontend/mobile` 的 `request.js` 与 `Login.vue`，且小程序端拿到的 JWT 也要享受令牌版本吊销 + 刷新令牌 + 无感续期。本 change 在 `auth-token-lifecycle` 归档前不动 `miniapp/src/utils/request.js` 的续期逻辑，只做登录 + 基础请求封装。
- **非目标（明确排除）**：不做小程序的 H5 编译目标（uniapp 一统移动端属于未来 change）、不做小程序的 App 平台编译、不做 unionid 跨主体去重、不做手机号授权、不做小程序订阅消息。
