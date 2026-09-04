# fix-mobile-deploy-runtime — Tasks

## 1. 移动端构建与路由按 /m/ 子路径对齐

- [x] 1.1 在 `mobile/vite.config.js` 用 `defineConfig(({ command }) => ...)`：`build` 时 `base: '/m/'`，`serve`（dev）时保持 `base: '/'`，不改现有 dev 体验
- [x] 1.2 将 `mobile/src/router/index.js` 的 `createWebHistory()` 改为 `createWebHistory(import.meta.env.BASE_URL)`，使 router base 与 vite base 始终同源（注释注明：BASE_URL 由 vite base 注入，只允许经 vite base 一处配置）
- [x] 1.3 在 mobile 执行 `npm run build`，确认 `mobile/dist/index.html` 的 script/link 引用以 `/m/assets/` 开头（对应 spec：产物资源路径带 /m/ 前缀）

## 2. 后端 /m/ 子路径 SPA 托管

- [x] 2.1 在 `backend/router/serve.go` 移除 `r.Static("/m/", "./dist-mobile")`，改为 `hasMobile` 时注册 `/m/*filepath` 处理器
- [x] 2.2 处理器实现：拼接 `dist-mobile/` 前先 `path.Clean` 并校验无目录穿越（拒绝 `..`/绝对路径）；请求指向目录（如 `/m/`）时追加 `index.html`；目标文件存在则 `c.File`，否则回退返回 `dist-mobile/index.html`
- [x] 2.3 保留桌面 NoRoute 对 `/m/`、`/api/`、`/static/` 前缀的排除逻辑（防 HTML 泄漏给未注册资源请求），并跑 `cd backend && go vet ./...` 确认无编译问题

## 3. 打包脚本补齐品牌资源

- [x] 3.1 在 `scripts/package.sh` 组装阶段新增拷贝 `$ROOT_DIR/backend/static` → `$DEPLOY_DIR/static`
- [x] 3.2 校验：执行 `make package`（本地平台）后 `deploy/` 下存在 `static/logo.png`

## 4. 部署态验收（手动冒烟）

- [x] 4.1 在 `deploy/` 下执行 `./start.sh` 启动后端，断言 `curl -i http://localhost:8080/m/` 返回 200 且为移动端 HTML
- [x] 4.2 从 `deploy/dist-mobile/index.html` 取资源路径（如 `/m/assets/*.js`）`curl -I`，断言 200 且 content-type 为 `text/javascript`/`text/css`（非桌面 HTML）
- [x] 4.3 断言 `curl -i http://localhost:8080/m/login` 返回 200 + `dist-mobile/index.html`（history 子路由回退）
- [x] 4.4 断言 `curl -i http://localhost:8080/static/logo.png` 返回 200；回归桌面端 `curl -i http://localhost:8080/` 与 `curl -i -X POST http://localhost:8080/api/auth/login` 不受影响
- [ ] 4.5 浏览器实测：手机视口访问 `/m/` 可渲染登录页并完成登录、页面内跳转正常
