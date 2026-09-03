## ADDED Requirements

### Requirement: 新模块命名统一为复数

生成器产出的路由路径与权限码前缀 MUST 使用资源名复数形式（`<资源复数>:<动作>`），与存量五件套（`/users`、`roles:view`）及 REST 惯例保持一致。复数化 MUST 由生成器内置的 bash 规则表完成，不引入新依赖。

#### Scenario: 生成常规名词模块

- **WHEN** AI 执行 `make gen name=asset`
- **THEN** 生成器产出的路由组为 `/assets`，权限码前缀为 `assets:view` / `assets:create` / `assets:edit` / `assets:delete`

#### Scenario: 生成以 y 结尾的模块名

- **WHEN** AI 执行 `make gen name=asset_category`
- **THEN** 生成器产出的路由组为 `/asset_categories`，权限码前缀为 `asset_categories:view` 等，而非 `asset_categorys`

#### Scenario: 生成以 s/x/ch/sh 结尾的模块名

- **WHEN** AI 执行 `make gen name=box` 或 `make gen name=match`
- **THEN** 生成器产出的路由组为 `/boxes` / `/matches`

#### Scenario: 范例模板与生成器行为一致

- **WHEN** 阅读 `_example/router/example.go` 与 `_example/frontend/api.js` 的注释示例
- **THEN** 注释中展示的路径与权限码为复数形式，与生成器实际产出一致

### Requirement: 生成器完成提示与实际行为一致

生成器的完成提示 MUST 准确反映「哪些步骤已自动化、哪些仍需手工」，且 MUST NOT 指示开发者修改已被护栏禁止改动的既有约定。

#### Scenario: 完成提示不误导 AI 修改 Layout.vue

- **WHEN** AI 阅读 `make gen` 的完成提示第 3 条
- **THEN** 提示明确说明「只需在 `router/index.js` 注册路由，菜单自动出现，无需改动 `Layout.vue`」
- **AND** 提示中不出现要求编辑 `Layout.vue` 的 menus 的措辞

#### Scenario: 完成提示不列出已自动化的步骤

- **WHEN** 权限码注册与前端路由注册已由生成器自动完成时
- **THEN** 完成提示中不再将它们列为手工步骤

## MODIFIED Requirements

### Requirement: 模块代码生成器

项目 MUST 提供 `scripts/gen-module.sh <name>`，能从 `_example/` 生成新模块的完整骨架，且**三端全闭环**：

- 后端：`models/` + `services/` + `controllers/`
- 后端注入：`router.go` 路由注册、`database.go` 的 `AutoMigrate`、`database.go` 的 `initBaseData` 权限码注册
- 前端：`views/<name>/index.vue`、`api/index.js` 追加 API 定义、`router/index.js` 路由注册

所有生成文件 MUST 带明确的 `// TODO: 业务逻辑` 锚点。

#### Scenario: AI 用生成器创建新模块

- **WHEN** AI 执行 `make gen name=asset`
- **THEN** 生成器产出 `asset` 模块的后端三层、路由注册、AutoMigrate 注册、权限码注册、前端页面、API 定义与前端路由条目，且 `make test` 与 `make smoke` 均通过

#### Scenario: 生成器产出带业务锚点

- **WHEN** 生成完成后打开任一生成文件
- **THEN** 文件中存在明确的 `// TODO: 业务逻辑` 注释，标记 AI 需要手写的部分

#### Scenario: 生成器自动注册权限码

- **WHEN** AI 执行 `make gen name=asset` 后查看 `database.go`
- **THEN** `initBaseData` 的权限声明块中新增 `assets:view` / `assets:create` / `assets:edit` / `assets:delete` 四条记录

#### Scenario: 生成器自动注册前端路由

- **WHEN** AI 执行 `make gen name=asset` 后查看 `frontend/src/router/index.js`
- **THEN** 在 `【gen:route】` 锚点处新增一条指向 `views/asset/index.vue`、带 `meta.permission: 'assets:view'` 的路由条目

#### Scenario: 生成后目标已存在

- **WHEN** 同名模块文件已存在时执行 `make gen`
- **THEN** 生成器报错退出且不覆盖任何已有文件

### Requirement: 黄金路径一致性校验

guard 测试 MUST 校验 `_example/` 的结构不被腐化，且新生成模块与 `_example/` 的结构保持一致，防止范例自身变成新的漂移源。

#### Scenario: 范例模块结构被破坏

- **WHEN** 有人直接修改 `_example/` 使其偏离三层结构约定
- **THEN** 对应 guard 测试失败，提示范例已偏离标准答案

#### Scenario: 范例模板与生成器产出不同步

- **WHEN** 修改生成器行为（如复数化规则）但未同步更新 `_example/` 中的注释模板
- **THEN** 存在可执行的校验指出范例模板与生成器实际产出不一致，防止注释成为新的漂移源
