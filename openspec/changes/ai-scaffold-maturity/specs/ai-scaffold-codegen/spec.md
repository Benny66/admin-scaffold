# Spec: ai-scaffold-codegen

「黄金路径 + 代码生成」的落地：唯一范例模块 `_example/` 作为标准答案，代码生成器从范例产出新模块骨架，机器可读契约作为字段名唯一真相，使 AI 新增模块从「模仿已漂移的五件套」变成「跑一条命令，填业务逻辑」。

## ADDED Requirements

### Requirement: 唯一黄金路径范例模块

项目 MUST 提供唯一范例模块 `_example/`（后端）作为新增业务模块的标准答案，并在 `docs/map.md` 与根 `AGENTS.md` 中显式标注「五件套=历史模块，范例=`_example/`」，使 AI 不再以五件套为模仿对象。

#### Scenario: AI 需要新增一个业务模块

- **WHEN** AI 被告知「新增一个模块」
- **THEN** 它在 `_example/` 找到唯一范例，而不是去读五个互相漂移的 system 页面

#### Scenario: AI 误读了五件套作为范例

- **WHEN** 导航地图与根宪法明确标注五件套为「历史模块」
- **THEN** AI 能识别 `_example/` 才是标准答案，避免复制已漂移的模式

### Requirement: 模块代码生成器

项目 MUST 提供 `scripts/gen-module.sh <name>`，能从 `_example/` 生成新模块的完整骨架：`models/` + `services/` + `controllers/` + `router` 注册 + 前端 `views/` + `api/`，并带明确的 `// TODO: 业务逻辑` 锚点。

#### Scenario: AI 用生成器创建新模块

- **WHEN** AI 执行 `make gen name=asset`
- **THEN** 生成器产出 `asset` 模块的后端三层、路由注册、前端页面与 API 定义，且可编译通过

#### Scenario: 生成器产出带业务锚点

- **WHEN** 生成完成后打开任一生成文件
- **THEN** 文件中存在明确的 `// TODO: 业务逻辑` 注释，标记 AI 需要手写的部分

### Requirement: 黄金路径一致性校验

guard 测试 MUST 校验 `_example/` 的结构不被腐化，且新生成模块与 `_example/` 的结构保持一致，防止范例自身变成新的漂移源。

#### Scenario: 范例模块结构被破坏

- **WHEN** 有人直接修改 `_example/` 使其偏离三层结构约定
- **THEN** 对应 guard 测试失败，提示范例已偏离标准答案

### Requirement: 机器可读契约作为字段名唯一真相

项目 MUST 提供 `contracts/openapi.yaml` 定义字段名与响应形状（如 `page_size`、`{code,message,data}`、分页 `{list,total,page,page_size}`），前端 API 定义与后端结构体 tag 从该契约派生，杜绝三端漂移。

#### Scenario: AI 新增接口需要知道字段命名

- **WHEN** AI 需要为某个接口定义请求/响应字段
- **THEN** 它以 `contracts/openapi.yaml` 为准，使用 `page_size` 而非 `pageSize`

#### Scenario: 三端字段命名不一致被检出

- **WHEN** 后端 struct tag、前端 API 定义、契约三者中字段名出现不一致
- **THEN** 存在可执行的校验（契约校验或代码生成）能指出不一致处

## MODIFIED Requirements

（无。）

## REMOVED Requirements

（无。）
