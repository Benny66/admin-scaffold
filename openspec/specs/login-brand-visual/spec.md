# Spec: login-brand-visual

登录页的品牌化视觉体系：可配置背景图、左图右表单分栏、品牌块、两端共享的 footer，以及配套的加载降级、响应式折叠与登录表单 UX 规范。

## Requirements

### Requirement: 桌面端登录页采用左图右表单分栏布局

桌面端（`frontend/src/views/Login.vue`）MUST 采用分栏布局：左栏为背景展示区（背景层 + scrim + 品牌块），右栏为登录表单区（含 footer）；视口宽度 < 900px 时 MUST 折叠为单列（左栏隐藏，背景铺满，表单卡片居中）。

左栏渐变背景的颜色源 MUST 使用 CSS 变量 `--brand-from` 与 `--brand-to`（由 `brand-color-extract` 能力注入），禁止硬编码色值。scrim 的半透明色 MUST 从同一变量派生。

> **CSS 变量覆盖陷阱：** 登录页局部作用域 MUST NOT 重新声明 `--brand-from` / `--brand-to`。
> 自定义属性在元素自身声明的值会覆盖继承自 `:root` 的注入值，而
> `--brand-from: var(--brand-from, #1f6feb)` 是自引用循环，整个变量会变成
> guaranteed-invalid，导致品牌色永远不生效。fallback 只在取值处写：
> `linear-gradient(135deg, var(--brand-from, #1f6feb) 0%, var(--brand-to, #0d3b8c) 100%)`。

#### Scenario: 宽屏渲染分栏

- **WHEN** 视口宽度 ≥ 900px 且已加载系统信息
- **THEN** 页面呈现左右两栏，左栏含背景层与品牌块，右栏含登录表单与 footer

#### Scenario: 窄屏折叠为单列

- **WHEN** 视口宽度 < 900px
- **THEN** 左栏隐藏，背景铺满整个视口，表单卡片居中显示，无横向滚动条

#### Scenario: 渐变跟随主题色

- **WHEN** `--brand-from` 与 `--brand-to` 被注入 `:root`
- **THEN** 左栏渐变背景与 scrim 使用这两个变量值，不出现硬编码蓝色

#### Scenario: 主题色未注入时回退

- **WHEN** `--brand-from` 与 `--brand-to` 未被注入（如提取失败回退场景）
- **THEN** CSS 变量使用取值处的 fallback 值（桌面端 `#1f6feb` / `#0d3b8c`）

### Requirement: 移动端登录页采用上部品牌区与下方表单的分段布局

移动端（`mobile/src/views/Login.vue`）MUST 采用分段布局：页面上部为品牌区（背景层 + scrim + 品牌块），下部为表单区（纯色背景 + 登录表单 + footer）；表单 MUST NOT 浮于背景图之上。

品牌区渐变同样取 `--brand-from` / `--brand-to`，其 fallback 值为移动端自己的 `#1989fa` / `#0d3b8c`。

#### Scenario: 移动端渲染分段布局

- **WHEN** 在移动端视口打开登录页
- **THEN** 页面上部为品牌区（含背景层与品牌块），下部为表单区，两部分不重叠

#### Scenario: 背景图在品牌区完整可见

- **WHEN** 配置了移动端背景图
- **THEN** 该图在品牌区内可见，不被表单卡片遮挡

#### Scenario: 表单区可读性不依赖 scrim

- **WHEN** 表单区渲染
- **THEN** 表单与输入区位于纯色背景之上，可读性不依赖半透明卡片或遮罩

#### Scenario: 渐变跟随主题色

- **WHEN** `--brand-from` 与 `--brand-to` 被注入 `:root`
- **THEN** 移动端品牌区渐变使用这两个变量值

### Requirement: 品牌块展示 logo 与配置文案

品牌块 MUST 展示品牌 logo（无 logo 时回退系统名称文字）与 `subtitle`；副标题文案 MUST 来自 `GetSystemInfo` 的 `subtitle` 字段，禁止硬编码。

#### Scenario: 展示 logo 与 subtitle

- **WHEN** 系统信息返回非空 `logo` 与 `subtitle`
- **THEN** 品牌块显示 logo 图片与 subtitle 文案

#### Scenario: 无 logo 时回退

- **WHEN** 系统信息返回空 `logo`
- **THEN** 品牌块回退显示系统名称文字，不产生破图

#### Scenario: subtitle 为空

- **WHEN** `subtitle` 为空
- **THEN** 品牌块不渲染副标题行，不留空白占位

#### Scenario: logo 配置存在但文件加载失败

- **WHEN** `logo` 非空，但 `/static/<file>` 实际返回 404（文件未放入 `backend/static/`）
- **THEN** 所有 logo 展示点（登录页品牌块、Layout 侧边栏、移动端品牌区、移动端首页头像）回退为文字，不产生破图

#### Scenario: logo 失败状态不持久化

- **WHEN** logo 曾加载失败，使用者随后把文件补进 `backend/static/` 并刷新页面
- **THEN** logo 正常显示，无需手动清缓存

### Requirement: 登录页背景图可通过 config.yaml 配置

`AppConfig` MUST 提供 `login_bg`（桌面端）与 `login_bg_mobile`（移动端）两个字段，`GetSystemInfo` MUST 将其返回为 `/static/<file>` 完整路径；移动端解析顺序 MUST 为 `login_bg_mobile` → `login_bg` → 无背景。

#### Scenario: 配置桌面端背景图

- **WHEN** `config.yaml` 的 `app.login_bg` 设为 `bg.png` 且 `backend/static/bg.png` 存在
- **THEN** 桌面端登录页左栏显示该图片

#### Scenario: 移动端使用专属背景图

- **WHEN** `login_bg_mobile` 已配置
- **THEN** 移动端显示该图，桌面端仍显示 `login_bg`

#### Scenario: 移动端回退桌面端图

- **WHEN** `login_bg_mobile` 未配置但 `login_bg` 已配置
- **THEN** 移动端显示 `login_bg`

#### Scenario: 未配置背景图

- **WHEN** `login_bg` 与 `login_bg_mobile` 均为空
- **THEN** 两端登录页回退为渐变背景，品牌块与表单正常显示

### Requirement: 背景图加载失败时优雅降级

背景图 MUST 通过 `new Image()` 预加载，加载成功后淡入显示；加载失败或文件不存在时 MUST 回退为渐变背景，且不得出现破图、白屏或控制台报错。

#### Scenario: 背景图加载成功

- **WHEN** 配置的背景图可正常加载
- **THEN** 图片淡入显示在背景层，不出现「白闪后突然出现」

#### Scenario: 背景图 404

- **WHEN** `login_bg` 指向 `backend/static/` 下不存在的文件
- **THEN** 背景层保持渐变，页面无破图、无白屏，控制台仅输出一条 warn

#### Scenario: 加载期间表单可用

- **WHEN** 背景图仍在加载中
- **THEN** 渐变背景已可见，表单可正常输入与提交，不依赖背景图加载完成

### Requirement: 登录页渲染 footer

桌面端登录页的 footer MUST 渲染在右栏（表单区）底部，而非跨全宽。右栏（`.login-form-area`）MUST 使用 `flex-direction: column`，表单卡片区 `flex: 1` 撑满上方，footer 自然落底。窄屏单列模式下 footer 仍在页面底部，行为不变。

移动端登录页的 footer 仍在页面底部渲染（分段布局不变）。

两端登录页 MUST 渲染 `<AppFooter>`，内容取 `app.footer`；`footer` 为空时 MUST 不渲染占位。移动端 MUST 处理底部安全区。

#### Scenario: 桌面端 footer 在右栏底部

- **WHEN** 视口宽度 ≥ 900px 且 `app.footer` 非空
- **THEN** footer 出现在右栏底部，左栏不受影响，footer 不跨左栏

#### Scenario: 窄屏 footer 在页面底部

- **WHEN** 视口宽度 < 900px 且 `app.footer` 非空
- **THEN** footer 仍在页面底部显示，不受分栏折叠影响

#### Scenario: footer 已配置

- **WHEN** `app.footer` 非空
- **THEN** 桌面端与移动端登录页底部显示该文案

#### Scenario: footer 未配置

- **WHEN** `app.footer` 为空
- **THEN** 登录页不渲染 footer 区域，不留空白

#### Scenario: 移动端安全区

- **WHEN** 在有底部安全区的设备（如 iPhone）上打开移动端登录页
- **THEN** footer 位于安全区之上，不被系统横条遮挡

### Requirement: footer 渲染逻辑两端复用

`frontend/src/components/AppFooter.vue` 与 `mobile/src/components/AppFooter.vue` MUST 作为 footer 的唯一渲染实现，`Layout.vue` 与 `Login.vue` MUST 复用它而非各自实现。

#### Scenario: Layout 复用组件

- **WHEN** 系统内页渲染页脚
- **THEN** 通过 `<AppFooter>` 渲染，行为与改动前一致

#### Scenario: footer 内容视为纯文本

- **WHEN** `app.footer` 中包含 HTML 片段（如 `<a href=...>`）
- **THEN** 按纯文本原样展示，不被解析为 HTML

### Requirement: 登录页不泄漏默认凭据

默认账号提示 MUST 仅在开发态（`import.meta.env.DEV`）渲染，生产构建 MUST NOT 出现。

#### Scenario: 开发环境

- **WHEN** 以 vite dev 模式访问登录页
- **THEN** 显示默认账号提示

#### Scenario: 生产构建

- **WHEN** 访问 `make build` 产出的登录页
- **THEN** 页面不含默认账号/密码明文

### Requirement: 登录失败在表单内联呈现且不被全局错误呈现重复提示

登录请求 MUST 携带 `skipGlobalError` 标记，豁免 `utils/request.js` 的一切全局错误呈现（toast、「登录已过期」弹窗、清 token）；登录失败时 MUST 且仅 MUST 在表单内联展示一次错误，不得重复提示，也不得产生未捕获的 Promise rejection。

#### Scenario: 密码错误

- **WHEN** 用户输入错误的密码并提交
- **THEN** 表单内联展示「用户名或密码错误」一次，不出现「登录已过期」弹窗，也不出现重复该消息的全局 toast

#### Scenario: 后端返回真 401 状态码

- **WHEN** 登录接口以 HTTP 401 状态码响应
- **THEN** 仍走表单内联错误展示，不触发全局会话过期流程

#### Scenario: 网络不可用

- **WHEN** 登录请求因网络失败而未收到响应
- **THEN** 表单内联展示「网络连接失败，请检查网络」，不重复弹全局 toast

#### Scenario: 会话过期仍在系统内页生效

- **WHEN** 已登录用户在系统内页遇到 401
- **THEN** 全局「登录已过期」处理照常触发

#### Scenario: 其他请求的全局提示不受影响

- **WHEN** 系统内页的普通请求返回非 200
- **THEN** 全局 toast 照常弹出（豁免仅对显式声明 `skipGlobalError` 的请求生效）

### Requirement: 登录表单支持密码管理器与键盘操作

两端登录表单 MUST 提供 `autocomplete="username"` 与 `autocomplete="current-password"`，两个输入框 MUST 都支持 Enter 提交。桌面端 MUST 额外做到：用户名框自动聚焦、密码框提示大小写锁定状态。

#### Scenario: 密码管理器填充

- **WHEN** 用户使用浏览器或第三方密码管理器
- **THEN** 凭据可被正确填充到对应输入框

#### Scenario: 键盘提交

- **WHEN** 焦点在用户名输入框时按 Enter
- **THEN** 触发登录提交

#### Scenario: 桌面端自动聚焦

- **WHEN** 桌面端登录页加载完成
- **THEN** 用户名输入框已获得焦点，用户可直接键入

#### Scenario: 桌面端大小写锁定提示

- **WHEN** 焦点在密码框且键盘处于 CapsLock 开启状态
- **THEN** 提示大小写已锁定

#### Scenario: 移动端不自动聚焦

- **WHEN** 移动端登录页加载完成
- **THEN** 不自动聚焦输入框——避免一进页面就弹出软键盘遮挡大半个屏幕

> **范围说明：** 自动聚焦与大小写锁定是**桌面端专属**。移动端自动聚焦会立即唤起软键盘，
> 遮挡品牌区与表单，属体验倒退；大小写锁定依赖 `KeyboardEvent.getModifierState`，
> 而软键盘不产生该状态（`FocusEvent` 亦无此方法），故移动端不实现这两项。
