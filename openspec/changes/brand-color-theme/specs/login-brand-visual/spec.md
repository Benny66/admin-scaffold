# Spec: login-brand-visual

登录页的品牌化视觉体系：可配置背景图、左图右表单分栏、品牌块、两端共享的 footer，以及配套的加载降级、响应式折叠与登录表单 UX 规范。

## MODIFIED Requirements

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

### Requirement: 桌面端登录页采用左图右表单分栏布局

桌面端（`frontend/src/views/Login.vue`）MUST 采用分栏布局：左栏为背景展示区（背景层 + scrim + 品牌块），右栏为登录表单区（含 footer）；视口宽度 < 900px 时 MUST 折叠为单列（左栏隐藏，背景铺满，表单卡片居中）。

左栏渐变背景的颜色源 MUST 使用 CSS 变量 `--brand-from` 与 `--brand-to`（由 `brand-color-extract` 能力注入），禁止硬编码色值。scrim 的半透明色 MUST 从同一变量派生。

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
- **THEN** CSS 变量使用定义时的 fallback 值（`#1f6feb` / `#0d3b8c`）
