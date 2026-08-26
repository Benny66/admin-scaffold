# API 协议

## 基础路径与认证

- 基础路径：`/api`
- 认证方式：请求头携带 `Authorization: Bearer <token>`
- 公开接口无需鉴权，受保护接口需 JWT。

## 统一响应结构

所有接口返回统一结构：

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| code | int | 业务状态码 |
| message | string | 提示消息 |
| data | any | 业务数据 |

## 分页结构

分页接口的 `data` 固定为：

```json
{
  "list": [],
  "total": 100,
  "page": 1,
  "page_size": 10
}
```

请求分页参数：`page`（页码，从 1 开始）、`page_size`（每页条数，默认 10，上限 100）。

## 状态码语义

| code | HTTP 状态码 | 含义 |
|---|---|---|
| 200 | 200 | 成功 |
| 400 | 200 | 参数错误 / 业务校验失败 |
| 401 | 401 | 未认证（token 缺失/无效/过期） |
| 403 | 403 | 无权限 / 账号禁用 |
| 404 | 200 | 资源不存在 |
| 500 | 200 | 服务器内部错误 |

> 注：业务错误（400/404/500）HTTP 状态码仍为 200，通过 `code` 字段区分；认证/鉴权错误（401/403）使用真实 HTTP 状态码，便于前端拦截器统一跳转。

## 后端实现约定

- 通过 `utils/response.go` 的方法返回，禁止手写 `c.JSON` 拼响应。
- 方法清单：`Success` / `SuccessWithMsg` / `Fail` / `SuccessPage` / `Unauthorized` / `Forbidden` / `FailWithStatus`。

## 前端消费约定

- 通过 `utils/request.js` 的 axios 实例发起请求。
- 响应拦截器统一处理：`code === 200` 返回 `res`；否则弹错误并 reject。
- HTTP 401 → 清 token 跳登录；403 → 提示无权限；blob → 直接返回。

## 主要接口清单

| 模块 | 方法 | 路径 |
|---|---|---|
| 登录 | POST | `/api/auth/login` |
| 退出 | POST | `/api/auth/logout` |
| 用户信息 | GET | `/api/auth/info` |
| 改密码 | PUT | `/api/auth/password` |
| 用户列表 | GET | `/api/users` |
| 角色列表 | GET | `/api/roles` |
| 权限列表 | GET | `/api/permissions` |
| 字典类型 | GET | `/api/dict/types` |
| 字典项 | GET | `/api/dict/items` |
| 操作日志 | GET | `/api/logs/operation` |
| 登录日志 | GET | `/api/logs/login` |
| 系统信息 | GET | `/api/system/info` |
