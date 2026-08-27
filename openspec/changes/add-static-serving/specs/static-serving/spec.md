# Spec: add-static-serving

后端对桌面端/移动端前端产物的条件托管与 SPA 回退，支持无产物时优雅降级为纯 API。

## ADDED Requirements

### Requirement: 后端条件托管桌面端产物

后端 MUST 在 `dist/` 目录存在时，通过 `r.Static("/", "./dist")` 托管桌面端前端产物，使访问根路径能返回前端页面。

#### Scenario: dist 目录存在

- **WHEN** 后端启动且工作目录下存在 `dist/` 目录
- **THEN** 访问 `/` 返回 `dist/index.html`，访问 `/assets/xxx.js` 等静态资源正常返回

#### Scenario: dist 目录不存在

- **WHEN** 后端启动但 `dist/` 目录不存在
- **THEN** 后端跳过桌面端托管，行为与纯 API 服务一致（不影响 `/api/*`）

### Requirement: 后端条件托管移动端产物

后端 MUST 在 `dist-mobile/` 目录存在时，通过 `r.Static("/m/", "./dist-mobile")` 托管移动端 H5 产物，使访问 `/m/` 能返回移动端页面。

#### Scenario: dist-mobile 目录存在

- **WHEN** 后端启动且 `dist-mobile/` 目录存在
- **THEN** 访问 `/m/` 返回移动端 `index.html`

#### Scenario: dist-mobile 目录不存在

- **WHEN** `dist-mobile/` 目录不存在
- **THEN** 后端跳过移动端托管，不影响其他路由

### Requirement: SPA 前端路由回退

后端 MUST 在托管桌面端时提供 SPA 回退：对非 `/api/`、非 `/static/` 前缀的未命中路径，返回 `dist/index.html`，使前端 history 路由（如 `/system/user` 刷新）不 404。

#### Scenario: 刷新前端子路由

- **WHEN** 用户直接访问 `/system/user`（前端 history 路由）且 dist 存在
- **THEN** 后端返回 `dist/index.html`，由前端路由接管渲染

#### Scenario: API 未命中不返回 HTML

- **WHEN** 请求 `/api/xxx`（未注册的 API 路径）
- **THEN** 后端返回 JSON 404（而非 index.html），前端拦截器能正确处理

### Requirement: 品牌静态资源不受 SPA 回退影响

后端 MUST 确保 `/static/` 前缀（品牌 logo/favicon）不受 SPA 回退影响，仍正常返回静态文件。

#### Scenario: 访问品牌 logo

- **WHEN** 请求 `/static/logo.png` 且文件存在
- **THEN** 返回图片（200），不被 SPA 回退拦截
