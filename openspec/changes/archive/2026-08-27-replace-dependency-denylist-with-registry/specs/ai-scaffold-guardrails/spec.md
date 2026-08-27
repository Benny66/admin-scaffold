# Spec: ai-scaffold-guardrails

后端架构护栏（guard 测试），把软约束编译成会失败的静态检查。本次新增「依赖登记制」护栏。

## ADDED Requirements

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

#### Scenario: 基座基线首次运行即绿

- **WHEN** `deps.yaml` 初始内容与当前三端实际依赖一一对应
- **THEN** 依赖登记制 guard 测试在基座首次运行即通过，不产生误报
