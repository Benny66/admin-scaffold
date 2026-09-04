## MODIFIED Requirements

### Requirement: 依赖清单单一真相

项目 MUST 在根目录维护 `deps.yaml`，作为后端（`go.mod` 直接依赖）、前端（`package.json` dependencies）、移动端（`package.json` dependencies）、小程序（`package.json` dependencies）**四端**依赖的机器可读登记表，每个条目附一句引入理由。四端的直接依赖 MUST 全部登记，guard 测试双向校验清单与各端 `go.mod` / `package.json` 的一致性。

#### Scenario: 新增依赖的登记入口

- **WHEN** AI 或开发者需要为一具体项目引入新依赖（如 excelize 做导入导出、echarts 做报表、`@dcloudio/uni-popup` 做小程序弹窗）
- **THEN** 其在 `deps.yaml` 对应端下追加一条 `{ package, reason }`，而非仅 `go get` / `npm install`

#### Scenario: 登记理由可追溯

- **WHEN** 审查 `deps.yaml` 中任一依赖
- **THEN** 能读到它为何被引入（reason 字段），从而判断该依赖是否为有意识决策

#### Scenario: miniapp 段与 package.json 一致

- **WHEN** guard 测试运行
- **THEN** `deps.yaml` 的 `miniapp:` 段每个条目都在 `miniapp/package.json` 的 dependencies 中存在；`miniapp/package.json` 的每个直接依赖都在 `deps.yaml` 的 `miniapp:` 段登记
