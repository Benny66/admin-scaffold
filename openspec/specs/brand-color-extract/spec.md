# Spec: brand-color-extract

从 logo 图片提取主色并派生全站主题变量族的能力：Canvas 像素采样、色相分桶取众数、饱和度/亮度过滤、Element Plus 变体派生、localStorage 缓存、失败回退。

## Requirements

### Requirement: 从 logo 提取主色

系统 MUST 提供前端工具函数 `extractDominantColor(logoUrl)`，通过 Canvas 读取 logo 像素并提取主色。提取算法 MUST 满足：
- 将 logo 缩放到 ≤ 64×64 的 canvas 上降低计算量。
- 遍历像素时 MUST 跳过透明像素（alpha < 128）。
- 像素转 HSL 后 MUST 跳过低饱和度（S < 0.15）、极暗（L < 0.1）与极亮（L > 0.9）的像素，避免把白/灰/黑当作品牌色。
- 剩余像素按色相分桶（每 10° 一桶，共 36 桶），取像素数最多的桶，桶内 RGB 平均作为主色。
- 所有像素都被跳过时 MUST 返回 null，表示提取失败。
- logo 图片加载失败时 MUST 返回 null。

#### Scenario: 提取到主色

- **WHEN** logo 图片可正常加载且含非透明、非灰像素
- **THEN** 返回主色的 hex 字符串（如 `#3b82f6`）

#### Scenario: 透明背景 logo

- **WHEN** logo 是透明背景 PNG，主体仅占少量像素
- **THEN** 透明像素被跳过，从主体像素中提取主色，不把透明背景色算入

#### Scenario: 纯灰 logo

- **WHEN** logo 仅含灰/白/黑像素（所有像素饱和度 < 0.15）
- **THEN** 返回 null

#### Scenario: logo 加载失败

- **WHEN** logo URL 指向不存在的文件
- **THEN** 返回 null，不抛异常

### Requirement: 派生 Element Plus 主题变量族

系统 MUST 提供函数 `deriveThemeVars(primary)`，从主色 hex 派生 Element Plus 所需的 CSS 变量族。MUST 生成以下变量：
- `--el-color-primary`：主色本身。
- `--el-color-primary-light-3`：主色与白色混合，白色占 30%。
- `--el-color-primary-light-5`：主色与白色混合，白色占 50%。
- `--el-color-primary-light-7`：主色与白色混合，白色占 70%。
- `--el-color-primary-light-8`：主色与白色混合，白色占 80%。
- `--el-color-primary-light-9`：主色与白色混合，白色占 90%。
- `--el-color-primary-dark-2`：主色与黑色混合，黑色占 20%。
- `--brand-from`：主色本身（登录页渐变起点）。
- `--brand-to`：主色与黑色混合，黑色占 30%（登录页渐变终点）。
- `--brand-scrim-from`：主色加 0.72 透明度（登录页品牌区遮罩起点）。
- `--brand-scrim-to`：`--brand-to` 的颜色加 0.85 透明度（遮罩终点）。

混合 MUST 在 RGB 线性空间完成：`result = color × (1 − weight) + target × weight`。

> **口径说明：** 上述权重与 Element Plus 的 SCSS 派生保持一致（`mix(white, $color, N)`），
> 可用 EP 默认主色 `#409eff` 自检：light-3 应得 `#79bbff`、dark-2 应得 `#337ecc`。
> 早期草稿把 light-3 写作「主色与白色按 3:7 混合」，按字面（主色占 3）实现会得到 70% 白，
> 与 EP 原生值不符，且 dark-2 会暗到接近全黑，故以 EP 原生派生为准。

#### Scenario: 生成完整变量族

- **WHEN** 传入主色 `#3b82f6`
- **THEN** 返回包含上述全部键值对的对象

#### Scenario: 亮色 logo 的主色被钳制

- **WHEN** 提取的主色明度过高（L > 0.6）
- **THEN** 在派生前将其明度钳制到 0.5，避免 light-9 变体几乎不可见

### Requirement: 注入主题变量到 :root

系统 MUST 提供函数 `applyThemeVars(vars)`，将 `deriveThemeVars` 的输出通过 `document.documentElement.style.setProperty` 注入到 `:root`，使 Element Plus 全站组件跟随变化。注入由 store 的 `applyBrandTheme` 在 `fetchSystemInfo` 内部触发。

系统 MUST 同时提供函数 `clearThemeVars()`，用 `document.documentElement.style.removeProperty` 撤销已注入的全部变量。其键集合 MUST 从 `deriveThemeVars` 现场取得，不得另维护一份常量清单（避免与派生输出漂移）。

#### Scenario: 全站按钮变色

- **WHEN** 主色被注入 `:root`
- **THEN** Element Plus 按钮的默认色、链接色、选中态色跟随主色变化

#### Scenario: 登录页渐变跟随主色

- **WHEN** `--brand-from` 与 `--brand-to` 被注入
- **THEN** 登录页左栏渐变背景使用这两个变量值

### Requirement: 提取结果按 logo URL 缓存

系统 MUST 将提取结果缓存到 localStorage，缓存 key MUST 包含 logo URL（或其哈希），使得更换 logo 后缓存自动失效并重新提取。

缓存 key MUST 同时包含结构版本号，使得 `deriveThemeVars` 的输出结构或算法变化后旧缓存整体失效。

#### Scenario: 首次提取后缓存

- **WHEN** 首次对某 logo URL 提取主色
- **THEN** 提取完成后结果写入 localStorage

#### Scenario: 二次加载命中缓存

- **WHEN** 同一 logo URL 再次加载
- **THEN** 直接从 localStorage 读取缓存结果，不重复执行 Canvas 提取

#### Scenario: 换 logo 后缓存失效

- **WHEN** logo URL 变化（配置改了 logo 文件名）
- **THEN** localStorage 旧 key 未命中，重新执行 Canvas 提取并写入新 key

#### Scenario: 缓存读写不可用时降级

- **WHEN** localStorage 被禁用（无痕模式）或已写满
- **THEN** 读写异常被静默吞掉，改为每次重新提取，主题注入不受影响

### Requirement: 提取失败时回退默认蓝色

当 `extractDominantColor` 返回 null、`logo` 字段为空、或 `applyBrandTheme` 抛出任何异常时，系统 MUST 回退到默认蓝色主题。

回退 MUST 以「不注入」的方式实现：`applyBrandTheme` 调用 `clearThemeVars()` 撤销已注入的变量后直接返回，**不得**注入一组默认值。默认外观由消费方 CSS 的 `var(--x, <默认值>)` fallback 兜底：

- 桌面端 `--brand-from` fallback `#1f6feb`、`--brand-to` fallback `#0d3b8c`
- 移动端 `--brand-from` fallback `#1989fa`、`--brand-to` fallback `#0d3b8c`
- Element Plus 主色未注入时沿用 EP 自带的 `#409eff`

**理由**：桌面端与移动端登录页的渐变起点本来就不同色（`#1f6feb` vs `#1989fa`）。若回退时统一注入 `#1f6feb`，移动端会在「无 logo」这一默认配置下发生视觉变化，「无破坏性 / 零配置开箱即用」目标即不成立；同理，注入 `--el-color-primary: #1f6feb` 会把 EP 默认的 `#409eff` 按钮色一并改掉。交给各端自己的 CSS fallback，回退结果才与现状逐像素一致。

#### Scenario: 无 logo 时回退

- **WHEN** `GetSystemInfo` 返回空 `logo` 字段
- **THEN** 不执行 Canvas 提取，且不注入任何主题变量；各端 CSS fallback 生效，视觉与改动前完全一致

#### Scenario: 提取失败时回退

- **WHEN** logo 存在但提取结果为 null（纯灰、加载失败）
- **THEN** 同上不注入，全站呈现各端默认蓝色

#### Scenario: 已注入后回退

- **WHEN** 先前已注入品牌形象变量，随后 `logo` 被移除或提取失败
- **THEN** 已注入的变量被 `clearThemeVars()` 撤销，回到未注入状态，而非残留旧品牌色

### Requirement: 桌面端与移动端共享提取逻辑

`frontend/src/utils/colorExtract.js` 与 `mobile/src/utils/colorExtract.js` MUST 内容逐字节相同，提供相同的导出接口（`extractDominantColor`、`deriveThemeVars`、`applyThemeVars`、`clearThemeVars`、`DEFAULT_PRIMARY`、`DEFAULT_BRAND_TO`），两端 store 的 `applyBrandTheme` action 行为一致。

该一致性 MUST 由 `backend/internal/guard/color_extract_test.go` 编译成会失败的静态检查：不仅比对导出集合，还 MUST 比对两端文件是否逐字节相同，以挡住「只改了一端算法」的漂移。

#### Scenario: 两端接口一致

- **WHEN** 对比两端 colorExtract 模块的导出签名
- **THEN** 函数名与参数签名完全一致

#### Scenario: 一端算法被单独修改

- **WHEN** 只修改了 `frontend/src/utils/colorExtract.js` 而未同步 `mobile/src/utils/colorExtract.js`
- **THEN** `make test` 中的 guard 测试失败并指出两端文件内容不一致
