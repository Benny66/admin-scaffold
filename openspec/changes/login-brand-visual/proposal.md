# Proposal: login-brand-visual

## Why

基座是给人 clone 之后当门面用的，但登录页现在是一个 380px 的白卡片飘在渐变背景上：没有 logo、没有页脚、副标题硬编码「系统管理基座」。这个画面传递的信息是「这是一个 Demo」，而不是「这是一个可以交付的系统」。

更具体的问题有三个：

1. **品牌资产没用上。** `appStore.logo` 早就有值（brand-config 已跑通 `/static` 链路），但登录页完全没渲染它；`GetSystemInfo` 返回的 `subtitle` 没被两端 store 消费，于是副标题被硬编码在 `frontend/src/views/Login.vue:5`。
2. **两端各画各的。** 桌面端是居中卡片，移动端是居中表单，没有任何共享的品牌层，改一次品牌要改两遍。
3. **登录页没有 footer。** `app.footer` 只在 `frontend/src/layout/Layout.vue:58` 的 `.main` 里渲染，而 `/login` 是 router 的顶层兄弟节点（`frontend/src/router/index.js:7-11`），压根不经过 Layout——所以「页脚」这个品牌要素在系统的第一屏是缺席的。

同时，登录页还挂着几处不该出现在生产环境的东西：明码标价的默认账号提示（`frontend/src/views/Login.vue:30`、`mobile/src/views/Login.vue:32`）、`frontend/index.html:6` 里指向不存在的 `/favicon.ico`（每次加载 404）、缺失的 `autocomplete` 属性（密码管理器填不进去），以及移动端首页外链的 Vant demo 猫图（`mobile/src/views/Home.vue:14`）。

## What Changes

**视觉与布局**

- 桌面端登录页改为**左图右表单**分栏：左栏是可配置背景图（scrim 遮罩 + 品牌块：logo / 系统名称 / subtitle），右栏是登录表单。宽度 < 900px 时折叠为单列全屏背景 + 居中卡片。
- 移动端登录页支持可配置背景图（全屏 + scrim），表单卡片浮于其上，底部渲染 footer。
- 新增 `frontend/src/components/AppFooter.vue` 与 `mobile/src/components/AppFooter.vue`，`Layout.vue` 与 `Login.vue` 共用，消除 footer 渲染逻辑的重复。

**配置**

- `AppConfig` 新增 `login_bg`（桌面端背景图文件名）与 `login_bg_mobile`（移动端背景图，缺省回退 `login_bg`），由 `GetSystemInfo` 返回 `/static/<file>` 完整路径。
- 背景图加载失败或未配置时，回退到现有渐变——**基座开箱好看，配图更好看**。

**UX 修复**

- 默认账号提示改为仅在 `import.meta.env.DEV` 下显示。
- 补 `autocomplete="username"` / `autocomplete="current-password"`、username 框的 Enter 提交、autofocus、大小写锁定提示。
- 桌面端校验方式对齐移动端：用 `el-form :rules` 内联校验替代 `ElMessage.warning`。
- 登录请求标记豁免全局 401 拦截器（见 design D4），失败时在表单内联展示错误态，并消除未捕获的 Promise rejection。
- 移除 `frontend/index.html` 中不存在的 `/favicon.ico` link；移除移动端首页外链的 Vant demo 猫图。

## Capabilities

### New Capabilities

- `login-brand-visual`: 登录页的品牌化视觉体系——可配置背景图、分栏布局、品牌块、两端共享的 footer 组件，以及配套的加载降级、响应式折叠与登录表单 UX 规范。

### Modified Capabilities

- `brand-config`: 新增 `login_bg` / `login_bg_mobile` 两个品牌字段，并要求两端 store 消费（由 brand-config-guard 的护栏强制四处副本同步）。

## Impact

- 修改文件：
  - `backend/config/config.go`（4 处：AppConfig / 默认值 / yamlFile / 覆盖链）、`backend/controllers/system.go`、`backend/config/config.example.yaml`
  - `frontend/src/views/Login.vue`、`frontend/src/layout/Layout.vue`、`frontend/src/components/AppFooter.vue`（新增）、`frontend/src/stores/app.js`、`frontend/src/utils/request.js`、`frontend/index.html`
  - `mobile/src/views/Login.vue`、`mobile/src/components/AppFooter.vue`（新增）、`mobile/src/stores/app.js`、`mobile/src/utils/request.js`、`mobile/src/views/Home.vue`
- 无新第三方依赖（背景图走既有 `/static` 链路，scrim / 渐变 / fade-in 均为原生 CSS）。
- 无破坏性：背景图未配置时回退现有渐变；布局从「居中卡片」变为「分栏」属预期内改版。
- **依赖 `brand-config-guard` 先落地**：新增 2 个字段 = 8 处改动，护栏会把漏改变成构建红灯，而非静默产生「配了不生效」。
