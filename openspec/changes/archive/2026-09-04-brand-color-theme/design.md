## Context

`login-brand-visual` 已把登录页品牌化做到 35/39（分栏布局、logo 回退、背景图配置、footer 复用组件等），但品牌色仍是硬编码蓝色（`#1f6feb → #0d3b8c`），与用户配置的 logo 毫无关联。一个换橙色 logo 的系统仍是蓝色登录页、蓝色按钮，品牌感断裂。同时用户希望登录页分栏后 footer 只落在右栏底部，不跨全宽。

本 change 解决两件事：
1. 从 logo 自动提取主色，派生 Element Plus 全站主题变量族 + 登录页渐变变量，注入 `:root`。
2. 登录页 footer 从全宽底部条移到右栏底部。

约束：
- 不引入 Go 图像处理依赖（基座干净性）。
- 不新增后端字段（logo 已通过 `GetSystemInfo` 返回）。
- 提取失败必须回退当前蓝色，基座零配置开箱即用。
- 桌面端与移动端共享提取逻辑。

## Goals / Non-Goals

**Goals:**
- 从 logo 自动提取主色，派生 Element Plus 主题变量族并注入 `:root`，全站按钮/链接/选中态跟随。
- 登录页渐变背景跟随 logo 主色。
- 提取结果按 logo URL 缓存到 localStorage，换 logo 自动失效重算。
- 提取失败回退默认蓝色 `#1f6feb → #0d3b8c`。
- 登录页 footer 移到右栏底部，窄屏单列仍在页面底部。

**Non-Goals:**
- 不做后端图像处理或预计算（纯前端方案）。
- 不做多色提取或完整调色板（只提取一个主色）。
- 不做暗色模式 / 主题切换器（仅亮色主题，跟随 logo）。
- 不做管理员上传 logo 的交互流程（logo 配置仍走 config.yaml）。
- 不做移动端 footer 位置改动（移动端分段布局已合理，footer 仍在页面底部）。

## Decisions

### D1: 纯前端 Canvas 提取，不做后端预计算

**选择**：前端 Canvas 读取 logo 像素提取主色。

**理由**：
- 基座后端目前不碰图像逻辑，加 Go `image` 处理是额外认知负担。
- logo 走 `/static/logo.png`，与前端同源，Canvas 不会被 taint。
- 提取一次缓存到 localStorage，后续加载命中缓存，无重复计算成本。
- Go stdlib 虽能做，但每次 logo 变更都要重启服务才重算；前端方案换 logo 即时生效。

**备选**：后端启动时读取 logo 文件预计算，存 config 字段返回。被否，原因如上。

### D2: 提取算法——色相分桶取众数

**选择**：缩放到 64×64 → 跳过透明/低饱和度/极暗极亮像素 → 按色相分桶（36 桶 × 10°）→ 取像素最多的桶 → 桶内 RGB 平均。

**理由**：
- 跳过透明像素：logo 常有透明背景，不跳会把背景色算进去。
- 跳过低饱和度（S < 0.15）与极暗极亮（L < 0.1 / > 0.9）：避免把白/灰/黑当作品牌色（logo 常有黑色文字或白色描边）。
- 色相分桶取众数：比"取平均色"更准——双色 logo 不会变成中间色，而是取面积大的那个。
- 缩放到 64×64：4096 像素遍历 < 1ms，用户无感。

**备选 A**：缩小到 1×1 取平均色。被否——双色 logo 会变成中间泥色。
**备选 B**：K-means 聚类。被否——对基座功能来说过度工程，且 K-means 的 K 值选择本身就是个问题。
**备选 C**：中位切分（median cut）。被否——算法虽成熟但实现复杂度高，收益对本场景不明显。

### D3: 派生 Element Plus 变量族——RGB 线性混合

**选择**：`mix(color, target, weight) = color × (1 − weight) + target × weight`，在 RGB 线性空间完成。

**理由**：
- Element Plus 的 SCSS 源码就是这么生成 light-N / dark-N 变量的（`mix($color, $white, $weight)`），保持一致避免视觉偏差。
- RGB 线性混合虽然不如 HSL 空间混合在视觉过渡上平滑，但与 EP 原生行为一致，避免组件在 hover/active 态出现意外色偏。

**备选**：HSL 空间调整明度派生。被否——与 EP 原生派生方式不一致，hover/active 态可能偏色。

### D4: 亮色主色的明度钳制

**选择**：提取的主色明度 L > 0.6 时，在派生前钳制到 0.5。

**理由**：
- 如果 logo 主色是浅黄（L=0.85），派生的 `light-9` = mix(primary, white, 90%) 几乎是纯白，Element Plus 的浅色背景（如 tag 默认背景、disabled 态）会看不见。
- 钳制到 0.5 保证 light 变体仍有可见色差。
- 0.5 是经验值，大多数品牌色明度在 0.4-0.6 区间，钳制不影响正常 logo。

### D5: localStorage 缓存 key 包含 logo URL

**选择**：缓存 key 为 `brand_theme_<logoUrl>`，直接用 URL 字符串（不哈希，因为 logo URL 通常很短如 `/static/logo.png`）。

**理由**：
- URL 变化（换 logo 文件名）→ key 变化 → 自动失效重算。
- 不用哈希是因为 URL 本身就是合法的 localStorage key 后缀，且短到无性能问题。
- 不用 logo 文件内容的 hash 是因为读取文件内容算 hash 本身就要把图加载完，不如直接做提取。

### D6: 注入时机——App.vue onMounted

**选择**：在 `App.vue` 的 `onMounted` 中调用 `appStore.applyBrandTheme(logoUrl)`，紧跟在 `fetchSystemInfo` 之后。

**理由**：
- `fetchSystemInfo` 已拿到 logo URL，紧接着提取主色是自然的流水线。
- `onMounted` 保证 DOM 就绪，`document.documentElement.style.setProperty` 可用。
- 注入发生在 Vue 组件树渲染前（onMounted 是同步的，但 CSS 变量注入是同步生效的），首屏即可见正确主题色。
- 即使有轻微闪烁（fetchSystemInfo 网络延迟），登录页在 fetchSystemInfo 完成前已有 loading 态，不是问题。

### D7: footer 移到右栏内部——纯 CSS 调整

**选择**：`Login.vue` 的 `<footer>` 从 `.login-page` 直接子节点移入 `.login-form-area`，右栏改为 `flex-direction: column`，表单卡片区 `flex: 1`，footer 自然落底。

**理由**：
- 纯 CSS + 模板结构调整，无 JS 逻辑。
- 窄屏单列模式下 `.login-form-area` 仍是 flex column，footer 仍在底部，行为不变。
- 左栏不受影响，视觉重心更聚拢在右栏。

### D8: 回退以「不注入」实现，而非注入一组默认值

**选择**：无 logo / 提取失败 / 任一异常时，`applyBrandTheme` 调 `clearThemeVars()` 撤销已注入的变量后直接返回；默认外观由各端 CSS 的 `var(--x, <默认值>)` fallback 提供。

**理由**：
- 桌面端与移动端登录页的渐变起点本来就不同色（`#1f6feb` vs `#1989fa`）。初版按 spec 字面实现「回退时注入 `--brand-from: #1f6feb`」，会让移动端在「无 logo」这一**默认配置**下发生视觉变化。
- 同一问题也影响 Element Plus 主色：全仓 grep `--el-color-primary` 零命中，说明基座从未覆盖过 EP 主色，现全站按钮/链接/选中态是 EP 默认的 `#409eff`。注入 `#1f6feb` 会把它一并改掉。
- 这两处变化都违背 proposal 的「无破坏性：提取失败回退现有蓝色，视觉表现与现状一致」，且会让验收项「默认 logo 登录页渐变与按钮色无变化」无法成立。
- 交给各端自己的 CSS fallback 后，回退结果与改动前**逐像素一致**，且省掉一份需要额外维护的默认值注入逻辑。

**备选**：回退时注入 `deriveThemeVars(#1f6feb)`。被否，原因如上——它会把两端的渐变起点与 EP 主色都拉到桌面端的值。

## Risks / Trade-offs

- **[logo 是 SVG]** → Canvas 可以画 SVG，但 SVG 的 `foreignObject` 或外部资源可能 taint canvas。基座默认 logo 是 PNG，问题不大；用户换 SVG 时若遇 taint，提取失败回退蓝色，不会报错。 → 文档提示推荐用 PNG 或纯色 SVG。
- **[Element Plus 版本升级改变变变量名]** → 当前用 `--el-color-primary-*`，EP 4.x 不会动这个命名体系。 → 低风险，升级时测一下即可。
- **[提取的主色与 EP 组件视觉不协调]** → 某些饱和度极高的主色（如纯红 `#ff0000`）在 Element Plus 的 button/link/tag 上可能很刺眼。 → 不做额外钳制，用户选什么 logo 就是什么色，过度干预反而违背"跟随 logo"的初衷。
- **[首次加载有短暂蓝色闪烁]** → fetchSystemInfo 网络延迟期间 CSS 变量是 fallback 蓝色，注入后切到 logo 色。 → 登录页已有 loading 态，可接受；系统内页首屏已有内容，切换不明显。
- **[localStorage 被禁用]** → 无痕模式下 localStorage 可能抛错。 → applyBrandTheme 内 try-catch，失败回退蓝色，不影响正常使用。
