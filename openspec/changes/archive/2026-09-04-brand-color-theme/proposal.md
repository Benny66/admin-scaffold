## Why

登录页与全站主题色目前是硬编码的蓝色（`#1f6feb → #0d3b8c`），与用户配置的 logo 毫无关系。一个换了橙色 logo 的系统仍然是蓝色登录页、蓝色按钮、蓝色链接，品牌感断裂。同时 `login-brand-visual` 已把 footer 做成全宽条，但用户希望登录页分栏后 footer 只落在右栏底部，视觉重心更聚拢。

## What Changes

**全站主题色从 logo 自动派生**

- 新增前端颜色提取工具：用 Canvas 读取 logo 像素，按色相分桶取众数提取主色，派生 Element Plus 主题变量族（`--el-color-primary` 及 light-3/5/7/8/9、dark-2）与登录页渐变变量（`--brand-from` / `--brand-to`）。

- 提取结果按 logo URL 缓存到 localStorage，换 logo 自动失效重算。

- 在 `App.vue` 启动时调用 `appStore.applyBrandTheme(logoUrl)`，通过 `document.documentElement.style.setProperty` 注入 `:root`，Element Plus 全站按钮 / 链接 / 选中态自动跟随。

- 提取失败（无 logo、纯透明、低饱和度）回退当前蓝色 `#1f6feb → #0d3b8c`，基座零配置开箱即用。

- 桌面端与移动端共享同一份提取逻辑与派生算法。

**登录页 footer 移到右栏底部**

- `frontend/src/views/Login.vue` 的 `<footer>` 从 `.login-page` 直接子节点移入 `.login-form-area`，右栏改为 `flex-direction: column`，表单区 `flex: 1`，footer 自然落底。

- 窄屏单列模式下 footer 仍在页面底部，行为不变。

## Capabilities

### New Capabilities

- `brand-color-extract`: 从 logo 图片提取主色并派生主题变量族的能力——Canvas 像素采样、色相分桶取众数、饱和度/亮度过滤、Element Plus 变体派生、localStorage 缓存、失败回退。

### Modified Capabilities

- `login-brand-visual`: 登录页 footer 位置从全宽底部条改为右栏底部；登录页渐变色源从硬编码改为 `--brand-from` / `--brand-to` 变量（由 `brand-color-extract` 注入）。

## Impact

- 修改文件：

  - `frontend/src/utils/colorExtract.js`（新增）、`frontend/src/stores/app.js`（加 `applyBrandTheme`）、`frontend/src/App.vue`（启动调用）、`frontend/src/views/Login.vue`（footer 位置 + 渐变变量源）

  - `mobile/src/utils/colorExtract.js`（新增）、`mobile/src/stores/app.js`（同上）、`mobile/src/App.vue`（同上）、`mobile/src/views/Login.vue`（渐变变量源）

- 无新第三方依赖（Canvas / localStorage / CSS 变量均为浏览器原生 API）。

- 无后端改动（logo 已通过 `GetSystemInfo` 返回，颜色提取纯前端）。

- 无破坏性：提取失败回退现有蓝色，视觉表现与现状一致。

