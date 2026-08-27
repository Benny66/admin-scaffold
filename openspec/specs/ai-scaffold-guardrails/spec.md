# ai-scaffold-guardrails Specification

## Purpose
TBD - created by archiving change ai-scaffold-maturity. Update Purpose after archive.
## Requirements
### Requirement: 分层铁律由 guard 测试强制

后端 MUST 提供 guard 测试，扫描源码以确保分层约束不可被 AI 静默违反：`services/` 不得导入 Gin，`controllers/` 不得直接操作 GORM。

#### Scenario: AI 在 service 层引入了 gin 依赖

- **WHEN** `services/` 下任一 Go 文件 `import "github.com/gin-gonic/gin"`
- **THEN** 对应 guard 测试失败，CI 变红，AI 得到「service 层不得触碰 gin.Context」的明确失败信号

#### Scenario: AI 在 controller 层直接操作 GORM

- **WHEN** `controllers/` 下任一 Go 文件 `import "gorm.io/gorm"` 或直接引用 `database.DB`
- **THEN** 对应 guard 测试失败，阻断合并

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

