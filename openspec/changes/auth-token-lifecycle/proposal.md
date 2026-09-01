# Proposal: auth-token-lifecycle

## Why

登出是一个空实现：

```44:46:backend/controllers/auth.go
func Logout(c *gin.Context) {
	utils.Success(c, nil)
}
```

配合 JWT 无吊销机制（`utils/jwt.go:27` 默认 3600s 有效期内无条件有效），形成三个连锁问题：

| 现象 | 用户视角 | 安全视角 |
|---|---|---|
| 点「退出登录」 | 界面回登录页，但旧 token 照常能调通所有接口 | 令牌无法吊销 |
| 修改密码后 | 旧 token 依然有效到自然过期（最长 1 小时） | 改密不踢线 |
| 1 小时到期 | 正在填的表单被弹窗打断 | — |

第三个还是个纯交互缺陷：`frontend/src/utils/request.js:74-79` 的 401 处理里，
**点「确认」和点「取消」都跳 `/login`**——取消按钮没有任何意义，正在编辑的内容也无条件丢失。

还有一处被同一根因掩盖的问题：`middleware/jwt.go:37` 把 `claims.IsAdmin` 写进 context，
于是管理员被降权后，旧 token 仍以 `isAdmin=true` 直通 `PermissionRequired` 与 `AdminRequired`，
直到自然过期。

根因是同一个：**令牌一经签发，在有效期内没有任何服务端状态可以推翻它。**

## What Changes

1. **令牌版本吊销**：`User` 新增 `TokenVersion` 字段，JWT claims 携带 `Ver`；`JWTAuth` 逐请求
   比对版本，不一致即 401。登出与修改密码时递增版本号。
2. **`isAdmin` 以数据库为准**：`JWTAuth` 用 DB 的 `user.IsAdmin` 覆盖 claims 中的值，降权立即生效。
3. **访问令牌 + 刷新令牌**：登录同时颁发两者，`POST /api/auth/refresh` 静默续期。
4. **前端无感续期**：`request.js` 在 401 时静默刷新并重放原请求，并发 401 单飞（只发一次刷新）。
   只有刷新失败才跳登录，且带 `?redirect=` 回跳。
5. **登出接口实装**：`Logout` 移入受保护路由组（需要身份才能吊销谁），递增版本号。
6. **冒烟测试扩展**（前置任务）：当前 `scripts/smoke.sh` 只断言 happy path，且 `curl -sf`
   遇到非 2xx 会直接退出——改鉴权核心却没有红灯护着，风险不对称。

**非目标（明确排除）**：不做登录限流与失败锁定、不做验证码、不改默认密钥
（`config.go:111` 的 `base-backend-secret-key-change-me`）。这些属于「登录安全」而非
「令牌生命周期」，另行立案。

## Capabilities

### New Capabilities

- `auth-token-lifecycle`: 令牌的签发、续期与吊销——服务端可主动使令牌失效、客户端可无感续期，
  以及「登出/改密/降权必须立即生效」的回归验证。

### Modified Capabilities

（无。）

## Impact

- 修改文件：`backend/models/user.go`、`backend/utils/jwt.go`、`backend/middleware/jwt.go`、
  `backend/middleware/permission.go`、`backend/services/auth_service.go`、
  `backend/controllers/auth.go`、`backend/router/router.go`、`backend/config/config.go`、
  `backend/config/config.example.yaml`、`backend/scripts/smoke.sh`、
  `frontend/src/utils/request.js`、`frontend/src/stores/app.js`、`frontend/src/layout/Layout.vue`、
  `frontend/src/views/Login.vue`、`mobile/src/utils/request.js`、`mobile/src/stores/app.js`、
  `mobile/src/views/Login.vue`。
- 无新第三方依赖（继续使用 `golang-jwt/jwt/v5`）。
- 新增配置项 `jwt.refresh_expire_seconds`（默认 604800），需同步三层配置链与 `config.example.yaml`。
- **破坏性**：`POST /api/auth/logout` 从公开组移入受保护组（吊销谁必须先知道是谁），
  无 token 调用将返回 401。
- 行为变化：登出/改密后令牌立即失效（这是本 change 的目的）；已登录用户的**旧令牌仍有效**
  （版本号缺省 0 与新增字段默认值 0 一致，属有意设计的平滑升级，见 design D11）。
- **前置依赖**：`login-brand-visual`（进行中）正在改 `Login.vue` / `request.js` / `stores/app.js`
  的两端文件，与本 change 高度重叠。**本 change 必须等其归档后再开工**，否则合并冲突成本高。
