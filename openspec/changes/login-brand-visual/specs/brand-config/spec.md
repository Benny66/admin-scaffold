# Spec: brand-config (delta)

新增登录页背景图配置字段，并要求两端 store 消费。

## MODIFIED Requirements

### Requirement: 品牌配置由 config.yaml 驱动

后端 MUST 在 `config.go` 的 `AppConfig` 中提供 `name`、`logo`、`favicon`、`footer`、`login_bg`、`login_bg_mobile` 六个字段，均支持默认值 + YAML 覆盖，使品牌要素与登录页背景均可通过 `config.yaml` 设置。品牌段不支持环境变量覆盖；环境变量通道仅适用于 `server`、`database`、`jwt` 三段。

#### Scenario: 用户设置 config.yaml 品牌字段

- **WHEN** `config.yaml` 的 `app` 段填写 `name`、`logo`、`favicon`、`footer`、`login_bg`、`login_bg_mobile`
- **THEN** 后端启动后这些值生效，通过 `GetSystemInfo` 返回给前端

#### Scenario: 未配置时使用安全默认值

- **WHEN** 字段留空
- **THEN** 后端返回空串，前端渲染点回退到文字/渐变/不显示，不产生破图或报错

### Requirement: GetSystemInfo 返回完整品牌信息

`GetSystemInfo` MUST 返回 `name`、`subtitle`、`logo`、`favicon`、`footer`、`login_bg`、`login_bg_mobile` 七个字段，其中 `logo`、`favicon`、`login_bg`、`login_bg_mobile` 为拼好的完整路径 `/static/<file>`，`favicon` 缺省回退 `logo`；且这些字段 MUST 全部被 `frontend/src/stores/app.js` 与 `mobile/src/stores/app.js` 的 `fetchSystemInfo` 解构消费（由 `brand-config-guard` 的 guard 测试强制）。

#### Scenario: 背景图路径拼接

- **WHEN** `login_bg` 配置为 `bg.png`
- **THEN** `GetSystemInfo` 返回的 `login_bg` 为 `/static/bg.png`

#### Scenario: 背景图未配置

- **WHEN** `login_bg` 与 `login_bg_mobile` 均为空
- **THEN** 返回空串，前端回退渐变背景

#### Scenario: 新增字段被两端消费

- **WHEN** `GetSystemInfo` 返回 `login_bg` 与 `login_bg_mobile`
- **THEN** 两端 store 均已解构并持久化这两项，否则 `go test ./internal/guard/` 失败
