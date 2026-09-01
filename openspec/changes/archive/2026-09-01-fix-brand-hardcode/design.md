# Design: fix-brand-hardcode

## Context

脚手架在 `frontend/CLAUDE.md` 立规「系统名称 MUST 来自 appStore.systemName，禁止硬编码『企业管理系统』」，但 `frontend/index.html:7` 与 `mobile/index.html:6` 的 `<title>` 就是硬编码的「企业管理系统」——自相矛盾。`scripts/init.sh` 只做 ASCII 替换（`base-backend` → 项目名），不碰中文品牌，导致 `make init` 后浏览器标签页仍显示「企业管理系统」。

品牌名称的运行时机制（brand-config）已就位：`/api/system/info` 返回 `name`，前端 `fetchSystemInfo` 存入 `systemName`，`router.beforeEach` 已用 `systemName` 设 `document.title`。问题在于 index.html 的 `<title>` 是**编译期硬编码**，运行时虽被覆盖，但初始闪现 + init 未替换残留中文。

## Goals / Non-Goals

**Goals:**

1. `frontend/index.html`、`mobile/index.html` 的 `<title>` 改中性占位，消除硬编码中文品牌。
2. `scripts/init.sh` 新增 `--app-name` 参数，把「企业管理系统」中文品牌纳入替换，覆盖 index.html 标题 + 其他中文残留。
3. 消解 CLAUDE.md 与 index.html 的自相矛盾。

**Non-Goals:**

- 不改 brand-config 的运行时机制（已就位，本 change 只补 init 的编译期替换闭环）。
- 不做 favicon/logo 的中文化替换（那是 brand-config 的运行时职责，且 logo 是二进制文件，init 文本替换不适用）。

## Decisions

### D1：index.html title 改为中性占位 `Base Admin`

两个 index.html 的 `<title>` 从「企业管理系统」改为 `Base Admin`（中性，业务无关）。

**Why：** 中性占位本身不泄漏品牌、不违反 CLAUDE.md 的「禁止硬编码企业管理系统」；运行时由 `fetchSystemInfo` + `document.title` 覆盖为真实品牌名。

**Alternatives considered：**
- 保留空 `<title>` → 拒绝，无标题时浏览器显示 URL，体验差。
- 保留「企业管理系统」→ 拒绝，正是要消除的硬编码。

### D2：init.sh 新增 `--app-name` 参数，替换中文品牌

`init.sh` 新增 `--app-name <名>`，把「企业管理系统」替换为该值。缺省时（未传 `--app-name`）不替换中文品牌，仅把 index.html 占位保持中性。

**Why：** 中文品牌替换是「新项目名」的语义，应显式传入而非猜；缺省不替换保持向后兼容（避免误改用户已自定义的中文）。

**Alternatives considered：**
- 默认用项目名替换中文 → 拒绝，项目名是英文/ASCII，中文品牌名可能不同（如「我的系统」vs `my-system`），不能臆断。
- 硬编码替换「企业管理系统」为固定新名 → 拒绝，同样无法通用。

### D3：消解 CLAUDE.md 矛盾

`frontend/CLAUDE.md` 第 38 条的表述保持（「禁止硬编码企业管理系统」），并补一句「index.html 的 title 为中性占位，运行时由 brand-config 覆盖」。

**Why：** 让规矩与事实一致——index.html 不再有「企业管理系统」，CLAUDE.md 的禁令成立。

## Risks / Trade-offs

- [`--app-name` 替换「企业管理系统」可能漏掉个别变体（如「企业管理」短写）] → init.sh 的替换以精确全串「企业管理系统」为准，变体由使用者在 init 后自行查漏（`grep 企业`）。
- [运行时 title 覆盖前，index.html 占位会短暂闪现「Base Admin」] → 可接受（中性占位无品牌泄漏，且 SPA 首屏挂载很快）。
- [mobile 端 `document.title` 覆盖逻辑若缺失，移动端标签页仍是占位] → 本 change 需确认 mobile 端也执行 title 覆盖（见 tasks）。

## Open Questions

- 移动端目前是否在 `fetchSystemInfo` 后设置 `document.title`？前端 router 有，移动端需核对；若缺失，本 change 一并补上。
