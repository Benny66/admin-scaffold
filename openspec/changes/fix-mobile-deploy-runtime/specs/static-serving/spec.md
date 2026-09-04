# static-serving delta — fix-mobile-deploy-runtime

## MODIFIED Requirements

### Requirement: 后端条件托管移动端产物

后端 MUST 在 `dist-mobile/` 目录存在时，将移动端 H5 产物托管于 `/m/` 子路径，并满足：对 `/m/` 及其下任意路径，若 `dist-mobile/` 中存在对应静态文件（含 HTML 引用的 `/m/assets/*` 等资源），MUST 返回该文件；否则 MUST 回退返回 `dist-mobile/index.html`，使移动端 history 子路由（如 `/m/login`）直达/刷新不 404。`dist-mobile/` 不存在时后端 MUST 跳过移动端托管，不影响其他路由。

#### Scenario: dist-mobile 目录存在且访问移动端入口

- **WHEN** 后端启动且 `dist-mobile/` 目录存在，浏览器访问 `/m/`
- **THEN** 返回 `dist-mobile/index.html`（200）

#### Scenario: 移动端静态资源可加载

- **WHEN** 浏览器访问 `/m/assets/index-xxx.js`（`dist-mobile/assets/` 下真实存在的文件）
- **THEN** 返回该文件内容（200，MIME 正确），而非桌面端 `dist/index.html` 或 404

#### Scenario: 移动端 history 子路由刷新

- **WHEN** 用户直接访问 `/m/login`（`dist-mobile/` 下无 `login` 文件，属前端 history 路由）
- **THEN** 后端返回 `dist-mobile/index.html`（200），由前端路由接管渲染

#### Scenario: dist-mobile 目录不存在

- **WHEN** 后端启动但 `dist-mobile/` 目录不存在
- **THEN** 后端跳过移动端托管，其他路由（含 `/api/*`、桌面 `/`）行为不变
