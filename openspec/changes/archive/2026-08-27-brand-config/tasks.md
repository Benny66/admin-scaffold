# Tasks: brand-config

## 1. 后端：配置字段与静态服务

- [x] 1.1 `backend/config/config.go` 的 `AppConfig` 新增 `logo`、`favicon` 字段（`footer` 已存在），补默认值（`logo` 默认 `"logo.png"`）、YAML 解析、环境变量覆盖
- [x] 1.2 `backend/router/router.go` 注册静态服务 `r.Static("/static", "./static")`
- [x] 1.3 创建 `backend/static/` 目录，生成一张默认占位 logo 到 `backend/static/logo.png` 并提交进版本库（`favicon` 缺省回退该 logo）
- [x] 1.4 `backend/controllers/system.go` 的 `GetSystemInfo` 扩展返回 `logo`/`favicon`（拼 `/static/<file>` 路径，favicon 缺省回退 logo）/`footer`
- [x] 1.5 `backend/config/config.example.yaml` 补 `logo`/`favicon` 示例

## 2. 前端：渲染与 favicon 动态设置

- [x] 2.1 `frontend/src/stores/app.js` 扩展 state（logo/favicon/footer）+ setter + `fetchSystemInfo` 拉取并动态设置 favicon
- [x] 2.2 `frontend/src/layout/Layout.vue` 侧边栏改为 `<img>` logo（无则回退文字），底部渲染 footer
- [x] 2.3 `frontend/vite.config.js` 的 dev proxy 加 `/static` → `http://localhost:8080`

## 3. 移动端：补齐链路与渲染

- [x] 3.1 `mobile/src/stores/app.js` 补 `fetchSystemInfo` 与 logo/footer 字段（与前端同构）
- [x] 3.2 `mobile/src/views/Home.vue` 挂载时拉取系统信息，navbar 展示 logo
- [x] 3.3 `mobile/src/views/Login.vue` header 展示 logo
- [x] 3.4 `mobile/vite.config.js` 的 dev proxy 加 `/static`

## 4. 收口与验证

- [x] 4.1 `docs/配置体系.md` 补品牌四要素说明 + 生产环境 `/static` 转发部署提示
- [x] 4.2 跑 `make test` + `make smoke` + 前后端 `npm run build`，确认全绿
- [x] 4.3 手工验证：临时在 `backend/static/` 放一个测试 logo，启动后访问 `/static/<file>` 返回 200，前端侧边栏显示图片
