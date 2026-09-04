## ADDED Requirements

### Requirement: miniapp 端骨架存在

项目 MUST 在仓库根提供 `miniapp/` 目录（与 `frontend/`、`mobile/` 平级），基于 uniapp + vue3 + pinia 构建，编译目标为微信小程序（`uni-mp-weixin`）。基线 `miniapp/` 必须能 `npm run dev:mp-weixin` 启动并产出可供微信开发者工具打开的产物。

#### Scenario: clone 基座后启动 miniapp

- **WHEN** 开发者 clone 基座并执行 `cd miniapp && npm install && npm run dev:mp-weixin`
- **THEN** 产出 `miniapp/dist/dev/mp-weixin/` 目录，可被微信开发者工具导入并预览

#### Scenario: miniapp 与 mobile 并存不冲突

- **WHEN** 同时存在 `mobile/`（Vant H5）与 `miniapp/`（uniapp）
- **THEN** 两端 package.json name 不同、vite 配置互不影响、各自独立构建，互不引用对方代码

### Requirement: 多端统一铁律在 miniapp 落地

`miniapp/` MUST 沿用基座跨端铁律：状态管理目录统一为 `stores/`（复数，禁止 `store/`）；路径别名 `@` 指向 `src/`；后端字段 JSON tag 沿用 snake_case 透传。但页面目录 MUST 为 `pages/`（uniapp `pages.json` 硬约定，不强求与前端 `views/` 一致）。

#### Scenario: 状态管理目录为 stores 复数

- **WHEN** 检查 `miniapp/src/` 目录结构
- **THEN** 存在 `stores/` 目录，不存在 `store/` 单数目录

#### Scenario: 路径别名 @ 指向 src

- **WHEN** 在 `miniapp/src/` 下任意 `.vue` / `.js` 文件中 `import x from '@/stores/app'`
- **THEN** 解析到 `miniapp/src/stores/app.js`，而非报错或解析到其他位置

#### Scenario: 页面目录为 pages

- **WHEN** 检查 `miniapp/src/` 下页面文件位置
- **THEN** 页面位于 `pages/<page>/index.vue`，由 `pages.json` 注册；不存在 `views/` 目录

### Requirement: HTTP 请求必须走封装层

miniapp MUST 提供 `miniapp/src/utils/request.js` 作为 `uni.request` 的封装层，统一附带 `Authorization: Bearer <token>`、统一处理 401（清 token 跳登录）、403（提示无权限）、`code === 200` 返回 `res`。所有业务请求 MUST 走该封装，禁止直接调用 `uni.request`。封装层自身是唯一豁免。

#### Scenario: 业务页面发起请求

- **WHEN** `pages/index/index.vue` 需要拉取系统信息
- **THEN** 调用 `@/utils/request` 暴露的实例（如 `request.get('/system/info')`），不直接 `uni.request`

#### Scenario: 401 自动跳登录

- **WHEN** 任一请求响应 `code === 401`（或 HTTP 401）
- **THEN** 封装层清空本地 token，跳转 `pages/login/index`，而非把 401 抛给业务代码

#### Scenario: token 自动附带

- **WHEN** 已登录用户发起任一业务请求
- **THEN** 封装层从 storage 读取 token，自动设置 `Authorization: Bearer <token>` 头

### Requirement: eslint 护栏覆盖 miniapp

`eslint.config.js`（flat config）MUST 扩展覆盖 `miniapp/src/`，沿用 `no-restricted-imports` 禁止 `@/store/` 单数回潮；同时新增"禁止直接调用 `uni.request`"模式，强制所有请求走 `@/utils/request`。`miniapp/src/utils/request.js` 自身是唯一豁免（合法调用 `uni.request`）。

#### Scenario: AI 在 miniapp 业务文件直接调 uni.request

- **WHEN** `miniapp/src/pages/`、`miniapp/src/api/`、`miniapp/src/stores/` 下任一文件出现 `uni.request(`
- **THEN** ESLint 报错，提示必须使用 `@/utils/request` 的封装实例

#### Scenario: 封装层自身合法调用

- **WHEN** `miniapp/src/utils/request.js` 内部调用 `uni.request(`
- **THEN** 该调用被豁免，不触发规则

#### Scenario: stores 单数目录在 miniapp 也被禁

- **WHEN** miniapp 下任一文件 `import ... from '@/store/...'`（单数）
- **THEN** ESLint 报错，与 frontend/mobile 行为一致

### Requirement: 最小范例页面

`miniapp/` MUST 提供两个最小范例页面作为"黄金路径"，对标 `backend/_example/` 思路，禁止参照 `views/system/` 五件套历史模块：
- `pages/login/index.vue`：调 `POST /api/auth/mp-login`（接 wx.login 的 code），登录成功后存 token 跳 `pages/index/index`。
- `pages/index/index.vue`：调 `GET /api/system/info` 拉取系统名与 logo，渲染到 navbar。

#### Scenario: 范例页面跑通完整链路

- **WHEN** 在微信开发者工具中打开 miniapp 并完成登录
- **THEN** 登录成功跳首页，首页显示从后端拉取的 `systemName`，证明鉴权链路完整

#### Scenario: 新增业务页面参照范例

- **WHEN** AI 需要新增 miniapp 业务页（如资产盘点页）
- **THEN** 参照 `pages/login/` 与 `pages/index/` 的写法（请求走封装、store 走 `@/stores/`），而非参照其他历史项目

### Requirement: 工程基础设施扩展

`Makefile` MUST 提供 `dev-mp`、`build-mp` 目标，分别启动 miniapp dev 与构建 mp-weixin 产物；`make lint` MUST 扩展为「后端 vet + 前端 ESLint + 移动端 ESLint + miniapp ESLint」的聚合命令。`make dev`（并行起后端 + 前端）**不**自动起 miniapp，小程序 dev 需要微信开发者工具配合，按需手动起。

#### Scenario: 启动 miniapp dev

- **WHEN** 执行 `make dev-mp`
- **THEN** 在 `miniapp/` 下执行 `npm run dev:mp-weixin`，产出可供微信开发者工具打开的目录

#### Scenario: 构建小程序产物

- **WHEN** 执行 `make build-mp`
- **THEN** 在 `miniapp/` 下执行 `npm run build:mp-weixin`，产出 `miniapp/dist/build/mp-weixin/`

#### Scenario: lint 覆盖 miniapp

- **WHEN** 执行 `make lint`
- **THEN** 依次跑 backend go vet、frontend ESLint、mobile ESLint、miniapp ESLint，任一失败即非零退出码

### Requirement: 依赖登记

miniapp 直接依赖 MUST 登记到 `deps.yaml` 的 `miniapp:` 段（`@dcloudio/uni-app`、`@dcloudio/uni-mp-weixin`、`@dcloudio/vite-plugin-uni`、`vue`、`pinia`），由 guard 测试双向校验。

#### Scenario: 新增依赖登记

- **WHEN** AI 为 miniapp 引入新依赖（如 `uni-popup`）
- **THEN** 在 `deps.yaml` 的 `miniapp:` 段追加一条 `{ package, reason }`，而非仅 `npm install`

#### Scenario: 依赖登记与 package.json 一致

- **WHEN** guard 测试运行
- **THEN** miniapp package.json 的每个直接依赖都在 deps.yaml 登记，反之亦然

### Requirement: init.sh 扩展改 miniapp 标识

`scripts/init.sh` MUST 扩展为同时改写 `miniapp/package.json` 的 `name`、`miniapp/src/manifest.json` 的 `name` 字段；`manifest.json` 的 `appid` 字段保留基座占位（小程序 appid 由具体项目填入，不属于基座关注）。

#### Scenario: 一键初始化改 miniapp 包名

- **WHEN** 执行 `make init name=my-system`
- **THEN** `miniapp/package.json` 的 `name` 改为 `my-system-miniapp`，`miniapp/src/manifest.json` 的 `name` 改为 `my-system`

#### Scenario: appid 不被脚本改写

- **WHEN** 初始化执行完成
- **THEN** `miniapp/src/manifest.json` 的 `mp-weixin.appid` 保留为基座占位（如 `touristappid` 或空），由具体项目开发者填入

### Requirement: 产物不入版本库

`.gitignore` MUST 忽略 `miniapp/dist/`、`miniapp/unpackage/`（uniapp 默认产物目录）。

#### Scenario: 构建产物不被提交

- **WHEN** 执行 `make build-mp` 后 `git status`
- **THEN** `miniapp/dist/` 与 `miniapp/unpackage/` 不出现在未跟踪文件列表

### Requirement: 叙事自洽：多端而非三端

`README.md` 与 `AGENTS.md` MUST 把基座定位从「三端脚手架」改为「多端脚手架」，目录结构图与启动说明 MUST 包含 miniapp 一节。`AGENTS.md §1` 标题 MUST 从「三端统一」改为「多端统一」。

#### Scenario: 阅读 README 看到 miniapp

- **WHEN** 开发者查阅 `README.md` 目录结构图
- **THEN** 在 `backend/`、`frontend/`、`mobile/` 之外看到 `miniapp/` 一行及其定位说明

#### Scenario: AGENTS §1 标题更新

- **WHEN** 查阅 `AGENTS.md` §1
- **THEN** 标题为「多端统一」，正文涵盖 backend/frontend/mobile/miniapp 四端
