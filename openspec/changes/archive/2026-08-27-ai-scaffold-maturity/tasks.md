# Tasks: ai-scaffold-maturity

实施顺序遵循 design 的 Migration Plan：先落验证层（风险最低的增量），再修安全假象，再落生成层，最后收口一致性。

## 1. 验证层：架构护栏 + 冒烟 + 单一入口

- [x] 1.1 建立 guard 测试骨架（`backend/internal/guard/` 或集中 `*_test.go`），引入 AST 解析工具（`go/parser`/`go/ast`），不引入第三方依赖
- [x] 1.2 实现「分层铁律」guard：`services/` 不得 import gin，`controllers/` 不得直接操作 GORM
- [x] 1.3 实现「统一响应协议」guard：`controllers/` 不得手写 `c.JSON`
- [x] 1.4 实现「模型完整性」guard：`models/` 每个模型必须出现在 `database.go` 的 `AutoMigrate`
- [x] 1.5 编写 `backend/scripts/smoke.sh`：随机端口 + 临时 db → 启动 → 登录拿 token → 命中受保护路由 → 断言 200 → 清理
- [x] 1.6 编写根 `Makefile`，收敛 `test` / `lint` / `smoke` / `gen` / `dev` 命令
- [x] 1.7 更新 `.github/workflows/ci.yml`，加入 `make test` 与 `make smoke` 步骤（不再只 build）
- [x] 1.8 运行 `make test`，确认 guard 测试对现有代码通过（基线绿）

## 2. 安全接线：RBAC 权限码

- [x] 2.1 在 `backend/router/router.go` 为受保护路由按资源挂接 `PermissionRequired("...")`
- [x] 2.2 将仅超管接口挂接 `AdminRequired()`（如清空日志等）
- [x] 2.3 修复 `middleware/logger.go` 异步写日志的 goroutine 竞态（不在 goroutine 内访问 `gin.Context`）
- [x] 2.4 运行 `make smoke`，确认鉴权接线生效、冒烟通过

## 3. 生成层：黄金路径 + 代码生成 + 契约

- [x] 3.1 从五件套中提炼唯一范例模块 `backend/_example/`，注释标注「这是模式」
- [x] 3.2 编写 `backend/scripts/gen-module.sh <name>`，从 `_example` 生成后端三层 + router 注册 + 前端 `views/` + `api/`，带 `// TODO: 业务逻辑` 锚点
- [x] 3.3 实现「黄金路径一致性」guard：校验 `_example/` 与生成模块结构一致
- [x] 3.4 编写 `contracts/openapi.yaml`（统一响应、分页、系统五件套接口的字段定义）
- [x] 3.5 用生成器跑通一个示例模块，确认生成骨架可编译、`make test` 仍绿

## 4. 一致性收口

- [x] 4.1 拆分根 `AGENTS.md`：只留通用铁律 + 指针，新增 `backend/CLAUDE.md`、`frontend/CLAUDE.md` 域宪法
- [x] 4.2 编写 `docs/map.md` 导航地图，标注「五件套=历史模块，范例=`_example/`」
- [x] 4.3 消除「企业管理系统」硬编码残留（`frontend/src/router/index.js`、`frontend/src/stores/app.js`、`config.go` 副标题）
- [x] 4.4 统一 `store/` / `stores/` 命名矛盾（移动端 `store/` 改为 `stores/` 或明确例外并写进域宪法）
- [x] 4.5 运行全量 `make test` + `make smoke`，确认收口后全绿
