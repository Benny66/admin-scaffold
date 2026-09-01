# Tasks: login-brand-visual

> 前置：`brand-config-guard` 已落地且 `make test` 全绿。本 change 新增 2 个配置字段 = 8 处改动，护栏会把漏改变成红灯。

## 1. 后端：新增背景图配置字段

- [x] 1.1 `backend/config/config.go` ① `AppConfig` 新增 `LoginBg` 与 `LoginBgMobile` 两个字段（yaml tag：`login_bg` / `login_bg_mobile`）
- [x] 1.2 ② `init()` 默认值中补两字段（空串）
- [x] 1.3 ③ `yamlFile.App` 影子结构补两字段（护栏 G1 会校验）
- [x] 1.4 ④ 覆盖链补两个非空覆盖分支（护栏 G2 会校验）
- [x] 1.5 `backend/controllers/system.go` 的 `GetSystemInfo` 返回 `login_bg` 与 `login_bg_mobile`，经 `staticURL()` 拼成 `/static/<file>`
- [x] 1.6 `backend/config/config.example.yaml` 的 `app` 段补两字段注释，说明「放在 backend/static/ 下，留空回退渐变」及建议尺寸

## 2. 两端 store 消费新字段

- [x] 2.1 `frontend/src/stores/app.js`：state 增 `loginBg` / `loginBgMobile`，`fetchSystemInfo` 解构并持久化（护栏 G3 会校验）
- [x] 2.2 `mobile/src/stores/app.js`：同上
- [x] 2.3 移动端解析顺序实现：`login_bg_mobile` → `login_bg` → 空（回退渐变）

## 3. 桌面端登录页视觉重构

- [x] 3.1 `frontend/src/views/Login.vue` 改为分栏布局：左栏（背景层 + 品牌块：logo / systemName / subtitle）+ 右栏（表单）；scrim 仅在有背景图时叠加（D1）
- [x] 3.2 响应式：< 900px 折叠为单列，左栏变为顶部品牌带（品牌块保留，不整体隐藏）
- [x] 3.3 背景层实现 `new Image()` 预加载 + 淡入 + `onerror` 回退渐变（D3）
- [x] 3.4 移除硬编码的「系统管理基座」，改用 store 的 `subtitle`
- [x] 3.5 品牌块渲染 logo，无 logo 时回退系统名称文字
- [x] 3.6 底部挂载 `<AppFooter>`
- [x] 3.7 为登录页引入局部 CSS 变量 token（见 Open Questions 的倾向方案）

## 4. 移动端登录页视觉重构

- [x] 4.1 `mobile/src/views/Login.vue` 改为分段布局：上部品牌区（背景图 + scrim + logo + 系统名 + subtitle）+ 下方表单区（D7，表单不浮于背景之上）
- [x] 4.2 复用与桌面端对齐的「预加载 + 淡入 + 回退」逻辑，scrim 同样仅在有图时叠加
- [x] 4.3 底部挂载 `<AppFooter>`，并处理 `env(safe-area-inset-bottom)` 安全区
- [x] 4.4 移除硬编码的「移动端」副标题，改用 store 的 `subtitle`

## 5. footer 组件抽取

- [x] 5.1 新增 `frontend/src/components/AppFooter.vue`，内容为纯文本 `app.footer`，为空时不渲染
- [x] 5.2 `frontend/src/layout/Layout.vue` 的 `.app-footer` 替换为 `<AppFooter>`（行为不变）
- [x] 5.3 新增 `mobile/src/components/AppFooter.vue`，移动端登录页使用

## 6. UX 与生产环境隐患修复

- [x] 6.1 两端默认账号提示改为 `v-if` + `import.meta.env.DEV`（D6）
      > 实现修正：仅靠 `v-if` 包裹字面量时，字符串仍会留在生产 bundle 中（`grep dist` 可命中 `admin123`，
      > 安全扫描照样报警）。改为三元表达式常量，使生产构建在 `DEV` 替换为 `false` 后折叠掉整段。已验证
      > 两端 dist 均不含该字符串，源码中仍保留。
- [x] 6.2 补 `autocomplete="username"` 与 `autocomplete="current-password"`
- [x] 6.3 username 输入框补 Enter 提交，与 password 对齐；username 框自动聚焦
- [x] 6.4 桌面端校验改为 `el-form :rules` 内联校验，替代 `ElMessage.warning`
- [x] 6.5 密码框补大小写锁定提示（`getModifierState('CapsLock')`）
- [x] 6.6 `frontend/src/utils/request.js` 与 `mobile/src/utils/request.js`：支持 `skipGlobalError` 标记，跳过一切全局错误呈现（toast / 「登录已过期」弹窗 / 清 token），并 reject 携带 message 的 Error（D4）
- [x] 6.7 两端 `Login.vue` 的登录请求带上该标记，失败时在表单内联错误区展示一次，并 catch 掉 Promise rejection
      > 标记原命名为 `skipAuthRedirect`（只豁免 401 重定向），但那会留下「全局 toast +
      > 内联错误」双重提示同一条消息。故扩为 `skipGlobalError`：错误由调用方自行呈现时，
      > 拦截器不再做任何全局呈现。非豁免请求的 reject 值与提示文案保持原样，无行为变更。
- [x] 6.8 `frontend/index.html` 移除指向不存在的 `/favicon.ico` 的 link（运行时由 `fetchSystemInfo` 注入）
- [x] 6.9 `mobile/src/views/Home.vue` 的外链猫图替换为 `appStore.logo` 或本地占位

## 7. 验证

- [x] 7.1 `make test` 全绿（含 brand-config-guard 对新字段的校验）
- [x] 7.2 `make lint` 无新增告警（`src/` 两端全清；`vite.config.js` 的 `process is not defined` 为既存问题，与本 change 无关）；`make smoke` 通过
- [ ] 7.3 手工验证：未配置背景图 → 渐变 + 品牌块；配置桌面图 → 左栏图片淡入；配置移动图 → 移动端生效、桌面端不受影响
      > 已完成接口层验证：`/api/system/info` 在 `login_bg`/`login_bg_mobile` 配置后正确返回 `/static/bg.png`、
      > `/static/bg_m.png`，未配置时返回空串。浏览器端渲染效果待人工过目。
- [ ] 7.4 手工验证：把 `login_bg` 指向不存在的文件 → 回退渐变，无破图、无白屏、无控制台报错
- [ ] 7.5 手工验证：1024 / 1280 / 1440 / 375 四个宽度下布局正常，footer 可见
- [x] 7.6 手工验证：生产构建下默认账号提示不显示（`grep -r admin123 dist/` 两端均无命中）
- [ ] 7.7 手工验证：密码错误时表单内联报错，不弹「登录已过期」
      > 已完成分支验证：后端登录失败返回 `HTTP 200 + code=401`（`utils.Fail` 不发真 401），
      > 走拦截器成功分支 → `ElMessage.error` + reject → `handleLogin` 的 catch 展示内联错误，无未捕获异常。
      > 浏览器端表现待人工过目。
