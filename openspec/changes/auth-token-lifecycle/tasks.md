# Tasks: auth-token-lifecycle

> **开工前置**：`login-brand-visual`（进行中）正在改写两端的 `Login.vue` / `request.js` /
> `stores/app.js`，与本 change 改动面高度重叠。MUST 等其归档后再开始，否则合并冲突成本高。

## 1. 前置：先把红灯装好（改鉴权核心前必须完成）

- [ ] 1.1 `backend/scripts/smoke.sh`：抽出「返回 http_code 与 body 分离」的请求函数，
      替换 `curl -sf`——`-f` 会在遇到 401/403 时直接非零退出，无法断言「期望 401」（D10）
- [ ] 1.2 断言：无 token 访问 `/api/users` → `401`
- [ ] 1.3 断言：无效 token（如 `Bearer garbage`）访问 `/api/auth/info` → `401`
- [ ] 1.4 断言：把 refresh token 当作 access token 使用 → `401`（D3 Type 校验）
- [ ] 1.5 断言：登出后用旧 token 访问 `/api/auth/info` → `401`（本 change 的核心回归）
- [ ] 1.6 验证红灯有效性：在当前（未改造）代码上跑 `make smoke`，确认 1.2/1.3 通过而
      1.5 **按预期报红**——证明断言真的会失败，而不是恒真

## 2. 后端：令牌版本吊销

- [ ] 2.1 `backend/models/user.go`：`User` 新增 `TokenVersion int`（`gorm:"default:0" json:"-"`）
- [ ] 2.2 `backend/utils/jwt.go`：`Claims` 新增 `Ver int` 与 `Type string`（`access` / `refresh`）（D3）
- [ ] 2.3 `GenerateToken` 签名扩展为接收 `ver`，签发 `Type = "access"` 的令牌
- [ ] 2.4 `backend/middleware/jwt.go`：`JWTAuth` 解析后按 `claims.UserID` 查询 user，
      比对 `claims.Ver != user.TokenVersion` 时返回 `401`（D1）
- [ ] 2.5 `JWTAuth`：用 DB 的 `user.IsAdmin` 覆盖 `claims.IsAdmin` 再写入 context（D5）
- [ ] 2.6 `JWTAuth`：把查到的 user 写入 context（`c.Set("user", user)`）供下游复用（D4）
- [ ] 2.7 `backend/middleware/permission.go`：`PermissionRequired` 复用 context 中的 user，
      去掉重复的 `Preload("Roles").First(&user)` 查询（D4）
- [ ] 2.8 确认 `database.go` 的 `AutoMigrate` 无需改动（`&models.User{}` 已在迁移列表中）

## 3. 后端：刷新令牌

- [ ] 3.1 `backend/config/config.go`：新增 `jwt.refresh_expire_seconds`（默认 `604800`），
      同步默认值 → YAML → 环境变量（`BASE_BACKEND_JWT_REFRESH_EXPIRE_SECONDS`）三层链，
      并更新 `config.example.yaml`
- [ ] 3.2 `utils/jwt.go`：新增 `GenerateRefreshToken`，签发 `Type = "refresh"` 且有效期取
      `refresh_expire_seconds` 的令牌（D3）
- [ ] 3.3 `services/auth_service.go`：新增 `RefreshAccessToken(refreshToken string)`，
      依次校验签名、`Type == "refresh"`、`Ver` 与 DB 一致，通过后颁发新 access（D3）
- [ ] 3.4 `services/auth_service.go`：`Login` 的 `LoginResult` 新增 `RefreshToken` 字段，
      字段名沿用 `token` + 新增 `refresh_token`，**不改名**（D2，否则会打断 smoke.sh:57 与前端存储层）
- [ ] 3.5 `backend/controllers/auth.go`：新增 `Refresh` 端点
- [ ] 3.6 `backend/router/router.go`：注册 `POST /api/auth/refresh`（**公开**端点，
      刷新时 access 已过期，不能挂 `JWTAuth`）

## 4. 后端：登出与改密

- [ ] 4.1 `services/auth_service.go`：新增 `Logout(userID uint)`，递增该用户的 `token_version`（D1）
- [ ] 4.2 `controllers/auth.go`：`Logout` 从 context 取 `userID` 调用 service，替换现有空实现
- [ ] 4.3 `router.go`：将 `POST /api/auth/logout` 从公开 `auth` 组**移入 `protected` 组**
      （吊销谁必须先知道是谁，D8）
- [ ] 4.4 `services/auth_service.go`：`ChangePassword` 在**同一次** `Updates` 中同时更新
      `password` 与 `token_version`，避免两次写的中间态（D9）

## 5. 前端与移动端：无感续期与登出

- [ ] 5.1 `frontend/src/utils/request.js`：401 分支改为无感刷新——对非登录页请求、
      且非 refresh 请求本身的 401，发起 `POST /api/auth/refresh` 后重放原请求（D6）
- [ ] 5.2 `request.js`：实现**单飞**——刷新进行中的并发 401 挂入等待队列，刷新成功后统一重放；
      MUST NOT 为每个 401 各发一次 refresh（D6）
- [ ] 5.3 `request.js`：刷新失败时清本地并跳转 `/login?redirect=<当前完整路径>`，
      不再弹「确认/取消」二选一（D7）
- [ ] 5.4 `frontend/src/views/Login.vue`：登录成功后同时持久化 `token` 与 `refresh_token`
- [ ] 5.5 `frontend/src/router/index.js`：登录页读取 `redirect` 查询参数，登录成功后回跳
- [ ] 5.6 `frontend/src/stores/app.js`：`logout()` 改为**先**调用 `/api/auth/logout` 再清本地；
      接口失败 MUST NOT 阻断登出（网络不通时也必须能退出去）（D8）
- [ ] 5.7 `stores/app.js`：`logout()` 清理 `refresh_token`
- [ ] 5.8 `frontend/src/layout/Layout.vue`：`handleLogout` 改走 store 的 `logout()` action，
      不再直接清本地存储
- [ ] 5.9 `mobile/src/` 同步 5.1–5.7（`utils/request.js`、`views/Login.vue`、
      `stores/app.js`、`router`），两端行为 MUST 一致（D13）

## 6. 端到端验证

- [ ] 6.1 `make test` 全绿、`make lint` 无新增告警
- [ ] 6.2 **新登录** → 登出 → 用该 token 访问 `/api/auth/info` → `401`
      （注意：必须用**新登录**的令牌验证，旧令牌按 D11 仍然有效）
- [ ] 6.3 修改密码 → 旧 token 在下一次请求时 → `401`
- [ ] 6.4 无 token 调用 `/api/auth/logout` → `401`（确认移组生效）
- [ ] 6.5 并发场景：打开列表页（同时发出多个请求）且 access 已过期 →
      只发出**一次** refresh，全部请求重放成功（D6 单飞）
- [ ] 6.6 把 refresh token 直接当 access token 用 → `401`（D3 Type 校验）
- [ ] 6.7 管理员被降权后，其旧 token 不再直通 `AdminRequired`（D5）
- [ ] 6.8 `make smoke` 全绿（含 task 1 新增的全部断言）
- [ ] 6.9 移动端重复 6.2 / 6.3 验证（D13）
- [ ] 6.10 平滑升级验证：本 change 部署前签发的旧令牌在部署后仍可用（D11，属预期行为，确认未被破坏）
