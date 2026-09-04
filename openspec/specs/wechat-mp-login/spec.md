# Spec: wechat-mp-login

微信小程序登录链路：后端 `mp-login` 接口、User 的 OpenID 关联、wechat 配置段、零第三方依赖的 jscode2session 调用，以及与账号密码登录等价的 JWT 签发。

## Requirements

### Requirement: 微信小程序登录接口存在

后端 MUST 提供 `POST /api/auth/mp-login` 接口（位于公开路由组，无需 token 即可访问），接收微信小程序 `wx.login` 颁发的 `code`，调用微信 `jscode2session` 拿到 `openid` 与 `session_key`，根据 `openid` 查找或创建 `User`，签发与账号密码登录**完全相同**的 JWT（同一签名密钥、同一 claims 结构、同一过期时间）。

#### Scenario: 小程序首次登录

- **WHEN** 客户端 POST `/api/auth/mp-login` body 含 `{ code: "<wx.login code>" }`，且后端 `wechat.appid` + `wechat.secret` 已配置
- **THEN** 后端调微信 `jscode2session` 成功，拿到 `openid`；若该 `openid` 对应的 `User` 不存在则创建之；返回 `{ token, user, permissions }`，结构同 `POST /api/auth/login`

#### Scenario: 同一 openid 二次登录

- **WHEN** 同一 `openid` 再次提交新 `code` 登录
- **THEN** 后端找到既有 `User`（不再创建），签发新 token，旧 token 在版本号未变前仍有效（与现有登录行为一致）

#### Scenario: code 失效或非法

- **WHEN** 提交的 `code` 已被微信过期或非法（后端 wechat 段已配置）
- **THEN** 微信侧返回 `errcode != 0`，后端响应 400 + `{ code: 400, message: "微信登录失败：<errcode>" }`，不签发 token

> **状态码口径：** 「wechat 段未配置」是**服务端配置缺失**，返回 **500**；「code 非法/过期」是
> **客户端参数问题**，返回 **400**。两者不可混用——把一个当成另一个会让排查方向跑偏
> （配置缺失去查 code，或 code 失效去查配置）。

### Requirement: User 模型新增 OpenID 字段

`User` 模型 MUST 新增 `OpenID` 字段（`string`，可空、唯一索引），用于关联微信小程序用户。`username/password` 登录方式保留不动，`OpenID` 为空的用户走原登录链路；`OpenID` 非空的用户除可走原登录外，可被 `mp-login` 命中。

#### Scenario: 自动迁移加 openid 列

- **WHEN** 后端首次启动且 `User` 表已存在
- **THEN** GORM AutoMigrate 添加 `openid` 列（含唯一索引），老用户该列为 NULL，登录方式不变

#### Scenario: 老账号无 openid 走原登录

- **WHEN** 老用户（`openid IS NULL`）提交 username/password 登录
- **THEN** 走原 `/api/auth/login` 链路，与变更前行为一致

#### Scenario: openid 唯一性

- **WHEN** `mp-login` 拿到一个已绑定既有 User 的 openid
- **THEN** 命中该 User，不会创建重复账号；数据库唯一索引兜底，并发场景下重复插入会失败

### Requirement: wechat 配置段

后端 `AppConfig` MUST 新增 `Wechat` 段，包含 `AppID` 与 `Secret` 两个字段，支持默认值 + YAML 覆盖；默认值为空串（开发态可填测试号），生产部署 MUST 在 `config.yaml` 配置真实值。该段不支持环境变量覆盖（与 `app` 品牌段一致，避免密钥通过 env 泄露到子进程）。

#### Scenario: 配置 wechat 段

- **WHEN** `config.yaml` 填写 `wechat.appid` 与 `wechat.secret`
- **THEN** 后端启动后 `mp-login` 调微信接口使用该值，登录链路正常

#### Scenario: 未配置 wechat 段

- **WHEN** `config.yaml` 未填 `wechat` 段，或字段为空
- **THEN** 后端启动正常，但 `mp-login` 返回 500 + 含「未配置」字样的明确指引；不影响其他接口，基座零配置仍可开箱启动

#### Scenario: wechat 段不受 env 覆盖

- **WHEN** 设置环境变量 `WECHAT_APPID`
- **THEN** `AppConfig.Wechat.AppID` 不受影响，仍取 config.yaml 或默认值

### Requirement: jscode2session 调用零第三方依赖

后端调用微信 `jscode2session` MUST 使用标准库 `net/http` + `encoding/json`，MUST NOT 引入 `github.com/wechat-...` 之类第三方 SDK，避免基座引入非通用业务依赖。调用封装在 `backend/utils/wechat.go`（或同级位置）的 `JsCode2Session(appid, secret, code)` 函数中。

#### Scenario: 成功调用

- **WHEN** `JsCode2Session` 收到合法 appid/secret/code
- **THEN** 返回 `{ openid, session_key, unionid? }`，调用方据此签发 JWT

#### Scenario: 微信侧错误

- **WHEN** 微信返回 `errcode != 0`（如 code 失效）
- **THEN** 函数返回 Go 错误，错误信息含 errcode 与 errmsg，调用方据此响应 400

#### Scenario: 依赖清单不变（后端侧）

- **WHEN** 本能力落地后检查 `deps.yaml` 的 `backend:` 段
- **THEN** 无新增登记项（未引入第三方后端依赖）

### Requirement: 签发的 JWT 与账号密码登录等价

`mp-login` 签发的 JWT MUST 与 `/api/auth/login` 签发的 JWT 在 claims 结构、签名密钥、过期时间上完全一致；`User` 是同一种实体，权限码数组、`isAdmin` 等元数据也走同一查询路径。后续中间件（`JWTAuth`、`PermissionRequired`、`AdminRequired`）MUST NOT 区分 token 来自哪种登录方式。

#### Scenario: mp-login token 调受保护接口

- **WHEN** 用 `mp-login` 拿到的 token 调 `GET /api/users`（需 `users:view` 权限）
- **THEN** 与 username/password 登录拿到的 token 走相同的中间件链路，权限校验结果一致

#### Scenario: 与 auth-token-lifecycle 协同

- **WHEN** `auth-token-lifecycle` 落地后，mp-login 签发的 token 也带 `Ver` 字段
- **THEN** 该 token 受版本吊销约束；用户登出或改密时版本号递增，mp-login token 与原登录 token 同步失效

#### Scenario: 路由组无差异

- **WHEN** 后端启动并枚举公开路由组
- **THEN** `/api/auth/login` 与 `/api/auth/mp-login` 同在公开组（无需 token）；`/api/auth/logout`、`/api/auth/refresh` 在受保护组

### Requirement: 冒烟测试覆盖 mp-login

`backend/scripts/smoke.sh` MUST 扩展冒烟用例：在 admin 登录 + 受保护路由命中之外，新增 `mp-login` 可达性断言——用一个占位 code 调用未配置 wechat 段的后端，断言响应 `code === 500` 且 message 含「未配置」字样，证明接口已注册且配置缺失时给出明确指引。

#### Scenario: mp-login 接口可达性冒烟

- **WHEN** 执行 `make smoke`
- **THEN** 包含一步 `curl -sf -X POST /api/auth/mp-login -d '{"code":"smoke-test"}'`，断言响应 `code === 500` 且含「未配置」指引；若接口未注册则会是 404，冒烟失败

#### Scenario: 既有冒烟用例不被破坏

- **WHEN** 扩展冒烟后执行 `make smoke`
- **THEN** 原有 admin 登录 + 受保护路由 + 403/401 用例全部仍通过
