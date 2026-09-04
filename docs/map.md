# 导航地图 — 哪类代码在哪个文件

AI 定位代码时**先读本文件**，不要用 `find` 盲扫。这是「上下文最优」的关键：知道去哪读，就不用全局加载。

## 核心原则

- **范例**：新增业务模块只看 `backend/_example/`（黄金路径唯一范例）。
- **历史模块**：`views/system/` 五件套与 `backend/` 下对应的 system 模块是**历史代码，已互相漂移，禁止作为模仿对象**。

## 新模块命名约定

`make gen name=asset` 时模块名是单数，但**路由路径与权限码前缀统一用复数**：

| 位置 | 形态 | 示例 |
|---|---|---|
| 模块名 / 前端目录 | 单数 | `asset`、`views/asset/index.vue` |
| 后端路由路径 | 复数 | `/assets` |
| 前端分组 path（自建时） | 复数 | `assets` |
| 前端叶子 path（自建时） | 空串 | `''`（首叶子 path:''，URL 为 `/assets`） |
| 前端叶子 path（注入时） | 复数 | `asset_categories`（URL 为 `/<group>/<复数>`） |
| 权限码前缀 | 复数 | `assets:view` / `assets:create` / `assets:edit` / `assets:delete` |
| Go 标识符 | PascalCase | `Asset`、`GetAssetList` |

复数规则由 `backend/scripts/pluralize.sh` 统一产出（s/x/z/ch/sh → +es，辅音+y → ies，其余 +s）。
它是生成器与 guard 测试的共同单一真相，**不要在别处另写一套规则**。

> **菜单结构**：前端 `router/index.js` 的 `path: '/'` children 是两级（分组 + 叶子），
> 不是一级平铺。新增菜单 MUST 归入分组（`make gen group=` 或按 `menu-grouping` 结构手写），
> 由 ESLint 自定义规则 `eslint-rules/menu-group.js` 强制（`make lint` 报红）。

## 后端（backend/）

| 我要做 | 去这里 | 关键说明 |
|---|---|---|
| 新增一个业务模块 | `make gen name=<模块名>` | 从 `_example/` 生成三层骨架，并自动注入路由、AutoMigrate、权限码、前端页面/API/路由 |
| 理解分层范式 | `backend/_example/{models,services,controllers}/` | 唯一标准答案 |
| 加一条路由 | `backend/router/router.go` | 受保护路由挂 `PermissionRequired`，其权限码 MUST 已在 `initBaseData` 注册（有护栏） |
| 建表/迁移 | `backend/database/database.go` | `AutoMigrate` + `initBaseData`（权限按 code 幂等补齐） |
| 复数化规则 | `backend/scripts/pluralize.sh` | 路由路径与权限码前缀的单一真相 |
| 鉴权中间件 | `backend/middleware/{jwt,permission,logger}.go` | JWT / RBAC / 操作日志 |
| 统一响应 | `backend/utils/response.go` | `Success`/`Fail`/`SuccessPage` 等 |
| 架构护栏 | `backend/internal/guard/guard_test.go` | 分层铁律等约束的静态检查 |

## 前端（frontend/src/）

| 我要做 | 去这里 | 关键说明 |
|---|---|---|
| 新增页面 | `views/<domain>/index.vue` | `make gen` 会自动生成 |
| 加接口定义 | `api/index.js` | 按模块分组追加 |
| 加路由/菜单 | `router/index.js` | 菜单由 `Layout.vue` 从路由声明派生（两级分组），改这一处即可；**禁止**改 `Layout.vue`（有护栏）；**禁止**裸挂顶层（menu-grouping ESLint 规则强制） |
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
make lint    # go vet + 四端 ESLint（含 menu-grouping 菜单结构规则）
make smoke   # 冒烟：启动→登录→命中受保护路由→断言
make gen name=<模块名>                       # 生成新模块骨架（自建分组）
make gen name=<模块名> group=<分组 path>     # 生成新模块骨架（注入已有分组）
```
