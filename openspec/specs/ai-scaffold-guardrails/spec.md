# ai-scaffold-guardrails Specification

## Purpose
TBD - created by archiving change ai-scaffold-maturity. Update Purpose after archive.
## Requirements
### Requirement: 分层铁律由 guard 测试强制

后端 MUST 提供 guard 测试，扫描源码以确保分层约束不可被 AI 静默违反：`services/` 不得导入 Gin，`controllers/` 不得直接操作 GORM（`gorm.io/gorm` 或本项目 `<module>/database` 包）。

#### Scenario: AI 在 service 层引入了 gin 依赖

- **WHEN** `services/` 下任一 Go 文件 `import "github.com/gin-gonic/gin"`
- **THEN** 对应 guard 测试失败，CI 变红，AI 得到「service 层不得触碰 gin.Context」的明确失败信号

#### Scenario: AI 在 controller 层直接操作 GORM

- **WHEN** `controllers/` 下任一 Go 文件 `import "gorm.io/gorm"` 或 `import "<当前模块名>/database"`
- **THEN** 对应 guard 测试失败，阻断合并

#### Scenario: 项目改名后护栏依然生效

- **WHEN** 复制基座后把 `go.mod` 的 module 从 `base-backend` 改为其他模块名
- **THEN** guard 测试从 `go.mod` 动态读取模块名，controller 直连 `<新模块名>/database` 仍被正确拦截，护栏不因改名而失效

### Requirement: 统一响应协议由 guard 测试强制

后端 MUST 提供 guard 测试，确保 controller 层所有响应都通过 `utils/response.go` 的方法返回，禁止手写 `c.JSON` 拼响应。

#### Scenario: AI 在 controller 手写 c.JSON

- **WHEN** `controllers/` 下任一 Go 文件出现 `c.JSON(` 调用
- **THEN** 对应 guard 测试失败，提示必须使用 `utils.Success` / `utils.Fail` / `utils.SuccessPage` 等统一方法

### Requirement: 模型必须注册进 AutoMigrate

后端 MUST 提供 guard 测试，确保 `models/` 下每个模型结构体都出现在 `database/database.go` 的 `AutoMigrate` 列表，防止 AI 新增模型后漏建表。

#### Scenario: AI 新增模型但未加入 AutoMigrate

- **WHEN** `models/` 下存在一个模型结构体，且未在 `database.go` 的 `AutoMigrate(...)` 列表中出现
- **THEN** 对应 guard 测试失败，提示将该模型加入迁移列表

### Requirement: 冒烟验证闭环

项目 MUST 提供一条冒烟命令（`make smoke`），能在一条命令内完成「启动 → 登录 → 命中受保护路由 → 断言」，证明服务可运行、可鉴权、可响应。

#### Scenario: AI 修改后端后运行冒烟

- **WHEN** 执行 `make smoke` 且后端代码可正常编译启动
- **THEN** 脚本用默认管理员登录拿到 token，带着 token 请求一个受保护路由，断言返回 code 200

#### Scenario: 鉴权接线被破坏时冒烟失败

- **WHEN** 受保护路由的鉴权中间件被移除或破坏
- **THEN** `make smoke` 在「未携带 token 请求受保护路由应返回 401」或「携带 token 应返回 200」的断言上失败

### Requirement: RBAC 权限码校验接线

受保护路由 MUST 按资源挂接 `PermissionRequired` 权限码中间件，`AdminRequired` MUST 用于仅超管可访问的接口，使权限码从「已定义未启用」变为「实际拦截」。

#### Scenario: 非管理员用户访问需权限码的接口

- **WHEN** 一个持有效 token 但无 `users:view` 权限的用户请求 `GET /api/users`
- **THEN** 后端返回 403，且不返回用户列表数据

#### Scenario: 超管访问任意受保护接口

- **WHEN** `is_admin=true` 的用户请求任意挂接权限码的接口
- **THEN** 中间件直接放行，正常返回数据

### Requirement: 单一入口命令

项目 MUST 提供 `Makefile`，收敛 `test` / `lint` / `smoke` / `gen` / `dev` 等高频命令，使 AI 无需记忆零散脚本路径。

#### Scenario: AI 需要运行全部验证

- **WHEN** AI 执行 `make test`（或等价聚合命令）
- **THEN** guard 测试、单测、冒烟按序执行，任一失败即非零退出码

### Requirement: 依赖登记制由 guard 测试强制

后端 MUST 提供 guard 测试，解析 `deps.yaml` 与三端依赖清单（`go.mod` 直接依赖、`frontend/package.json` dependencies、`mobile/package.json` dependencies），双向校验一致性：清单里的直接依赖 MUST 已登记，登记项 MUST 真实存在于清单。

#### Scenario: AI 引入依赖但未登记

- **WHEN** AI 执行 `go get` 引入一个新直接依赖，或 `npm install` 引入一个新 dependencies 包，但未在 `deps.yaml` 登记
- **THEN** 对应 guard 测试失败，`make test` 非零退出，提示「依赖 X 未登记，请在 deps.yaml 追加并附理由」

#### Scenario: 登记项真实存在

- **WHEN** `deps.yaml` 登记的每个 `package` 都能在对应依赖清单（go.mod / package.json）中找到
- **THEN** 反向校验通过，不误报

#### Scenario: 登记了但未实际引入

- **WHEN** `deps.yaml` 中存在一条记录，但对应清单中并无该依赖（typo 或已移除）
- **THEN** 对应 guard 测试失败，提示该登记项为「僵尸条目」

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

#### Scenario: 被注释掉的权限码不算已注册

- **WHEN** 某条权限码声明被 `//` 注释掉
- **THEN** guard MUST 将其视为未注册并失败；提取前 MUST 先剔除行注释，否则会产生假阴性（注释一行是开发者最常见的临时操作）

#### Scenario: 护栏自身失效必须报警

- **WHEN** guard 未能从 `router.go` 解析出任何 `PermissionRequired` 码，或未能从 `database.go` 解析出任何权限码，或 `database.go` 缺少 `【gen:permissions】` 锚点
- **THEN** guard 测试 `Fatal` 失败并提示解析规则可能已变更，而非当作「无引用」静默通过

#### Scenario: 基座基线首次运行即绿

- **WHEN** 基座当前状态下首次运行该 guard（存量 `PermissionRequired` 码与 `initBaseData` 的 Permission 记录一一对应）
- **THEN** guard 测试通过，不产生误报

### Requirement: 权限初始化可增量补齐

`initBaseData` 的权限初始化 MUST 按 `code` 逐个幂等 upsert（不存在则创建，存在则跳过），MUST NOT 使用「整批 `Count == 0` 才执行」的守卫，使新增权限码对**已存在的数据库**同样生效。

`Sort` MUST 在运行时按当前已有最大值递增计算，MUST NOT 写死在注入的字面量中，以免与存量权限的 `Sort` 冲突。

角色权限关联（`role_permissions`）MUST 同样按 `(role_id, permission_id)` 幂等写入，否则仅有 `permissions` 记录而缺关联时，新权限码依然不生效。

#### Scenario: 已有数据库升级后补齐新权限码

- **WHEN** 一个已存在且 `permissions` 表非空的数据库，在服务启动时执行到一个新增了权限码的 `initBaseData`
- **THEN** 新增的权限码被写入 `permissions` 表，既有权限记录不被修改或删除

#### Scenario: 重复启动不产生重复条目

- **WHEN** 服务连续多次启动，`initBaseData` 被反复执行
- **THEN** `permissions` 表中每个 `code` 仍只有一条记录

#### Scenario: 新权限码的 Sort 不与存量冲突

- **WHEN** 新增权限码被写入一个已有若干条权限记录的数据库
- **THEN** 新增记录的 `Sort` 大于存量最大值，不产生重复排序值

#### Scenario: 普通用户角色自动获得新模块的只读权限

- **WHEN** 新增模块的 `<资源>:view` 权限码被写入
- **THEN** 「普通用户」角色按 `code LIKE '%:view'` 的既有规则获得该权限（该分配逻辑同样为按 code 幂等 upsert），使非管理员登录后菜单不为空

#### Scenario: 基座基线首次运行即绿

- **WHEN** `deps.yaml` 初始内容与当前三端实际依赖一一对应
- **THEN** 依赖登记制 guard 测试在基座首次运行即通过，不产生误报

