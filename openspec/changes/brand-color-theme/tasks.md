## 1. colorExtract 工具模块（桌面端）

- [x] 1.1 创建 `frontend/src/utils/colorExtract.js`，实现 `rgbToHsl(rgb)` / `hslToRgb(h, s, l)` 转换函数
- [x] 1.2 实现 `mixHex(color, target, weight)` —— RGB 线性空间混合：`color × (1-w) + target × w`，返回 hex
- [x] 1.3 实现 `extractDominantColor(logoUrl)` —— 加载 logo 到 Image，缩放到 64×64 canvas，遍历像素跳过透明/低饱和度/极暗极亮，色相分桶取众数，桶内 RGB 平均；全跳过返回 null
- [x] 1.4 实现 `deriveThemeVars(primaryHex)` —— 生成 9 个 CSS 变量键值对（`--el-color-primary` 及 light-3/5/7/8/9、dark-2、`--brand-from`、`--brand-to`）；主色 L > 0.6 时钳制到 0.5
- [x] 1.5 实现 `applyThemeVars(vars)` —— 遍历对象 `document.documentElement.style.setProperty(key, value)` 注入 :root
- [x] 1.6 默认蓝色常量 `DEFAULT_PRIMARY = '#1f6feb'` 与 `DEFAULT_BRAND_TO = '#0d3b8c'`，供回退路径使用

## 2. colorExtract 工具模块（移动端）

- [x] 2.1 创建 `mobile/src/utils/colorExtract.js`，与桌面端完全一致（同接口、同算法）
- [x] 2.2 确认两端导出签名一致：`extractDominantColor`、`deriveThemeVars`、`applyThemeVars`、`DEFAULT_PRIMARY`、`DEFAULT_BRAND_TO`

## 3. appStore 注入逻辑（桌面端）

- [x] 3.1 在 `frontend/src/stores/app.js` 新增 action `applyBrandTheme(logoUrl)`，流程：无 logo → clearThemeVars() 不注入（回退 = 不注入，由 CSS fallback 兜底，见 design D8）；有 logo → 查 localStorage `brand_theme_v1_<logoUrl>`，命中则 applyThemeVars(缓存)；未命中则 extractDominantColor，null → clearThemeVars() 回退，非 null → deriveThemeVars → applyThemeVars → 写 localStorage
- [x] 3.2 localStorage 读写 try-catch 包裹（无痕模式禁用 localStorage 时静默回退蓝色）
- [x] 3.3 `fetchSystemInfo` 在拿到 systemInfo 后调用 `applyBrandTheme(systemInfo.logo)`（紧接在现有 state 赋值之后）

## 4. appStore 注入逻辑（移动端）

- [x] 4.1 在 `mobile/src/stores/app.js` 新增 `applyBrandTheme(logoUrl)`，行为与桌面端一致
- [x] 4.2 `fetchSystemInfo` 在拿到 systemInfo 后调用 `applyBrandTheme(systemInfo.logo)`

## 5. Login.vue 渐变色源改用 CSS 变量

- [x] 5.1 `frontend/src/views/Login.vue` 的 `<style>` 中 `.login-page` 渐变 `--brand-from` / `--brand-to` 硬编码值改为 CSS 变量 fallback 写法：`var(--brand-from, #1f6feb)` / `var(--brand-to, #0d3b8c)`
- [x] 5.2 `.brand-left .scrim` 的 rgba 色源改用 `--brand-from` / `--brand-to` 派生（或直接用同一变量加透明度，但 rgba 不能直接用 CSS 变量，需用 `color-mix` 或预派生 scrim 色作为额外变量 `--brand-scrim-from` / `--brand-scrim-to` 注入）
- [x] 5.3 若新增 `--brand-scrim-from` / `--brand-scrim-to`，在 `deriveThemeVars` 中一并生成（主色 + 透明度 0.72 / 0.85）并注入
- [x] 5.4 `mobile/src/views/Login.vue` 同步把渐变 / scrim 色源改为 CSS 变量 fallback

## 6. Login.vue footer 移到右栏

- [x] 6.1 `frontend/src/views/Login.vue` 模板中把 `<footer class="login-footer">` 从 `.login-page` 直接子节点移入 `.login-form-area` 内部
- [x] 6.2 `.login-form-area` 的 CSS 改为 `display: flex; flex-direction: column`
- [x] 6.3 `.login-card`（或表单卡片的容器）加 `flex: 1` 撑满上方
- [ ] 6.4 验证窄屏单列模式下 footer 仍在页面底部，不被分栏折叠影响

## 7. App.vue 启动注入

- [x] 7.1 `frontend/src/App.vue` 确认 `onMounted` 中已调用 `appStore.fetchSystemInfo()`，且 `applyBrandTheme` 在 `fetchSystemInfo` 内部已触发（无需额外调用，或在此显式调用确认）
- [x] 7.2 `mobile/src/App.vue` 同步确认 `applyBrandTheme` 随 `fetchSystemInfo` 触发

## 8. guard 测试扩展

- [x] 8.1 `internal/guard/frontend_store_test.go` 确认扫描范围包含 `miniapp/src/stores/` 与新增的 `colorExtract` 引用解析（若 miniapp 端也接入则扫，当前仅 frontend/mobile）
- [x] 8.2 现有 `brand_config_guard` 不受影响（本 change 不新增 brand-config 字段，纯前端方案）
- [x] 8.3 新增或扩展 guard 测试：验证 `frontend/src/utils/colorExtract.js` 与 `mobile/src/utils/colorExtract.js` 导出接口一致（函数名集合相同）——可选，若 guard 框架支持文件级断言
- [x] 8.4 `make test` 全绿

## 9. 整体验收

- [x] 9.1 `make test` 全绿（含 guard 测试）
- [x] 9.2 `make lint` 全绿（四端 ESLint + backend vet）
- [x] 9.3 `make smoke` 全绿
- [ ] 9.4 手测：默认 logo（蓝色系）登录页渐变与按钮色无变化（回退路径生效）
- [ ] 9.5 手测：换一个橙色 logo，登录页渐变变橙、Element Plus 按钮变橙、全站链接变橙
- [ ] 9.6 手测：删除 logo 配置（空 logo），全站回退蓝色
- [ ] 9.7 手测：换纯灰 logo（无饱和度），全站回退蓝色
- [ ] 9.8 手测：桌面端登录页 footer 在右栏底部，窄屏单列在页面底部
- [ ] 9.9 手测：移动端登录页渐变跟随 logo 色
- [ ] 9.10 手测：localStorage 命中缓存（二次加载不触发 Canvas 提取，可通过 Network/Performance 面板观察）
