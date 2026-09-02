## ADDED Requirements

### Requirement: 权限码必须注册进 initBaseData

后端 MUST 提供 guard 测试，确保 `router.go` 中每个 `PermissionRequired("<code>")` 使用的权限码，都能在 `database/database.go` 的 `initBaseData` 权限声明块中找到对应注册，防止 AI 新增受保护路由后漏注册权限码。

本要求是既有「模型必须注册进 AutoMigrate」（防漏建表）的同构对偶：AutoMigrate 漏注册导致运行时表不存在，权限码漏注册导致非管理员用户静默 403。

#### Scenario: AI 新增路由但漏注册权限码

- **WHEN** `router.go` 中出现 `PermissionRequired("assets:view")`，而 `initBaseData` 的权限声明块中没有 `Code: "assets:view"`
- **THEN** 对应 guard 测试失败，提示将 `assets:view` 注册进 `initBaseData`

#### Scenario: 权限码已正确注册

- **WHEN** `router.go` 中所有 `PermissionRequired` 码都能在 `initBaseData` 权限声明块中找到
- **THEN** guard 测试通过

#### Scenario: 护栏解析范围限定在权限声明块内

- **WHEN** guard 从 `database.go` 提取已注册权限码
- **THEN** 提取 MUST 限定在 `【gen:permissions】` 锚点起始的权限声明块内，不得把 `Role`、`DictType` 等其他模型的 `Code:` 字段误当作权限码

#### Scenario: 护栏自身失效必须报警

- **WHEN** guard 未能从 `router.go` 解析出任何 `PermissionRequired` 码，或未能从 `database.go` 解析出任何权限码
- **THEN** guard 测试 `Fatal` 失败并提示解析规则可能已变更，而非当作「无引用」静默通过

#### Scenario: 基座基线首次运行即绿

- **WHEN** 基座当前状态下首次运行该 guard（存量 17 个 `PermissionRequired` 码与 `initBaseData` 的 17 条 Permission 记录一一对应）
- **THEN** guard 测试通过，不产生误报

### Requirement: 权限初始化可增量补齐

`initBaseData` 的权限初始化 MUST 按 `code` 逐个幂等 upsert（不存在则创建，存在则跳过），MUST NOT 使用「整批 `Count == 0` 才执行」的守卫，使新增权限码对**已存在的数据库**同样生效。

`Sort` MUST 在运行时按当前已有最大值递增计算，MUST NOT 写死在注入的字面量中，以免与存量权限的 `Sort` 冲突。

#### Scenario: 已有数据库升级后补齐新权限码

- **WHEN** 一个已存在且 `permissions` 表非空的数据库，在服务启动时执行到一个新增了权限码的 `initBaseData`
- **THEN** 新增的权限码被写入 `permissions` 表，既有权限记录不被修改或删除

#### Scenario: 重复启动不产生重复条目

- **WHEN** 服务连续多次启动，`initBaseData` 被反复执行
- **THEN** `permissions` 表中每个 `code` 仍只有一条记录

#### Scenario: 新权限码的 Sort 不与存量冲突

- **WHEN** 新增权限码被写入一个已有 17 条权限记录的数据库
- **THEN** 新增记录的 `Sort` 大于存量最大值，不产生重复排序值

#### Scenario: 普通用户角色自动获得新模块的只读权限

- **WHEN** 新增模块的 `assets:view` 权限码被写入
- **THEN** 「普通用户」角色按 `code LIKE '%:view'` 的既有规则获得该权限（该分配逻辑同样为按 code 幂等 upsert），使非管理员登录后菜单不为空
