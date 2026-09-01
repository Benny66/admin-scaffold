# Design: auth-token-lifecycle

## 上下文

本 change 是「三洞修复」序列的第 3 个。

```
① 文档定位对齐        ✅ 已并入 frontend-rbac-enforcement 的 task 5
② frontend-rbac-enforcement   ✅ 制品就绪（待实施）
③ auth-token-lifecycle        ← 本 change，风险最高
```

> 注：原序列中的 ④「五件套按 `_example` 重写」经核实已取消（五件套与范例实质对齐、漂移极小）。

③ 改的是鉴权核心链路——出错会让整个系统不可用。因此本 change 把「先补红灯」放在第一位（task 1）。

**开工前置**：`login-brand-visual`（35/39，进行中）正在改写 `Login.vue` / `request.js` /
`stores/app.js` 的两端文件，与本 change 改动面高度重叠。MUST 等其归档后再开始。

---

## D1 · 吊销机制选型

| 方案 | 存储 | 多实例 | refresh 覆盖 | 结论 |
|---|---|---|---|---|
| **A. 令牌版本（`user.token_version`）** | 一个整数字段 | ✅ | ✅ 天然覆盖 | ✅ 选此 |
| B. 黑名单表 | 新表 + 过期清理 | ✅ | 需额外处理 | ❌ 多一张表要维护 |
| C. Redis 黑名单 | 外部依赖 | ✅ | 需额外处理 | ❌ 引入重依赖，违背开箱即用 |
| D. 进程内 map | 内存 | ❌ 多实例失效 | 需额外处理 | ❌ 不可水平扩展 |

选 A 的决定性理由：**它天然覆盖 refresh token**。B/C/D 都要为「刷新令牌如何吊销」单独设计，
而版本号机制下，refresh 与 access 携带同一个 `Ver`，递增一次两者同时失效。

## D2 · 字段命名：保留 `token`，新增 `refresh_token`

登录响应改为：

```json
{ "token": "...", "refresh_token": "...", "user": {...}, "permissions": [...] }
```

**不把 `token` 改名为 `access_token`**。原因：`backend/scripts/smoke.sh:57` 用
`sed -n 's/.*"token":"\([^"]*\)".*/\1/p'` 提取令牌，前端两端的 `localStorage.token` 也依赖此名。
改名会同时打断冒烟脚本与前端存储层，收益为零、破坏面却不小。

## D3 · 刷新令牌的形态与校验

`Claims` 新增两个字段：

```go
Ver  int    `json:"ver"`   // 令牌版本，与 user.token_version 比对
Type string `json:"typ"`   // "access" | "refresh"
```

`POST /api/auth/refresh`（公开端点，body 携带 `refresh_token`）校验三件事：

1. 签名有效且未过期
2. `Type == "refresh"` —— **防止 refresh token 被当作 access token 直接用**
3. `Ver` 与 DB 中 `user.token_version` 一致 —— 登出/改密后刷新不出新令牌

校验通过则颁发新的 access token（Ver 继承当前 DB 值）。

```
        客户端                                    服务端
          │                                         │
  access 过期 → 401                                 │
          │                                         │
          ├─ POST /auth/refresh {refresh_token} ───▶│
          │                                         ├─ 签名 ✓
          │                                         ├─ Type == refresh ✓
          │                                         └─ Ver == DB.token_version ✓
          │◀──────── { token: "<新 access>" } ──────┤
          │                                         │
          └─ 重放原请求（携带新 access）───────────▶│
                                                    └─ 200 ✓
```

## D4 · 查询成本：用上下文复用抵消

`JWTAuth` 原本只解析令牌、不查库（3 次查询全在 `PermissionRequired`：
`Preload("Roles").First(&user)` 2 次 + JOIN 权限 1 次）。

改造后：

| 路径 | 改造前 | 改造后 | 差值 |
|---|---|---|---|
| 有权限码的接口 | 0 + 3 | 1（查 user 校验版本）+ 2（复用 ctx user，仅补 Roles 关联） | **持平** |
| 无权限码的接口（`/auth/info` 等） | 0 | 1 | **+1 次主键查询** |

+1 次主键查询换来「登出/改密/降权立即生效」，代价可接受。若将来确实成为瓶颈，
可加一个 30s TTL 的 `userID → token_version` 内存缓存——**不在本 change 实现**，
因为多实例场景下缓存失效是另一个问题，不该在这里顺手引入。

## D5 · `isAdmin` 以数据库为准

`middleware/jwt.go:37` 当前写的是 `c.Set("isAdmin", claims.IsAdmin)`。
管理员被降权后，旧令牌仍以 `true` 直通到自然过期。

改为用本次查出的 `user.IsAdmin` 覆盖。零额外成本（D4 已经查了 user），
且消除了「降权后最长 1 小时的权限残留」。

## D6 · 前端无感续期与单飞

`request.js` 的 401 分支改为：

```
401 → 是登录页请求？─── 是 ──▶ 交给调用方（现有 skipGlobalError 机制）
        │
        否
        ▼
   已在刷新中？─── 是 ──▶ 挂入等待队列（单飞）
        │
        否
        ▼
   发起 refresh ──▶ 成功 → 更新 localStorage.token → 重放队列中的全部请求
        │
        └──────────▶ 失败 → 清本地 → 跳 /login?redirect=<当前路径>
```

**单飞（single-flight）是必需的，不是优化**：一个列表页会并发发出多个请求，
若不合并，会同时发出 N 次 refresh；在刷新令牌轮换场景下会产生竞态，导致部分请求拿到已失效的令牌。

## D7 · 401 交互：去掉没有意义的「取消」

现状 `request.js:74-79`：

```js
.then(() => { router.push('/login') })
.catch(() => { router.push('/login') })   // 点取消也跳，取消按钮形同虚设
```

有了无感续期后，401 弹窗只在「刷新也失败」时出现，此时界面上的数据都已无法加载，
「留在原地」没有有意义的下文。故改为：

- **能续期** → 静默续期，用户完全无感（弹窗消失）
- **不能续期** → 直接跳 `/login?redirect=<当前路径>`，不再弹二选一
- 登录后按 `redirect` 回跳，正在编辑的页面位置不丢（内容本身不保证，但入口回来了）

## D8 · 登出接口必须移入受保护组

`router.go:24` 当前把 `/api/auth/logout` 挂在公开组。要递增某个用户的版本号，
服务端必须先知道是谁——所以 logout 语义上**必须携带身份**。

决策：移入 `protected` 组。前端 `logout()` 的调用顺序相应调整为
**先请求接口（此时 token 仍在），成功后清本地**；且接口失败 MUST NOT 阻断登出
（网络不通时也必须能退出去，见 spec 对应 requirement）。

## D9 · 改密递增版本号，且当前请求仍可正常返回

`services.ChangePassword` 在更新密码的**同一次** `Updates` 中递增 `token_version`，
避免两次写操作中间态。

注意时序：本次请求的 `JWTAuth` 已经通过，版本号递增发生在处理过程中，
因此当前请求能正常返回 200；而该令牌在**下一次请求**时失效。
前端现有「改密成功后强制登出」的逻辑（`Layout.vue:147-149`）与之一致。

## D10 · 冒烟必须先扩展（task 1）

当前 `scripts/smoke.sh` 有两个问题：

1. 只断言 happy path（`system/info` → login → `auth/info` code=200）。
2. `curl -sf` 的 `-f` 会让脚本在遇到 401/403 时**直接非零退出**——
   想断言「期望 401」就必须换一个不断言 HTTP 状态码的取值方式。

因此需要先抽出 `http_code` 与 `body` 分离的请求函数，再补断言。核心回归断言是：

```
登出后旧 token 访问 /api/auth/info → 401   ← 本 change 的验收条件
```

## D11 · 平滑升级：既有令牌不失效

新增字段默认 0，旧令牌没有 `ver` claim，反序列化后也是 0 —— 两者一致，校验通过。

这是**有意设计**：本 change 部署后已登录用户不会被强制踢出，
避免一次发布造成全量重新登录。若某项目接受全量踢出，
可在迁移脚本里把所有用户的 `token_version` 初始化为 1。

## D12 · access 有效期保持 3600 秒不变

缩短 access 能降低令牌被盗用的窗口，但会提高刷新频率、放大 D4 的 +1 查询成本。

决策：**保持默认值不变，只提供机制**。有效期是安全策略，属于具体项目的选择，
基座不应替业务做决定——项目可在 `config.yaml` 里自行调短。

---

## 风险

| 风险 | 缓解 |
|---|---|
| 改坏 `JWTAuth` 导致全站 401 | task 1 先补冒烟断言；task 6 逐条验证登出前后行为 |
| 与 `login-brand-visual` 合并冲突 | 开工前置：等该 change 归档 |
| 刷新竞态导致部分请求失败 | D6 单飞；task 6.5 并发场景验证 |
| `Logout` 移组后前端未同步调用顺序 | task 5.3 明确「先请求后清本地」，task 6.4 验证 |
| 两端（frontend / mobile）逻辑漂移 | D13：两端同步实现，spec 中列两端统一的 requirement |
| 旧令牌仍然有效（D11）被误认为没生效 | task 6.2 用**新登录**的令牌验证失效，而非旧令牌 |

## D13 · 两端同等覆盖

后端改动对两端同时生效。前端的无感续期逻辑 MUST 在 `frontend/` 与 `mobile/` 两端同步实现——
否则移动端用户会在改密后被踢，而桌面端不会，形成行为漂移。
两端 request.js 本就是互为副本的关系（如现有 401 处理），本 change 保持这一对称性。
