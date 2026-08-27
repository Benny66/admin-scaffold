# Spec: dependency-registry

依赖登记制 —— 用 `deps.yaml` 作为项目依赖的单一真相，新增依赖须登记并附理由。

## ADDED Requirements

### Requirement: 依赖清单单一真相

项目 MUST 在根目录维护 `deps.yaml`，作为后端（`go.mod` 直接依赖）、前端（`package.json` dependencies）、移动端（`package.json` dependencies）三端依赖的机器可读登记表，每个条目附一句引入理由。

#### Scenario: 新增依赖的登记入口

- **WHEN** AI 或开发者需要为一个具体项目引入新依赖（如 excelize 做导入导出、echarts 做报表）
- **THEN** 其在 `deps.yaml` 对应端下追加一条 `{ package, reason }`，而非仅 `go get` / `npm install`

#### Scenario: 登记理由可追溯

- **WHEN** 审查 `deps.yaml` 中任一依赖
- **THEN** 能读到它为何被引入（reason 字段），从而判断该依赖是否为有意识决策

### Requirement: 依赖登记制取代硬禁单

项目 MUST 以「登记制」管理依赖：新增依赖的唯一准则是「已登记并附理由」，不再维护硬编码的「禁止引入」清单。基座基线不含 excelize/echarts/qrcode 等业务库，但**不禁止**具体项目登记引入。

#### Scenario: 项目需要被原禁单点名的包

- **WHEN** 一个具体项目需要引入 `excelize`（Excel 导入导出）
- **THEN** 其被允许，只需在 `deps.yaml` 登记 `{ package: github.com/xuri/excelize/v2, reason: 资产/员工导入导出 }`

#### Scenario: 不再存在「禁止引入」的散文规则

- **WHEN** 阅读根 `AGENTS.md`
- **THEN** 不再出现「禁止引入 excelize/echarts/qrcode」之类的硬禁单，取而代之的是「新增依赖 MUST 登记到 deps.yaml」
