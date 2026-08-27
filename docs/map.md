# 导航地图 — 哪类代码在哪个文件

AI 定位代码时**先读本文件**，不要用 `find` 盲扫。这是「上下文最优」的关键：知道去哪读，就不用全局加载。

## 核心原则

- **范例**：新增业务模块只看 `backend/_example/`（黄金路径唯一范例）。
- **历史模块**：`views/system/` 五件套与 `backend/` 下对应的 system 模块是**历史代码，已互相漂移，禁止作为模仿对象**。

## 后端（backend/）

| 我要做 | 去这里 | 关键说明 |
|---|---|---|
| 新增一个业务模块 | `make gen name=<模块名>` | 从 `_example/` 生成三层骨架 |
| 理解分层范式 | `backend/_example/{models,services,controllers}/` | 唯一标准答案 |
| 加一条路由 | `backend/router/router.go` | 受保护路由挂 `PermissionRequired` |
| 建表/迁移 | `backend/database/database.go` | `AutoMigrate` + `initBaseData` |
| 鉴权中间件 | `backend/middleware/{jwt,permission,logger}.go` | JWT / RBAC / 操作日志 |
| 统一响应 | `backend/utils/response.go` | `Success`/`Fail`/`SuccessPage` 等 |
| 架构护栏 | `backend/internal/guard/guard_test.go` | 分层铁律等约束的静态检查 |

## 前端（frontend/src/）

| 我要做 | 去这里 | 关键说明 |
|---|---|---|
| 新增页面 | `views/<domain>/index.vue` | `make gen` 会自动生成 |
| 加接口定义 | `api/index.js` | 按模块分组追加 |
| 加路由/菜单 | `router/index.js` + `layout/Layout.vue` 的 `menus` | 两处都要改 |
| 状态管理 | `stores/app.js` | 统一 `stores/` |
| 请求封装 | `utils/request.js` | 拦截器统一处理 401/403 |

## 契约与规范

| 我要做 | 去这里 |
|---|---|
| 查字段名/响应形状 | `contracts/openapi.yaml`（唯一真相） |
| 后端规范 | `backend/CLAUDE.md` + `docs/后端分层规范.md` |
| 前端规范 | `frontend/CLAUDE.md` + `docs/代码规范.md` |
| 通用铁律 | `AGENTS.md`（根宪法） |
| API 协议 | `docs/API协议.md` |

## 验证入口

```bash
make test    # 后端全部测试（含 guard 护栏）
make lint    # go vet
make smoke   # 冒烟：启动→登录→命中受保护路由→断言
make gen name=<模块名>   # 生成新模块骨架
```
