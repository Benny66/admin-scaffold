# ai-scaffold-codegen Specification

## Purpose
TBD - created by archiving change ai-scaffold-maturity. Update Purpose after archive.
## Requirements
### Requirement: 唯一黄金路径范例模块

项目 MUST 提供唯一范例模块 `_example/`（后端）作为新增业务模块的标准答案，并在 `docs/map.md` 与根 `AGENTS.md` 中显式标注「五件套=历史模块，范例=`_example/`」，使 AI 不再以五件套为模仿对象。

#### Scenario: AI 需要新增一个业务模块

- **WHEN** AI 被告知「新增一个模块」
- **THEN** 它在 `_example/` 找到唯一范例，而不是去读五个互相漂移的 system 页面

#### Scenario: AI 误读了五件套作为范例

- **WHEN** 导航地图与根宪法明确标注五件套为「历史模块」
- **THEN** AI 能识别 `_example/` 才是标准答案，避免复制已漂移的模式

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

#### Scenario: 资源名已被占用时拒绝生成

- **WHEN** 待生成模块的复数路由路径或权限码前缀已被占用（如 `box` 与 `boxe` 的复数同为 `boxes`，或模块名与存量五件套撞车）
- **THEN** 生成器在任何文件生成之前报错退出（fail fast），不留下「后端已生成、前端未注入」的不一致状态

### Requirement: 黄金路径一致性校验

guard 测试 MUST 校验 `_example/` 的结构不被腐化，且新生成模块与 `_example/` 的结构保持一致，防止范例自身变成新的漂移源。

#### Scenario: 范例模块结构被破坏

- **WHEN** 有人直接修改 `_example/` 使其偏离三层结构约定
- **THEN** 对应 guard 测试失败，提示范例已偏离标准答案

#### Scenario: 范例模板与生成器产出不同步

- **WHEN** 修改生成器行为（如复数化规则）但未同步更新 `_example/` 中的注释模板
- **THEN** 存在可执行的校验指出范例模板与生成器实际产出不一致，防止注释成为新的漂移源

### Requirement: 机器可读契约作为字段名唯一真相

项目 MUST 提供 `contracts/openapi.yaml` 定义字段名与响应形状（如 `page_size`、`{code,message,data}`、分页 `{list,total,page,page_size}`），前端 API 定义与后端结构体 tag 从该契约派生，杜绝三端漂移。

#### Scenario: AI 新增接口需要知道字段命名

- **WHEN** AI 需要为某个接口定义请求/响应字段
- **THEN** 它以 `contracts/openapi.yaml` 为准，使用 `page_size` 而非 `pageSize`

#### Scenario: 三端字段命名不一致被检出

- **WHEN** 后端 struct tag、前端 API 定义、契约三者中字段名出现不一致
- **THEN** 存在可执行的校验（契约校验或代码生成）能指出不一致处

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

#### Scenario: 复数规则的单一真相

- **WHEN** 生成器与 guard 测试都需要计算复数形式
- **THEN** 两者 MUST 共同调用 `backend/scripts/pluralize.sh`，不得在两处各写一份规则

#### Scenario: 范例模板与生成器行为一致

- **WHEN** 阅读 `_example/router/example.go` 与 `_example/frontend/api.js` 的注释示例
- **THEN** 注释中展示的路径与权限码为复数形式，与生成器实际产出一致

### Requirement: 生成器完成提示与实际行为一致

生成器的完成提示 MUST 准确反映「哪些步骤已自动化、哪些仍需手工」，且 MUST NOT 指示开发者修改已被护栏禁止改动的既有约定。

#### Scenario: 完成提示不误导 AI 修改 Layout.vue

- **WHEN** AI 阅读 `make gen` 的完成提示中关于前端路由/菜单的说明
- **THEN** 提示明确说明「菜单从路由派生，只需在 `router/index.js` 注册，无需也禁止改动 `Layout.vue`」
- **AND** 提示中不出现要求编辑 `Layout.vue` 的 menus 的措辞

#### Scenario: 完成提示不列出已自动化的步骤

- **WHEN** 权限码注册与前端路由注册已由生成器自动完成时
- **THEN** 完成提示中不再将它们列为手工步骤

#### Scenario: 完成提示给出的验证命令可执行

- **WHEN** AI 按完成提示执行验证命令
- **THEN** 命令在提示所指的目录下确实可用（`make` 目标 MUST 在仓库根目录执行，不得指示在 `backend/` 下执行 `make test` / `make smoke`）

