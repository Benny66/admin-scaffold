# Proposal: replace-dependency-denylist-with-registry

## Why

基座根 `AGENTS.md` 第 2 条用「硬编码禁单」管理依赖，列了 `excelize`、`echarts`、`jsbarcode`、`qrcode`、`html5-qrcode` 五个「禁止引入」的包，并给出判据「新增依赖 MUST 是可跨项目复用的通用依赖」。

这条规则有两个内在缺陷：

1. **自相矛盾**：被禁的五个包里，`excelize`（通用 Excel）、`echarts`（通用图表）、`qrcode`/`jsbarcode`（通用条码/二维码）恰恰都满足它自己「可跨项目复用」的判据——它们是通用库，不是业务专属。真正业务专属的只有「资产编码/打印相关业务包」这类东西。

2. **软约束无兜底**：这条规则没有任何 guard 测试兜底。`internal/guard/` 只锁分层、响应协议、模型完整性、范例锚点，**依赖这条完全靠 AI 自觉**。违反时 `make test` 依然是绿的。

更现实的问题是：**当用户 copy 这个基座去开发真实项目（比如资产管理系统）时，`excelize`/`echarts`/`qrcode` 恰恰是刚需依赖**。硬禁单从「保持基座精简」变成了「卡住真实项目」，用户被一条本就没有执行力的规则挡住。

本 change 把「硬禁单」替换为「**显式依赖登记制**」：新增依赖不是禁止，而是**必须登记 + 写理由**，并由 guard 测试强制执行——保留「约束即代码」的护栏价值，同时消除「禁单挡路」与「规则自相矛盾」两个问题。

## What Changes

- 新增 `deps.yaml`（项目级，跟项目走）：登记项目实际使用的前后端/移动端依赖，每个条目带一句话理由。
- 把根 `AGENTS.md` 第 2 条从「禁止引入的依赖（硬禁单）」改写为「**依赖登记制**」：新增依赖 MUST 登记到 `deps.yaml`，判据改为「登记 + 写理由」，不再硬编码「禁止」清单。
- `backend/internal/guard/` 新增依赖登记制 guard 测试：解析 `deps.yaml`，校验「已登记依赖与 `go.mod`/`package.json` 实际依赖一致」，未登记的新增依赖触发失败。
- `deps.yaml` 初始内容 = 当前基座三端实际依赖（`go.mod` + `frontend/package.json` + `mobile/package.json` 的直接依赖），并显式标注「基座基线不含 excelize/echarts/qrcode 等业务库，但**不禁止**项目登记引入」。
- 文档：`docs/` 新增或更新依赖管理说明，讲清「加依赖 = 在 deps.yaml 登记一行 + 写理由，而非偷偷 import」。

## Capabilities

### New Capabilities

- `dependency-registry`: 显式依赖登记制 —— 用 `deps.yaml` 作为项目依赖的单一真相，新增依赖须登记并附理由。

### Modified Capabilities

- `ai-scaffold-guardrails`: 新增「依赖登记制 guard 测试」，把「新增依赖必须登记」从软约束编译成会失败的静态检查。

## Impact

- **新增文件**：`deps.yaml`、`backend/internal/guard/deps_test.go`、`docs/依赖管理.md`。
- **修改文件**：根 `AGENTS.md`（第 2 条改写）。
- **无新第三方依赖**：guard 测试用 `go/parser`/`os`/`gopkg.in/yaml.v3`（yaml.v3 已是后端既有直接依赖，见 `go.mod`），沿用 guard「不引入新依赖」的既有约定。
- **对现有代码零破坏**：`deps.yaml` 初始内容与当前三端实际依赖一一对应，登记制 guard 首次运行即绿；`make test` 的既有 guard 不受影响。
- **范围外**：不迁移 asset-admin 的资产业务代码，不改变 `base-backend` 模块名，不引入任何被原禁单点名的包。
