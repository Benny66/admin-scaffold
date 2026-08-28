# Spec: ai-scaffold-guardrails

后端架构护栏（guard 测试）。本次变更：controller 禁止直连 database 的判断，从硬编码模块名改为从 go.mod 动态读取。

## MODIFIED Requirements

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
