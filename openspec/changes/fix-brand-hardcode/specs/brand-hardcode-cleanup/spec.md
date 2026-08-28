# Spec: brand-hardcode-cleanup

消除硬编码品牌残留，使 `init.sh` 的品牌改名闭环覆盖浏览器标题，占位与运行时品牌解耦。

## ADDED Requirements

### Requirement: 浏览器标题不硬编码品牌

`frontend/index.html` 与 `mobile/index.html` 的 `<title>` MUST 使用中性占位（如 `Base Admin`），禁止硬编码「企业管理系统」。

#### Scenario: 检查 index.html 标题

- **WHEN** 查看 `frontend/index.html` 与 `mobile/index.html` 的 `<title>` 标签
- **THEN** 内容为中性占位，不包含「企业管理系统」

#### Scenario: 运行时标题被品牌名覆盖

- **WHEN** 前端加载并成功拉取系统信息（`/api/system/info` 返回品牌 name）
- **THEN** 浏览器标签页标题更新为该品牌名（非占位）

### Requirement: init.sh 支持中文品牌替换

`scripts/init.sh` MUST 支持 `--app-name <名>` 参数，把「企业管理系统」替换为该值；缺省时不替换中文品牌。

#### Scenario: 带 app-name 初始化

- **WHEN** 执行 `scripts/init.sh my-system --app-name 我的系统`
- **THEN** 代码中（含 index.html 的占位）不再出现「企业管理系统」，替换为「我的系统」

#### Scenario: 不带 app-name 初始化

- **WHEN** 执行 `scripts/init.sh my-system`（未传 --app-name）
- **THEN** 中文品牌保持中性占位，不做臆断替换

### Requirement: 文档与代码一致

`frontend/CLAUDE.md` 关于「禁止硬编码企业管理系统」的表述 MUST 与实际代码一致，即 index.html 无该硬编码。

#### Scenario: 校验一致性

- **WHEN** 全局搜索「企业管理系统」
- **THEN** 代码文件（go/vue/js/html）中无该硬编码（仅可能出现在 init.sh 的替换逻辑与文档说明中）
