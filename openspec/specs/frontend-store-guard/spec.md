# frontend-store-guard Specification

## Purpose
TBD - created by archiving change frontend-store-guard. Update Purpose after archive.
## Requirements
### Requirement: 模板与脚本中的 store 成员引用必须可解析

前端与移动端 `src/` 下所有 `.vue` 与 `.js` 文件中出现的 `appStore.<成员>`，MUST 能在对应端的 `stores/app.js` 中解析到——即命中 state 键、getter、action 三者之一。Pinia 内置成员（`$` 前缀，如 `$reset` / `$patch`）MUST 被豁免。

#### Scenario: 引用了不存在的成员

- **WHEN** 某 `.vue` 文件出现 `appStore.loginBgUrl`，而 store 中只有 `loginBg`
- **THEN** `go test ./internal/guard/` 失败，错误信息指明文件名、成员名与所在端

#### Scenario: 拼写正确的引用

- **WHEN** 所有 `appStore.<成员>` 都能在 store 中解析到
- **THEN** guard 测试通过

#### Scenario: Pinia 内置成员不被误报

- **WHEN** 代码调用 `appStore.$reset()` 或 `appStore.$patch(...)`
- **THEN** guard 测试不报错（`$` 前缀成员豁免）

### Requirement: store 内部的方法调用必须可解析

`stores/app.js` 中出现的 `this.<成员>(...)`，MUST 能在该 store 自身解析到（action 或 getter）。

#### Scenario: 调用了未定义的 action

- **WHEN** `fetchSystemInfo` 中调用 `this.setLoginBg(...)`，但 store 的 actions 中没有 `setLoginBg`
- **THEN** guard 测试失败，指明缺失的 action 名与 store 文件路径

#### Scenario: 调用已定义的 action

- **WHEN** 所有 `this.<成员>()` 调用都有对应的 action 定义
- **THEN** guard 测试通过

### Requirement: 两端同等覆盖

护栏 MUST 对 `frontend/` 与 `mobile/` 各执行一组断言，任一端违规即失败。

#### Scenario: 仅一端漏定义

- **WHEN** 移动端 store 定义了某 action 而桌面端未定义，但两端 `fetchSystemInfo` 都调用了它
- **THEN** guard 测试仅对桌面端报红，并指明是哪一端

### Requirement: 护栏失效时必须报红而非静默通过

若从 store 文件解析出的成员集合为空（意味 store 写法已变更、解析规则失效），guard 测试 MUST 以 Fatal 失败，而非当作「无引用」静默通过。

#### Scenario: store 写法变更导致解析失效

- **WHEN** store 从 options 式改为 setup 式，导致三组正则解析不到任何成员
- **THEN** guard 测试 Fatal，提示「若 store 写法已变更，请同步更新本护栏」

### Requirement: 失败信息可定位且零新依赖

护栏 MUST 只使用标准库（`regexp`、`os`、`path/filepath`、`strings`），且每条失败信息 MUST 包含「哪个文件、哪个成员、哪一端」，并列出可用成员。

#### Scenario: 失败信息包含可用成员

- **WHEN** 某引用解析失败
- **THEN** 错误信息列出该 store 的可用成员集合，便于直接对照修正

#### Scenario: 依赖清单不变

- **WHEN** 本 change 落地后检查 `deps.yaml`
- **THEN** 无新增登记项（未引入第三方依赖）

#### Scenario: 纳入统一验证入口

- **WHEN** 执行 `make test`
- **THEN** 本护栏随 guard 包一同执行

