## MODIFIED Requirements

### Requirement: 两端同等覆盖

护栏 MUST 对 `frontend/`、`mobile/` 与 `miniapp/` 各执行一组断言，任一端违规即失败。

#### Scenario: 仅一端漏定义

- **WHEN** 移动端 store 定义了某 action 而桌面端未定义，但两端 `fetchSystemInfo` 都调用了它
- **THEN** guard 测试仅对桌面端报红，并指明是哪一端

#### Scenario: miniapp store 也被覆盖

- **WHEN** `miniapp/src/stores/app.js` 引用了不存在的成员
- **THEN** guard 测试对 miniapp 端报红，错误信息指明文件名、成员名与所在端

#### Scenario: 三端同等断言

- **WHEN** 三端中任一端的 `appStore.<成员>` 无法在对应 `stores/app.js` 中解析到
- **THEN** guard 测试失败，错误信息指明端别（frontend / mobile / miniapp）
