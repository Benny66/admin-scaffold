# Proposal: brand-config-guard

## Why

`backend/config/config.go` 里品牌配置**每个字段都有四份副本**：`AppConfig` 结构体（L23-29）、`init()` 默认值（L80-87）、`yamlFile` 影子结构（L152-159）、以及逐字段手写的 `if yf.App.X != ""` 覆盖链（L208-223）。加一个品牌字段要动 4 处，漏改任何一处都不会报错——只会静默产生「配了不生效」或「永远取默认值」的诡异行为。

这个风险马上要被兑现：`login-brand-visual` 需要新增 `login_bg` 与 `login_bg_mobile` 两个字段，即 **8 处改动、2 次漏改机会**。

同时存在一个**已经发生**的同类缺陷：`GetSystemInfo` 返回 `subtitle`（`controllers/system.go:31`），但两端 store 的解构都没接收它（`frontend/src/stores/app.js:45`、`mobile/src/stores/app.js:40`）——`subtitle` 是死配置。而登录页恰好硬编码了本该由它驱动的文案「系统管理基座」（`frontend/src/views/Login.vue:5`）。这与归档在 `openspec/changes/archive/2026-08-27-brand-config/design.md` 里吐槽 `footer`「定义了却从未被前端消费」是同一个反模式的复发。

本 change 把这两类一致性从「人工纪律」编译成 `make test` 会红的静态检查。

## What Changes

- 新增 `backend/internal/guard/brand_config_test.go`，提供三组断言：
  - **G1 结构奇偶**：`AppConfig` 与 `yamlFile.App` 的 yaml tag 集合必须完全一致（反射）。
  - **G2 覆盖链完整**：每个带 yaml tag 的字段，在 `config.go` 的覆盖链中必须存在对应的 `if yf.App.<F> != ""` 分支（go/ast）。
  - **G3 前端消费完整**：`GetSystemInfo` 中 `gin.H{}` 的每个 key，必须被 `frontend/src/stores/app.js` 与 `mobile/src/stores/app.js` 的 `fetchSystemInfo` 解构接收。
- **修复既有的 `subtitle` 死配置**：两端 store 补上 `subtitle` 的解构与持久化；移动端一并补 `favicon` 消费。让 G3 落地时构建是绿的。
- **修正 spec ↔ 代码漂移**：`openspec/specs/brand-config/spec.md:8` 声称品牌字段支持「环境变量覆盖」，但 `config.go` 的环境变量段（L116-136）只覆盖 server/database/jwt，完全没有 `app.*`。把 spec 表述收敛为「默认值 + YAML 覆盖」两层。

## Capabilities

### New Capabilities

- `brand-config-guard`: 把品牌配置的四处副本一致性、以及「后端返回字段必须被前端消费」编译成会失败的 guard 测试。

### Modified Capabilities

- `brand-config`: 补充 `subtitle` / `favicon` 的前端消费要求；修正环境变量支持的错误表述。

## Impact

- 修改文件：`backend/internal/guard/brand_config_test.go`（新增）、`frontend/src/stores/app.js`、`mobile/src/stores/app.js`、`openspec/specs/brand-config/spec.md`。`backend/config/config.go` 与 `backend/controllers/system.go` 本身不改，仅成为被测对象。
- 无新第三方依赖（沿用 guard 包既有的 go/ast + reflect，无新增）。
- 无破坏性：G3 断言在落地时因已修复 `subtitle` / `favicon` 而是绿的；此后任何新增后端返回字段都必须连同两端消费一起提交，否则构建红——这正是本 change 的意图。
- 本 change 是 `login-brand-visual` 的前置：先锁住基线，再在护栏下加字段。
