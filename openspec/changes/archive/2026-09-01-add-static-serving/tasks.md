# Tasks: add-static-serving

## 1. 后端静态托管

- [x] 1.1 新增 `backend/router/serve.go`，实现 `setupStaticServing(r *gin.Engine)`：`os.Stat` 判断 dist/dist-mobile 存在则 `r.Static` 托管
- [x] 1.2 实现 SPA 回退 `r.NoRoute`：排除 `/api/`、`/static/` 前缀，其余返回 `dist/index.html`；`/api/` 未命中返回 JSON 404
- [x] 1.3 `backend/router/router.go` 在 API 路由注册后调用 `setupStaticServing(r)`

## 2. 文档与验证

- [x] 2.1 更新 `docs/目录结构约定.md`：说明 dist/dist-mobile 部署目录与降级行为
- [x] 2.2 跑 `go build` + `make test` + `make smoke`，确认纯 API 模式（无 dist）行为不变
- [x] 2.3 手工验证：临时建 `dist/index.html`，启动后访问 `/` 返回该文件、访问 `/api/不存在` 返回 JSON 404、访问 `/static/logo.png` 仍正常
