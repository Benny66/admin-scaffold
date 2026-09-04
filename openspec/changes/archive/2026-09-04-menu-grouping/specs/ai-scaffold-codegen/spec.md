# Spec: ai-scaffold-codegen

本 delta 覆盖 `menu-grouping` 对 `ai-scaffold-codegen` 的修改：`make gen` 的前端路由注入从「注入平铺路由条目」改为「注入分组/归入已有分组」。

## MODIFIED Requirements

### Requirement: 模块代码生成器

项目 MUST 提供 `scripts/gen-module.sh <name> [group=<分组 path>]`，能从 `_example/` 生成新模块的完整骨架，且**三端全闭环**：

- 后端：`models/` + `services/` + `controllers/`
- 后端注入：`router.go` 路由注册、`database.go` 的 `AutoMigrate`、`database.go` 的 `initBaseData` 权限码注册
- 前端：`views/<name>/index.vue`、`api/index.js` 追加 API 定义、`router/index.js` 菜单路由注册（两级分组结构，见 `menu-grouping`）

前端路由注入 MUST 遵循分组语义：显式指定已存在的分组时注入该分组 `children`；未指定 `group` 或分组不存在时，以模块复数名为分组 path 新建分组（首叶子 `path: ''`）。所有生成文件 MUST 带明确的 `// TODO: 业务逻辑` 锚点。

#### Scenario: AI 用生成器创建新模块

- **WHEN** AI 执行 `make gen name=asset`
- **THEN** 生成器产出 `asset` 模块的后端三层、路由注册、AutoMigrate 注册、权限码注册、前端页面、API 定义与前端分组菜单（分组 path `assets`、首叶子 `path: ''`），且 `make test` 与 `make smoke` 均通过

#### Scenario: 生成器产出带业务锚点

- **WHEN** 生成完成后打开任一生成文件
- **THEN** 文件中存在明确的 `// TODO: 业务逻辑` 注释，标记 AI 需要手写的部分

#### Scenario: 生成器自动注册权限码

- **WHEN** AI 执行 `make gen name=asset` 后查看 `database.go`
- **THEN** `initBaseData` 的权限声明块中新增 `assets:view` / `assets:create` / `assets:edit` / `assets:delete` 四条记录

#### Scenario: 生成器注入已有分组

- **WHEN** AI 执行 `make gen name=asset_category group=asset` 且 `path: 'asset'` 分组已存在
- **THEN** 新叶子注入该分组 `children`，URL 为 `/asset/asset_category`，既有叶子不受影响，`make lint` 通过

#### Scenario: 生成器自建分组

- **WHEN** AI 执行 `make gen name=asset`（不带 `group`）或 `group=asset` 的分组不存在
- **THEN** 新建分组（path `assets`、首叶子 `path: ''`），URL 为 `/assets`，与改动前的寻址一致

#### Scenario: 生成后目标已存在

- **WHEN** 同名模块文件已存在时执行 `make gen`
- **THEN** 生成器报错退出且不覆盖任何已有文件

#### Scenario: 资源名已被占用时拒绝生成

- **WHEN** 待生成模块的复数路由路径或权限码前缀已被占用（如 `box` 与 `boxe` 的复数同为 `boxes`，或模块名与存量五件套撞车）
- **THEN** 生成器在任何文件生成之前报错退出（fail fast），不留下「后端已生成、前端未注入」的不一致状态

### Requirement: 生成器完成提示与实际行为一致

生成器的完成提示 MUST 准确反映「哪些步骤已自动化、哪些仍需手工」，且 MUST NOT 指示开发者修改已被护栏禁止改动的既有约定。提示 MUST 输出新模块的最终菜单 URL、归属分组，并提示「把 `title` 换成中文」。

#### Scenario: 完成提示不误导 AI 修改 Layout.vue

- **WHEN** AI 阅读 `make gen` 的完成提示中关于前端路由/菜单的说明
- **THEN** 提示明确说明「菜单从路由声明分组派生，只需在 `router/index.js` 注册/注入，无需也禁止改动 `Layout.vue`」
- **AND** 提示中不出现要求编辑 `Layout.vue` 的 menus 的措辞

#### Scenario: 完成提示不列出已自动化的步骤

- **WHEN** 权限码注册与前端路由注册已由生成器自动完成时
- **THEN** 完成提示中不再将它们列为手工步骤

#### Scenario: 完成提示给出可执行验证命令

- **WHEN** AI 按完成提示执行验证命令
- **THEN** 命令在提示所指的目录下确实可用（`make` 目标 MUST 在仓库根目录执行，不得指示在 `backend/` 下执行 `make test` / `make smoke`）

#### Scenario: 完成提示包含最终 URL 与中文标题 TODO

- **WHEN** AI 阅读完成提示
- **THEN** 提示包含新模块的最终访问 URL（自建分组为 `/assets`，注入已有分组为 `/asset/<子 path>`）与「把 `title` 换成中文」提醒
