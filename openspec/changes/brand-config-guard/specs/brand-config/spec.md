# Spec: brand-config (delta)

修正环境变量支持的错误表述，并补充 `subtitle` / `favicon` 必须被两端 store 消费的要求。

## MODIFIED Requirements

### Requirement: 品牌配置由 config.yaml 驱动

后端 MUST 在 `config.go` 的 `AppConfig` 中提供 `name`、`logo`、`favicon`、`footer` 四个字段，均支持默认值 + YAML 覆盖，使品牌四要素均可通过 `config.yaml` 设置。品牌段不支持环境变量覆盖；环境变量通道仅适用于 `server`、`database`、`jwt` 三段。

#### Scenario: 用户设置 config.yaml 品牌字段

- **WHEN** `config.yaml` 的 `app` 段填写 `name`、`logo`、`favicon`、`footer`
- **THEN** 后端启动后这些值生效，通过 `GetSystemInfo` 返回给前端

#### Scenario: 未配置时使用安全默认值

- **WHEN** 字段留空
- **THEN** 后端返回空串，前端渲染点回退到文字/不显示，不产生破图或报错

#### Scenario: 环境变量不覆盖品牌段

- **WHEN** 设置了 `APP_NAME` 之类的环境变量
- **THEN** 品牌字段不受影响，仍取 config.yaml 或默认值

### Requirement: GetSystemInfo 返回完整品牌信息

`GetSystemInfo` MUST 返回 `name`、`subtitle`、`logo`、`favicon`、`footer` 五个字段，其中 `logo`/`favicon` 为拼好的完整路径 `/static/<file>`，`favicon` 缺省回退 `logo`；且这五个字段 MUST 全部被 `frontend/src/stores/app.js` 与 `mobile/src/stores/app.js` 的 `fetchSystemInfo` 解构消费（由 `brand-config-guard` 的 guard 测试强制）。

#### Scenario: favicon 未配置时回退 logo

- **WHEN** `favicon` 为空且 `logo` 非空
- **THEN** `GetSystemInfo` 返回的 `favicon` 等于 `/static/<logo 文件名>`

#### Scenario: logo 与 favicon 分别配置

- **WHEN** `logo` 与 `favicon` 都非空
- **THEN** 返回各自的 `/static/<file>` 完整路径

#### Scenario: subtitle 被两端消费而非死配置

- **WHEN** 系统信息返回非空 `subtitle`
- **THEN** 两端 store 将其持久化，消费点读到该值（不再硬编码副标题文案）

#### Scenario: 移动端消费 favicon

- **WHEN** 移动端系统信息返回非空 `favicon`
- **THEN** 移动端设置 `link[rel="icon"]`，与前端行为对齐
