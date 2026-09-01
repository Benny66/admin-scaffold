# Spec: auth-token-lifecycle

让服务端能主动推翻已签发的令牌（登出、改密、降权立即生效），并让客户端在令牌过期时无感续期，
同时把「登出后旧令牌必须失效」编译成冒烟回归断言。

## ADDED Requirements

### Requirement: 登出必须使令牌立即失效

用户调用登出接口后，其名下全部已签发令牌 MUST 立即失效，继续使用 MUST 返回 `401`。
登出 MUST 同时使访问令牌与刷新令牌失效。

#### Scenario: 登出后访问受保护接口

- **WHEN** 用户登录、登出，随后用原 token 请求 `/api/auth/info`
- **THEN** 返回 `401`，而非 `200`

#### Scenario: 登出后尝试刷新

- **WHEN** 用户登出后用原 refresh token 请求 `/api/auth/refresh`
- **THEN** 返回 `401`，不颁发新令牌

#### Scenario: 不得影响其他用户

- **WHEN** 用户 A 登出
- **THEN** 用户 B 的令牌不受影响，仍可正常访问

### Requirement: 修改密码必须使令牌立即失效

修改密码成功后，该用户旧令牌在下一次请求时 MUST 返回 `401`；
修改密码的当次请求 MUST 仍能正常返回 `200`。

#### Scenario: 改密后旧令牌失效

- **WHEN** 用户修改密码，随后用原 token 请求 `/api/auth/info`
- **THEN** 返回 `401`

#### Scenario: 改密请求自身成功返回

- **WHEN** 用户提交正确的原密码与新密码
- **THEN** 该请求返回 `200`（版本号在处理过程中递增，不影响本次响应）

### Requirement: 管理员降权必须立即生效

鉴权中间件 MUST 以数据库中的 `is_admin` 为准判定管理员身份，MUST NOT 信任令牌 claims 中的值。

#### Scenario: 管理员被降权

- **WHEN** 某用户的 `is_admin` 被改为 `false`，其旧 token 请求受 `AdminRequired` 保护的接口
- **THEN** 返回 `403`

#### Scenario: 用户被提升为管理员

- **WHEN** 某用户的 `is_admin` 被改为 `true`，其旧 token 请求受 `AdminRequired` 保护的接口
- **THEN** 返回 `200`（无需重新登录）

### Requirement: 令牌必须携带版本声明并逐请求校验

签发的令牌 MUST 携带版本声明，鉴权中间件 MUST 逐请求将其与用户记录中的版本比对，
不一致时 MUST 返回 `401`。

#### Scenario: 版本一致

- **WHEN** 令牌中的版本与数据库中该用户的版本相同
- **THEN** 请求正常放行

#### Scenario: 版本不一致

- **WHEN** 令牌中的版本落后于数据库中该用户的版本
- **THEN** 返回 `401`

#### Scenario: 令牌未携带版本声明（旧令牌）

- **WHEN** 令牌中没有版本声明
- **THEN** 按 `0` 处理——与新增字段的默认值一致，因此**旧令牌仍然有效**（见「既有令牌必须平滑过渡」）

### Requirement: 登录必须同时颁发访问令牌与刷新令牌

登录成功 MUST 同时返回访问令牌与刷新令牌，访问令牌的响应字段名 MUST 保持为 `token`，
刷新令牌 MUST 使用新字段名 `refresh_token`。

#### Scenario: 登录响应包含两个令牌

- **WHEN** 用户以正确凭据登录
- **THEN** 响应体同时包含 `token` 与 `refresh_token`

#### Scenario: 字段名保持向后兼容

- **WHEN** 检查登录响应与既有消费方（冒烟脚本、前端本地存储）
- **THEN** 访问令牌字段名仍为 `token`，既有提取逻辑无需修改

### Requirement: 刷新令牌只能用于刷新端点

刷新端点 MUST 拒绝令牌类型不为 `refresh` 的请求；
鉴权中间件 MUST 拒绝把刷新令牌当作访问令牌使用的请求。

#### Scenario: 刷新令牌当作访问令牌

- **WHEN** 用 refresh token 作为 `Bearer` 请求 `/api/auth/info`
- **THEN** 返回 `401`

#### Scenario: 访问令牌用于刷新端点

- **WHEN** 用 access token 请求 `/api/auth/refresh`
- **THEN** 返回 `401`

#### Scenario: 刷新令牌用于刷新端点

- **WHEN** 用有效的 refresh token 请求 `/api/auth/refresh`
- **THEN** 返回新的访问令牌

### Requirement: 刷新令牌必须受版本约束

刷新端点 MUST 校验刷新令牌中的版本与数据库中该用户的版本一致，不一致 MUST 返回 `401`。

#### Scenario: 登出后刷新

- **WHEN** 用户登出后，用原 refresh token 请求刷新
- **THEN** 返回 `401`（版本号已递增）

#### Scenario: 改密后刷新

- **WHEN** 用户改密后，用原 refresh token 请求刷新
- **THEN** 返回 `401`

#### Scenario: 正常刷新

- **WHEN** 用户未登出未改密，refresh token 在有效期内
- **THEN** 返回新的访问令牌

### Requirement: 前端必须在访问令牌过期时静默续期并重放

前端请求层 MUST 在收到 `401` 时自动发起刷新并重放原请求，用户 MUST NOT 感知到中断。
刷新请求自身的 `401` MUST NOT 触发再次刷新（防止无限循环）。
登录页声明了自行呈现错误的请求 MUST NOT 被拦截。

#### Scenario: 访问令牌过期后自动续期

- **WHEN** 用户在列表页操作，access token 恰好过期
- **THEN** 请求层静默刷新后重放，界面数据正常返回，用户无感知

#### Scenario: 刷新请求自身返回 401

- **WHEN** `/api/auth/refresh` 返回 `401`（refresh token 也失效）
- **THEN** 直接进入失败处理流程，MUST NOT 再次发起刷新

#### Scenario: 登录页的凭据错误

- **WHEN** 登录请求返回 `401`（用户名或密码错误）
- **THEN** 由调用方在表单内联展示，请求层 MUST NOT 弹「登录已过期」或触发刷新

### Requirement: 并发的 401 必须合并为一次刷新

当多个请求同时收到 `401` 时，请求层 MUST 只发起**一次**刷新请求，其余请求 MUST 排队等待
并在刷新完成后统一重放。

#### Scenario: 列表页并发请求同时过期

- **WHEN** 页面同时发出 3 个请求且 access token 已过期
- **THEN** 只发出 1 次 `/api/auth/refresh`，3 个请求全部重放成功

#### Scenario: 并发期间刷新失败

- **WHEN** 上述场景中刷新失败
- **THEN** 全部排队请求一并进入失败处理，跳登录，MUST NOT 重复弹窗

### Requirement: 刷新失败必须跳转登录并保留回跳位置

刷新失败时请求层 MUST 清除本地凭据并跳转登录页，MUST 携带当前完整路径作为回跳参数，
MUST NOT 提供「取消」选项（此时留在原地没有有意义的下文）。

#### Scenario: 刷新失败

- **WHEN** refresh token 失效且 access token 已过期
- **THEN** 清除本地凭据，跳转 `/login?redirect=<当前路径>`

#### Scenario: 登录后回跳

- **WHEN** 用户在 `/system/user` 被踢出，完成登录
- **THEN** 自动回跳到 `/system/user`

### Requirement: 登出接口必须要求身份，且前端登出不得因接口失败而阻断

登出端点 MUST 置于需要鉴权的路由组下，MUST 拒绝无令牌的请求。
前端登出 MUST 先调用登出接口再清理本地凭据，且接口失败 MUST NOT 阻断登出流程。

#### Scenario: 无令牌调用登出

- **WHEN** 不带 Authorization 头请求 `/api/auth/logout`
- **THEN** 返回 `401`

#### Scenario: 正常登出

- **WHEN** 用户点击退出登录且网络正常
- **THEN** 服务端递增版本号，前端清理本地凭据并跳转登录页

#### Scenario: 登出接口不可用

- **WHEN** 用户点击退出登录，但登出接口返回错误或网络不通
- **THEN** 前端仍清理本地凭据并跳转登录页（不能因为接口挂了就退不出去）

### Requirement: 冒烟必须覆盖令牌失效回归

冒烟脚本 MUST 断言非 `2xx` 的期望响应（当前 `curl -sf` 会在此类响应下直接退出），
并 MUST 包含「登出后旧令牌返回 401」这一核心回归断言。

#### Scenario: 冒烟覆盖鉴权失败路径

- **WHEN** 执行 `make smoke`
- **THEN** 断言无令牌访问受保护接口为 `401`、无效令牌为 `401`、登出后旧令牌为 `401`

#### Scenario: 红灯有效性已被验证

- **WHEN** 在未改造的代码上运行新增断言
- **THEN** 「登出后旧令牌为 401」按预期失败，证明断言不是恒真的

### Requirement: 既有令牌必须平滑过渡

版本字段的默认值 MUST 与「令牌未携带版本声明」时的取值保持一致，
使本能力上线前签发的令牌仍然有效，MUST NOT 造成全量用户被强制登出。

#### Scenario: 上线前签发的令牌

- **WHEN** 本能力部署前签发的令牌在部署后被使用
- **THEN** 校验通过，请求正常放行

#### Scenario: 项目选择强制全量重新登录

- **WHEN** 某项目希望上线后所有用户重新登录
- **THEN** 可在迁移中把用户的版本初始化为 `1`，使全部旧令牌失效

### Requirement: 两端行为一致且无新增第三方依赖

桌面端与移动端的续期与登出行为 MUST 一致；本能力 MUST NOT 引入新的第三方依赖，
新增配置项 MUST 同步登记到三层配置链与示例配置文件。

#### Scenario: 移动端同样无感续期

- **WHEN** 移动端 access token 过期
- **THEN** 同样静默刷新并重放，行为与桌面端一致

#### Scenario: 依赖清单不变

- **WHEN** 本能力落地后检查 `deps.yaml`
- **THEN** 无新增登记项（继续使用 `golang-jwt/jwt/v5`）

#### Scenario: 新配置项可通过三种方式覆盖

- **WHEN** 检查刷新令牌有效期的配置
- **THEN** 代码默认值、`config.yaml`、环境变量 `BASE_BACKEND_JWT_REFRESH_EXPIRE_SECONDS`
      三层均可覆盖，且 `config.example.yaml` 已登记
