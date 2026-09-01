# Proposal: frontend-rbac-enforcement

## Why

基座的 RBAC 只做了一半。

后端是完整的：17 个权限码在 `backend/router/router.go` 上逐条接线，`middleware/permission.go`
做「JWT → Preload Roles → JOIN 权限 → 比对」，登录接口还把权限码数组一并返回
（`backend/controllers/auth.go:39`）。

前端断在最后一米：

```
controllers/auth.go:39     permissions: [...]        ← 权限码已经发到浏览器
        │
        ▼
stores/app.js:61           hasPermission(code) 已定义
        │
        ▼
        ✗                  全项目零调用
```

`frontend/src/layout/Layout.vue:104` 的菜单是硬编码的 5 项，不按权限过滤；
`frontend/src/router/index.js:21` 的 `meta` 只有 `title` 和 `icon`，没有权限字段。

于是 clone 基座的人拿到的是**「后端锁得住、前端装作没锁」**。三个用户可见的表现：

1. **菜单对所有人长一样** —— 只被授予 `logs:view` 的用户照样看到全部 5 个菜单，点进去 4 个报 403。
2. **按钮无差别展示** —— 没有 `users:delete` 的用户照样看到「删除」按钮，点了才知道不行。
3. **没有 403 页面** —— 路由守卫不拦截，用户撞上后端 403 只得到一个 toast，停在原地不知所措。

根因不是「忘了调用 `hasPermission`」，而是**没有可挂载的位置**：菜单在 Layout 里硬编码、
路由 meta 没有权限字段，两处都容不下权限信息。所以第一步不是补调用，是**先合并数据源**。

## What Changes

1. **菜单单一数据源**：删除 `Layout.vue:104` 的硬编码 `menus` 数组，改为从 `router.options.routes`
   派生。顺带消除「router meta 一份 + Layout 一份」的双写，使 `make gen` 生成新模块时只需改 router
   一处——与后端 `【gen:routes】` 锚点对称。
2. **`meta` 增加 `permission` 字段**，菜单按 `hasPermission` 过滤；受保护路由在守卫中拦截并跳转 `/403`。
3. **新增 `ErrorPage` 组件**与 `/403`、404 catch-all 兜底路由（当前仓库两者都缺）。
4. **新增 `v-permission` 指令**，为 `views/system/` 五个页面的操作按钮挂载权限码。
5. **新增两条 guard**：权限码前后端一致性（前端用到的码 MUST 在后端注册过）、Layout 不得硬编码菜单。
6. **（顺带）`README.md` 与 `AGENTS.md` 对五件套的定位对齐**——当前 `README.md:26` 称其为
   「最佳实践示范」，而 `AGENTS.md §4` 明令「禁止作为模仿对象」，两份文档对同一批代码给了相反的定位。
   本 change 正在修改这五个页面，顺手消除矛盾。

**非目标（明确排除）**：不做后端下发菜单、不引入 `menu` 表、不做菜单管理模块。前端权限过滤
仅为交互体验，**安全边界仍由后端 `PermissionRequired` 中间件保证**（见 design D8）。

## Capabilities

### New Capabilities

- `frontend-rbac`: 前端权限闭环——菜单与按钮按权限码过滤、无权限路由的 403 兜底，以及把
  「权限码前后端一致」编译成会失败的 guard 测试。

### Modified Capabilities

（无。）

## Impact

- 新增文件：`frontend/src/directives/permission.js`、`frontend/src/views/ErrorPage.vue`、
  `backend/internal/guard/frontend_rbac_test.go`、`openspec/changes/frontend-rbac-enforcement/` 下四个制品。
- 修改文件：`frontend/src/router/index.js`、`frontend/src/layout/Layout.vue`、`frontend/src/main.js`、
  `frontend/src/views/system/{user,role,permission,dict,log}/index.vue`、`README.md`、`AGENTS.md`。
- 无新第三方依赖（`v-permission` 为原生指令，guard 沿用标准库 `regexp`）。
- 无后端业务代码改动：本 change 不改 `router.go` / `middleware/` / `services/`，仅新增 guard 测试文件。
- 行为变化：非管理员登录后菜单与按钮会变少（这是预期内的修正）。当前默认种子数据中普通用户角色
  未分配任何权限，需一并补上只读权限，否则普通用户将看不到任何菜单（见 tasks 1.4）。
- 不并入 `frontend-store-guard`：后者 scope 是「store 成员引用完整性」，与「权限闭环」无关，
  合并会稀释两边的能力定义。
