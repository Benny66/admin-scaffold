# brand-config Specification

## Purpose
TBD - created by archiving change brand-config. Update Purpose after archive.
## Requirements
### Requirement: 品牌配置由 config.yaml 驱动

后端 MUST 在 `config.go` 的 `AppConfig` 中提供 `name`、`logo`、`favicon`、`footer` 四个字段，均支持默认值 + YAML 覆盖 + 环境变量（沿用既有三层配置范式），使品牌四要素均可通过 `config.yaml` 设置。

#### Scenario: 用户设置 config.yaml 品牌字段

- **WHEN** `config.yaml` 的 `app` 段填写 `name`、`logo`、`favicon`、`footer`
- **THEN** 后端启动后这些值生效，通过 `GetSystemInfo` 返回给前端

#### Scenario: 未配置时使用安全默认值

- **WHEN** 字段留空
- **THEN** 后端返回空串，前端渲染点回退到文字/不显示，不产生破图或报错

### Requirement: 后端托管品牌静态文件

后端 MUST 提供静态文件服务 `r.Static("/static", "./static")`，品牌图片放置在 `backend/static/` 下，通过 `/static/<file>` 访问。

#### Scenario: 访问配置的品牌 logo

- **WHEN** 浏览器请求 `/static/logo.png` 且 `backend/static/logo.png` 存在
- **THEN** 后端返回该图片（200），内容类型正确

#### Scenario: dev 环境访问静态文件

- **WHEN** 前端 dev 模式（5173/5174 端口）请求 `/static/logo.png`
- **THEN** vite proxy 将该请求转发到后端 8080，正常返回图片而非 404

### Requirement: GetSystemInfo 返回完整品牌信息

`GetSystemInfo` MUST 返回 `name`、`subtitle`、`logo`、`favicon`、`footer` 五个字段，其中 `logo`/`favicon` 为拼好的完整路径 `/static/<file>`，`favicon` 缺省回退 `logo`。

#### Scenario: favicon 未配置时回退 logo

- **WHEN** `favicon` 为空且 `logo` 非空
- **THEN** `GetSystemInfo` 返回的 `favicon` 等于 `/static/<logo 文件名>`

#### Scenario: logo 与 favicon 分别配置

- **WHEN** `logo` 与 `favicon` 都非空
- **THEN** 返回各自的 `/static/<file>` 完整路径

### Requirement: 前端渲染品牌 logo 与页脚

前端 `Layout.vue` 侧边栏 MUST 渲染品牌 logo（无 logo 时回退文字），底部 MUST 渲染 `footer`（为空不显示）。`fetchSystemInfo` MUST 拉取并持久化 logo/favicon/footer。

#### Scenario: logo 已配置

- **WHEN** 系统信息返回非空 logo
- **THEN** 侧边栏显示 `<img>` logo，而非纯文字

#### Scenario: logo 未配置

- **WHEN** 系统信息返回空 logo
- **THEN** 侧边栏回退显示品牌名称文字

### Requirement: 浏览器标签图标运行时跟随配置

前端 `fetchSystemInfo` MUST 在拉到 favicon 后动态设置 `document` 的 `link[rel="icon"]`，使浏览器标签图标跟随 config 而非编译期写死。

#### Scenario: favicon 已配置

- **WHEN** 系统信息返回非空 favicon
- **THEN** 浏览器标签图标更新为该 favicon

#### Scenario: 页面初次加载时 favicon 动态注入

- **WHEN** `document` 中不存在 `link[rel="icon"]`
- **THEN** 前端创建该 link 元素并挂到 head，再设置 href

### Requirement: 移动端补齐品牌链路

移动端 `stores/app.js` MUST 提供与前端同构的 `fetchSystemInfo`，Home/Login 页在挂载时调用，navbar/header 支持展示 logo。

#### Scenario: 移动端拉取品牌信息

- **WHEN** 移动端 Home 页挂载
- **THEN** 调用 `fetchSystemInfo`，`systemName` 与 `logo` 被填充，而非永远显示兜底值

#### Scenario: 移动端展示 logo

- **WHEN** 系统信息返回非空 logo
- **THEN** 移动端 navbar/header 显示 logo 图片

