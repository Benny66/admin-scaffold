# AGENTS.md — AI 开发规范（宪法）

本文件是 AI 在本仓库中生成、修改代码时必须遵守的硬性约束。违反以下规则的代码视为不合格。

---

## 1. 后端三层分层铁律

后端 MUST 采用 `controller → service → model` 三层，信号流固定为：

```
HTTP → router → middleware(JWT/RBAC/日志) → controller → service → model(GORM) → DB
```

| 层 | 职责 | 禁止事项 |
|---|---|---|
| `controllers/` | 解析请求参数、参数校验、调用 service、组装响应 | ❌ 不写业务逻辑、❌ 不直接操作 GORM |
| `services/` | 业务逻辑、事务、数据组装、跨模型操作 | ❌ 不触碰 `gin.Context`、❌ 不处理 HTTP 响应 |
| `models/` | 纯数据结构 + GORM tag + 表关联 | ❌ 不写业务方法 |

**新增业务模块时**：先在 `services/` 建 service，再在 `controllers/` 建 controller，最后在 `router/` 注册路由。

## 2. 统一响应协议

所有接口 MUST 返回 `{code, message, data}`，分页 MUST 使用 `{list, total, page, page_size}`。

错误码语义固定：

| code | HTTP | 含义 |
|---|---|---|
| 200 | 200 | 成功 |
| 400 | 200 | 参数错误/业务校验失败 |
| 401 | 401 | 未认证（token 缺失/无效/过期） |
| 403 | 403 | 无权限/账号禁用 |
| 500 | 200 | 服务器内部错误 |

响应统一通过 `utils/response.go` 的 `Success` / `Fail` / `SuccessPage` / `Unauthorized` / `Forbidden` 方法返回，禁止手写 `c.JSON` 拼响应。

## 3. 前端命名与目录约定

- 状态管理目录 MUST 统一为 `stores/`（禁止 `store/` 单复数混用）。
- 页面按业务域分目录：`views/<domain>/`，公共组件放 `components/`。
- API 定义集中在 `api/`，请求封装统一走 `utils/request.js`。
- 路径别名 `@` 指向 `src/`。

## 4. 请求封装规范

前端所有 HTTP 请求 MUST 走 `utils/request.js` 的 axios 实例，MUST 自动附带 `Authorization: Bearer <token>`，响应拦截器 MUST 统一处理 401（清 token 跳登录）、403（无权限提示）、blob（直接返回）。

## 5. 禁止引入的依赖

以下依赖属于业务/商业化专属，基座 MUST NOT 引入：

- 后端：`excelize`（导入导出）、资产编码/打印相关业务包
- 前端：`echarts`、`jsbarcode`、`qrcode`
- 移动端：`html5-qrcode`

新增依赖 MUST 是可跨项目复用的通用依赖。

## 6. 命名规范

- Go：包名小写单数；结构体大驼峰；方法大驼峰；文件 snake_case（如 `user_service.go`）。
- 前端：组件文件 PascalCase（如 `UserList.vue`）；JS 模块 camelCase；目录 kebab-case。
- 字段 JSON tag 统一 snake_case（如 `real_name`、`created_at`）。

## 7. 数据库约定

- 所有模型 MUST 内嵌 `models.BaseModel`（提供 `id`、`created_at`、`updated_at`）。
- 新表 MUST 加入 `database/database.go` 的 `AutoMigrate` 列表。
- 基础数据初始化统一在 `database.go` 的 `initBaseData` 中，MUST 保持幂等（先 `Count` 再 `Create`）。

## 8. 干净性

MUST NOT 提交：`*.db`、`*.exe`、`dist/`、`node_modules/`、`*.orig`、`*.rej`、`vite.config.js.timestamp-*.mjs`。这些已列入 `.gitignore`。
