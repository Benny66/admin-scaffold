# Proposal: ai-scaffold-maturity

## Why

当前脚手架自称「给 AI 使用」，但只完成了「约束层」——AGENTS.md 与 docs 告诉 AI 规则是什么。真正让 AI 高效且正确工作的「生成层」（模板/codegen/通用组件）与「验证层」（测试/契约/冒烟/护栏）整片缺失。结果是：AI 每新增一个模块都要模仿已漂移的五件套抄 300 行重复代码；写完只有 `go build` 能证明"编译过"，无法证明"能跑且没违反分层铁律"；RBAC 权限中间件定义了却从未接线，制造"权限已保护"的假象。

## What Changes

- 分层宪法：根 AGENTS.md 只留通用铁律，新增 `backend/CLAUDE.md` 与 `frontend/CLAUDE.md` 域宪法，就近可读。
- 黄金路径 + 代码生成：新增唯一范例模块 `_example/`，其余模块由 `scripts/gen-module.sh` 生成。
- 架构护栏：新增 guard 测试，把分层铁律/响应协议/AutoMigrate 完整性编译成会红的测试。
- 冒烟验证：新增 `scripts/smoke.sh`（启动 → 登录 → 命中受保护路由 → 断言）。
- 机器可读契约：新增 `contracts/openapi.yaml` 作为字段名唯一真相。
- 单一入口：新增 `Makefile` 收敛 test/lint/smoke/gen。
- 安全接线：为受保护路由挂上 `PermissionRequired` 权限码校验（修安全假象）。

## Capabilities

### New Capabilities

- `ai-scaffold-guardrails`: 分层宪法、架构护栏、机器可读契约、冒烟验证、单一入口——把「约束即代码」与「让漂移大声响」落地。
- `ai-scaffold-codegen`: 黄金路径范例模块 + 模块代码生成器，把 AI 的新增模块从「模仿五件套」变成「跑一条命令」。

### Modified Capabilities

（无。本 change 不改变现有 spec 级行为，纯增量基建。）

## Impact

- 新增文件：`backend/CLAUDE.md`、`frontend/CLAUDE.md`、`backend/_example/`、`backend/scripts/gen-module.sh`、`backend/scripts/smoke.sh`、`contracts/openapi.yaml`、`Makefile`、guard 测试若干。
- 修改文件：根 `AGENTS.md`（瘦身为指针）、`backend/router/router.go`（挂权限中间件）。
- 现有业务代码零破坏：五件套继续可用，只是从「隐性范例」降级为「历史模块」。
