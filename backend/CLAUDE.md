# backend/CLAUDE.md — 后端域宪法

AI 在 `backend/` 下编写 Go 代码时 MUST 遵守本文件。跨端通用规则见根 [`AGENTS.md`](../AGENTS.md)。

---

## 1. 三层分层铁律

后端 MUST 采用 `controller → service → model` 三层，信号流固定为：

```
HTTP → router → middleware(JWT/RBAC/日志) → controller → service → model(GORM) → DB
```

| 层 | 职责 | 禁止事项 |
|---|---|---|
| `controllers/` | 解析请求参数、参数校验、调用 service、组装响应 | ❌ 不写业务逻辑、❌ 不直接操作 GORM |
| `services/` | 业务逻辑、事务、数据组装、跨模型操作 | ❌ 不触碰 `gin.Context`、❌ 不处理 HTTP 响应 |
| `models/` | 纯数据结构 + GORM tag + 表关联 | ❌ 不写业务方法 |

**新增业务模块时**：`make gen name=<模块名>`（或参考 `_example/`），再填充业务逻辑。

> 这些铁律已被 `internal/guard/` 的 guard 测试编译成静态检查，违反即 `make test` 失败。

## 2. 统一响应协议

所有接口 MUST 返回 `{code, message, data}`，分页 MUST 使用 `{list, total, page, page_size}`。

| code | HTTP | 含义 |
|---|---|---|
| 200 | 200 | 成功 |
| 400 | 200 | 参数错误/业务校验失败 |
| 401 | 401 | 未认证（token 缺失/无效/过期） |
| 403 | 403 | 无权限/账号禁用 |
| 404 | 200 | 资源不存在 |
| 500 | 200 | 服务器内部错误 |

响应统一通过 `utils/response.go` 的 `Success` / `Fail` / `SuccessPage` / `Unauthorized` / `Forbidden` 方法返回，禁止手写 `c.JSON` 拼响应。

## 3. RBAC 权限接线

- 受保护路由 MUST 按权限码挂 `middleware.PermissionRequired("<资源>:<动作>")`。
- 仅超管接口挂 `middleware.AdminRequired()`。
- 权限码的资源名用**复数**（`assets:view`），命名规则见 [`docs/map.md`](../docs/map.md) 的「新模块命名约定」。
- 新增模块时，权限码由 `make gen` **自动注册**进 `database/database.go` 的 `initBaseData`
  权限声明块（`【gen:permissions】` 锚点内），无需手工补。
- 手工新增受保护路由时，MUST 同步在 `initBaseData` 的权限声明块注册该权限码，否则 guard 测试
  `Test_PermissionCodesRegisteredInBaseData` 会失败。漏注册的后果：非管理员用户请求该接口返回
  403，且权限管理界面看不到该码，管理员无法在界面上授权自救。
- `initBaseData` 的权限初始化按 `code` 幂等补齐（已存在则跳过，不覆盖），故新增权限码对已有
  数据库同样生效，无需删库重建。`Sort` 在运行时按当前最大值递增计算，MUST NOT 在声明里写死。

## 4. 命名规范

- Go：包名小写单数；结构体大驼峰；方法大驼峰；文件 snake_case（如 `user_service.go`）。
- 字段 JSON tag 统一 snake_case。

## 5. 数据库约定

- 所有模型 MUST 内嵌 `models.BaseModel`。
- 新表 MUST 加入 `database/database.go` 的 `AutoMigrate` 列表（生成器自动处理，或手动在 `【gen:migrate】` 锚点后插入）。
- 基础数据初始化统一在 `database.go` 的 `initBaseData` 中，MUST 保持幂等（先 `Count` 再 `Create`）。

## 6. 错误处理

- 错误 MUST 显式处理，禁止忽略。
- 业务错误通过 `errors.New` 返回中文消息，由 controller 转成响应。
- 数据库记录不存在统一用 `errors.New("xxx不存在")`。
