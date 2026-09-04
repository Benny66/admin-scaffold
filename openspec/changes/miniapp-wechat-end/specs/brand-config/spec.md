## MODIFIED Requirements

### Requirement: GetSystemInfo 返回完整品牌信息

`GetSystemInfo` MUST 返回 `name`、`subtitle`、`logo`、`favicon`、`footer` 五个字段，其中 `logo`/`favicon` 为拼好的完整路径 `/static/<file>`，`favicon` 缺省回退 `logo`；且这五个字段 MUST 全部被 `frontend/src/stores/app.js`、`mobile/src/stores/app.js` 与 `miniapp/src/stores/app.js` 的 `fetchSystemInfo` 解构消费（由 `brand-config-guard` 的 guard 测试强制，三端同等覆盖）。

#### Scenario: favicon 未配置时回退 logo

- **WHEN** `favicon` 为空且 `logo` 非空
- **THEN** `GetSystemInfo` 返回的 `favicon` 等于 `/static/<logo 文件名>`

#### Scenario: logo 与 favicon 分别配置

- **WHEN** `logo` 与 `favicon` 都非空
- **THEN** 返回各自的 `/static/<file>` 完整路径

#### Scenario: subtitle 被三端消费而非死配置

- **WHEN** 系统信息返回非空 `subtitle`
- **THEN** 三端 store 将其持久化，消费点读到该值（不再硬编码副标题文案）

#### Scenario: 移动端消费 favicon

- **WHEN** 移动端系统信息返回非空 `favicon`
- **THEN** 移动端设置 `link[rel="icon"]`，与前端行为对齐

#### Scenario: miniapp 消费 systemName 与 logo

- **WHEN** miniapp 首页挂载并调用 `fetchSystemInfo`
- **THEN** `miniapp/src/stores/app.js` 持久化 `systemName` 与 `logo`，首页 navbar 显示系统名（无 logo 时回退 systemName 文字）；不消费 `favicon`（小程序无浏览器标签概念，`favicon` 字段在 miniapp 端被忽略不报错）
