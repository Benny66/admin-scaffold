## Why

`make gen name=asset` 生成的模块存在一条**从生成到生产全程绿灯、只有真实非管理员用户会撞上**的失败链。

根因是一处不对等：生成器自动往 `router.go` 注入了 `PermissionRequired("asset:view")` 等权限码，但 `initBaseData` 里的 Permission 记录**必须手工补**。漏掉这一步之后：

| 关卡 | 结果 | 原因 |
|---|---|---|
| `make test` | 绿 ❗ | 唯一相关的护栏只校验「前端码 ⊆ router.go 的 `PermissionRequired`」，两边都有就通过，**它不看 `initBaseData`** |
| `make smoke` | 绿 ❗ | 冒烟用 admin 登录，而 `PermissionRequired` 对 `isAdmin` 直接 `c.Next()` 放行，**压根不查权限** |
| CI | 绿 ❗ | 同上，三端 job 全部通过 |
| 真实非管理员用户 | **403** | 中间件在 `role_permissions` 里查不到 `asset:*` |

而且这个失败**无法在界面上自救**：权限管理页的数据源是 `permissions` 表，表里没有 `asset:*`，管理员在「分配权限」里看不到这四个码，想授权都无处下手——只能手工写数据库，或改代码重建库。

这是每个新模块的**必经步骤**，不是偶发疏漏：生成器每产一个模块，就埋一次。

同一条路径上还有三处收尾不完整：

1. **前端路由需手工编辑**——后端有 `【gen:routes】` / `【gen:migrate】` 两个注入锚点，前端 `router/index.js` 没有，不对称。
2. **生成器的完成提示已过时**——第 3 条要求「在 `router/index.js` **与 `Layout.vue` 的 menus** 中新增入口」，但菜单早在 `frontend-rbac` 中改为从路由派生，且有护栏 `Test_LayoutMustNotHardcodeMenu` **禁止**在 `Layout.vue` 出现 `/system/` 字面量。照提示改会直接撞红灯。对本仓库尤其危险——主要使用者是 AI，它会照做。
3. **单复数三处各说各话**——`_example/router/example.go` 注释写 `/examples`（复数），生成器实际产 `/${NAME}`（单数），存量五件套是 `/users` `/roles`（复数）。

## What Changes

**1. 权限码闭环（根治 + 兜底）**

- 生成器自动向 `database.go` 的 `initBaseData` 注入该模块的 4 个权限码（`<资源复数>:view/create/edit/delete`）。
- `initBaseData` 的权限初始化从「整批 `Count==0` 才执行」改为**按 code 逐个幂等 upsert**——使新增的码对**已存在的老库同样生效**（当前实现下老库永远不会补，正是本漏洞的放大器）。
- 「普通用户」角色的只读权限分配（当前为 `WHERE code LIKE '%:view'` + `Count==0` 守卫）同步改为按 code 幂等 upsert，使新模块的 `:view` 码自动进入普通用户权限集，保持「非管理员登录后菜单不空」的开箱演示效果。
- 新增 guard 测试：`router.go` 中所有 `PermissionRequired` 码 **MUST** 能在 `initBaseData` 中找到对应注册。这是现有「模型必须注册进 AutoMigrate」（防漏建表）的同构对偶——防漏注册权限。

**2. 命名统一为复数**

- 生成器产出 `/assets` + `assets:view/create/edit/delete`，与存量五件套及 REST 惯例一致。
- **BREAKING**（仅影响新生成的模块）：产出 URL 与权限码前缀由单数改为复数，需同步 `_example/` 两个注释模板与文档。
- 移除完成提示第 4 条「如需复数请手动调整」——不再需要手工调整。

**3. 前端路由闭环**

- `frontend/src/router/index.js` 增加 `【gen:route】` 锚点，生成器自动注入路由条目（`path` / `name` / `component` / `meta.permission`），与后端锚点对称。
- 菜单无需任何改动——`Layout.vue` 已从路由派生并过滤。

**4. 修正完成提示**

- 第 3 条去掉 `Layout.vue`，明确「只需在 `router/index.js` 注册路由，菜单自动出现」。

**非目标（明确排除）**：不重构存量五件套的命名不一致（`dict:view` 为单数、`/dict/types` 为嵌套路径），不改动 `Permission` 的 `Type` 取值约定（存量全为 `api`，而 permission 页面的选项含 `menu`/`button`/`api`），不做前端 CRUD 抽象封装，不引入新依赖。

## Capabilities

### New Capabilities

（无。）

### Modified Capabilities

- `ai-scaffold-codegen`: 生成器的产出从「后端闭环、前端半手工」扩展为**三端全闭环**——新增 `initBaseData` 权限码注入与前端路由注入，产出命名统一为复数，完成提示与实际行为一致。
- `ai-scaffold-guardrails`: 新增「权限码必须注册进 `initBaseData`」护栏要求，作为既有「模型必须注册进 AutoMigrate」的同构补充，堵住漏注册权限导致的静默 403。

## Impact

- 修改文件：`backend/scripts/gen-module.sh`、`backend/database/database.go`、`frontend/src/router/index.js`、
  `backend/_example/router/example.go`、`backend/_example/frontend/api.js`、
  `backend/internal/guard/`（新增一个测试文件）、`docs/map.md`（同步命名约定）。
- 新增两个注入锚点：`database.go` 的 `【gen:permissions】`、`router/index.js` 的 `【gen:route】`。
- 无新第三方依赖。
- **BREAKING**：新生成的模块 URL 与权限码前缀由单数变复数（`/asset` → `/assets`、`asset:view` → `assets:view`）。存量五件套不受影响，已生成过的模块不受影响（`require_new` 会拒绝覆盖）。
- **基线安全**：已核对 `router.go` 现有 17 个 `PermissionRequired` 码与 `initBaseData` 的 17 条 Permission 记录**完全一致**，新 guard 在基座首次运行即绿，不产生误报。
- 行为变化：`initBaseData` 改为按 code 幂等 upsert 后，老库在升级时会**自动补齐**新增的权限码（当前行为是永不补齐）。
