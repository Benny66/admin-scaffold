# Proposal: guard-read-module-name

## Why

`internal/guard/guard_test.go` 的 `Test_ControllerLayerMustNotTouchGORM` 用**硬编码字面量** `"base-backend/database"` 来识别「这是 database 包」，从而拦截 controller 直连数据库。

这个硬编码让「模块名」不再是单一真相：`go.mod` 里写的才是真名，guard 里却另写了一份。后果是——一旦复制基座做新项目、把 `go.mod` 的 module 从 `base-backend` 改成 `my-system`，guard 里那句 `if imp == "base-backend/database"` 永远为 false，**controller 直连 database 的护栏被静默拆掉**，`make test` 还照常绿。

这是后续「一键初始化脚本」（init.sh 改名包名）的前置障碍：只要 guard 里还有这个字面量，改名脚本要么留隐患、要么写一堆排除逻辑。本 change 消除这个硬编码，让 guard 从 `go.mod` 动态读取真实 module 名，使「包名」成为 `go.mod` 这一处的单一真相。

## What Changes

- `backend/internal/guard/guard_test.go`：新增 `readModuleName()`，从 `backend/go.mod` 解析 module 名（首行 `module <name>`）。
- `Test_ControllerLayerMustNotTouchGORM`：把 `if imp == "base-backend/database"` 改为 `if imp == moduleName+"/database"`，错误信息里的包名也动态拼接。
- 其余 guard 测试不变，行为等价（`base-backend` 场景下判断结果完全一致）。

## Capabilities

### Modified Capabilities

- `ai-scaffold-guardrails`: 「分层铁律由 guard 测试强制」中，controller 禁止直连 database 的判断，从硬编码字面量改为从 go.mod 动态读取模块名。

### New Capabilities

（无。本 change 是纯重构，不改变 guard 的语义，只消除硬编码。）

## Impact

- **修改文件**：`backend/internal/guard/guard_test.go`（新增一个 `readModuleName` 函数 + 改一处判断）。
- **无新依赖**：解析 go.mod 用标准库 `os`/`strings`，符合 guard「不引入第三方依赖」既有约定。
- **对现有行为零破坏**：`base-backend` 项目下，`readModuleName()` 返回 `base-backend`，判断结果与原来逐字节一致。
- **前提关系**：为后续「一键初始化脚本（init.sh 改包名）」铺路；本 change 自身不依赖任何其他 change。
- **范围外**：不写 init.sh、不改 `gen-module.sh`、不改 `_example/` 的 import（那些属于 init 脚本的职责）。
