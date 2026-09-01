# Spec: brand-config-guard

把品牌配置的「四处副本一致性」与「后端返回字段必须被前端消费」编译成会失败的 guard 测试，杜绝漏改与死配置复发。

## ADDED Requirements

### Requirement: 品牌配置结构体与 YAML 影子结构字段一致

`config.AppConfig` 与 `config.yamlFile` 的 `App` 字段 MUST 拥有一致的 yaml tag 集合——两者的差集（任一方向）都必须为空。

#### Scenario: 新增字段时漏改 yamlFile

- **WHEN** 有人在 `AppConfig` 中新增一个带 yaml tag 的字段，但未在 `yamlFile.App` 中同步添加
- **THEN** `go test ./internal/guard/` 失败，错误信息指明缺失的字段名与应在的位置（`backend/config/config.go` 的 yamlFile 结构体）

#### Scenario: 反向漏改 AppConfig

- **WHEN** `yamlFile.App` 中存在 `AppConfig` 没有的字段
- **THEN** guard 测试失败，指明该僵尸字段

#### Scenario: 字段集合一致

- **WHEN** 两者 yaml tag 集合完全相同
- **THEN** guard 测试通过

### Requirement: 每个品牌字段都有非空覆盖分支

`config.go` 的 YAML 覆盖链 MUST 为每个带 yaml tag 的品牌字段提供对应的 `if yf.App.<Field> != ""` 分支，将其赋给 `GlobalConfig.App.<Field>`。

#### Scenario: 新增字段但漏写覆盖分支

- **WHEN** `AppConfig` 与 `yamlFile` 都已同步某字段，但覆盖链中没有对应的 if 分支
- **THEN** guard 测试失败，指明该字段缺少覆盖分支

#### Scenario: 覆盖分支完整

- **WHEN** 每个字段都有对应的非空覆盖分支
- **THEN** guard 测试通过

### Requirement: GetSystemInfo 返回字段必须被两端 store 消费

`GetSystemInfo`（`backend/controllers/system.go`）中 `gin.H{}` 的每个 key MUST 被 `frontend/src/stores/app.js` 与 `mobile/src/stores/app.js` 的 `fetchSystemInfo` 解构接收。

#### Scenario: 后端新增返回字段但前端未消费

- **WHEN** `GetSystemInfo` 的响应新增一个 key，而任一端 store 的 `const { ... } = res.data` 未包含它
- **THEN** guard 测试失败，指明该字段名与未消费的一端

#### Scenario: 两端都已消费

- **WHEN** 后端每个返回 key 都被两端 store 解构
- **THEN** guard 测试通过

#### Scenario: subtitle 不再是死配置

- **WHEN** 系统信息返回 `subtitle`
- **THEN** 两端 store 均将其持久化，消费点（登录页副标题）读到该值而非硬编码文案

#### Scenario: 移动端消费 favicon

- **WHEN** 移动端系统信息返回非空 `favicon`
- **THEN** 移动端设置 `link[rel="icon"]`，与前端行为对齐

### Requirement: guard 测试零新依赖且失败信息可定位

新增 guard 测试 MUST 只使用标准库（`reflect`、`go/parser`、`go/ast`、`regexp`、`os`）与 guard 包既有辅助函数，且每条断言的失败信息 MUST 指明「漏改了哪一处」。

#### Scenario: 依赖清单不变

- **WHEN** 本 change 落地后检查 `deps.yaml`
- **THEN** 无新增登记项（未引入第三方依赖）

#### Scenario: 失败信息可定位

- **WHEN** 任一护栏断言失败
- **THEN** 错误信息包含缺失的字段名与具体位置（结构体名 / 覆盖链 / 哪个端的 store 文件）

#### Scenario: 护栏纳入统一验证入口

- **WHEN** 执行 `make test`
- **THEN** 本 change 的三组断言随 guard 包一同执行
