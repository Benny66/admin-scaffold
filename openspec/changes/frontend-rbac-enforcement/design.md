# Design: frontend-rbac-enforcement

## 上下文

本 change 是「三洞修复」序列的第 2 个。序列为：

```
① 文档定位对齐    （本 change 的 task 5，顺带做掉）
② frontend-rbac-enforcement  ← 本 change
③ auth-token-lifecycle       （动鉴权核心，风险最高，排在②之后）
```

> 注：原序列中的 ④「五件套按 `_example` 重写」经核实已取消——五件套与 `_example`
> 实质对齐、漂移极小，「身份撕裂」用文档降级（task 5）即可解决，重写无实际对象。

② 是序列中最实质的一个：它不只是加权限，而是**顺带确定前端菜单的单一数据源**，
这会成为后续所有前端 change 的地基。

---

## D1 · 菜单数据源：前端 router 单一源，不做后端下发

| 方案 | 代价 | 结论 |
|---|---|---|
| A. 前端 router 单一源 | 新增页面仍需改两处（前端 router + 后端 `router.go`），由 guard 保证不漏 | ✅ 选此 |
| B. 后端下发菜单 | 需新增 `menu` 表 + 菜单管理模块 + `Type: "menu"` 权限类型 | ❌ 膨胀成新模块 |

选 A 的另一个理由：它与现有后端「`PermissionRequired("硬编码码")` 逐条接线」的方式一致，
不引入第二套权限来源。代价（改两处）由 guard G5a 兜住——漏改即构建红灯。

## D2 · 派生方式：`router.options.routes`，不用 `router.getRoutes()`

- `getRoutes()` 返回**扁平化**的所有路由记录，父子关系丢失，需要反推 `path` 前缀才能还原层级。
- `options.routes` 是**原始声明**，保留 `children` 嵌套与声明顺序。

菜单顺序是产品决策（当前是 用户→角色→权限→字典→日志），`getRoutes()` 不保证声明顺序，
`options.routes` 保证。故取后者：找到 `path === '/'` 的记录，遍历其 `children`。

## D3 · 权限码粒度与缺省语义

- **菜单项**用 `:view` 码（`users:view` / `roles:view` / `permissions:view` / `dict:view` / `logs:view`）。
- **按钮**用 `:create` / `:edit` / `:delete`。
- **缺省语义**：`meta.permission` 缺省 = 登录即可见（如将来的首页）。只有声明了 `permission`
  的路由才参与过滤。`/login` 不声明。
- admin 直通逻辑复用 `stores/app.js:61` 已有的 `isAdmin` 判断，不新增第二套。

## D4 · `v-permission` 指令的实现

```js
// 无权限时移除 DOM 节点，而非 v-show
el.parentNode?.removeChild(el)
```

- **为什么不是 `v-show`**：隐藏的按钮仍可被 Tab 键聚焦、仍占布局空间。
- **为什么不是 `disabled`**：保留一个「看得见但点不了」的按钮，用户无法理解为什么点不了。
- 支持 `string` 与 `string[]`，数组语义为**拥有任一即可**（`binding.value` 类型分派）。
- 指令内通过 `useAppStore()` 取权限——指令注册必须在 `app.use(pinia)` 之后（见 tasks 3.2）。

## D5 · 403 路由与死循环陷阱

新增 `views/ErrorPage.vue`，由 `route.meta.code` 决定渲染 403 还是 404
（为将来的 500 留位置，不在本 change 实现）。

新增两条顶层路由：

```
/403                      → ErrorPage（meta.code = 403）
/:pathMatch(.*)*          → ErrorPage（meta.code = 404）
```

**陷阱**：守卫必须在检查权限前放行 `/403` 与 `/login`，否则会无限重定向
（`无权限 → /403 → 守卫检查 /403 的权限 → 无权限 → /403 → ...`）。
白名单判断 MUST 先于权限判断（见 tasks 2.2）。

## D6 · guard G5a 只做单向：FE ⊆ BE

```
BE = backend/router/router.go 中 PermissionRequired("...") 的码集合
FE = frontend/src 下 v-permission="'...'" 与 permission: '...' 的码集合

断言：FE ⊆ BE        前端用到的码，后端必须注册过
不做：BE ⊆ FE        反向会大量误报
```

**为什么不做反向**：`BE` 里有 `users:view` 被 5 个接口共用（`GET /users`、
`GET /users/:id` 等），而这类「接口级码」在前端没有对应的菜单或按钮，反向断言会把正常的
后端接线全部报红。反向校验的收益（发现「后端注册了码但前端没用」）远小于噪声成本。

## D7 · 护栏必须能感知自己瞎了

沿用 `frontend-store-guard` 的 D4 经验（`frontend_store_test.go:61-66`）：解析结果为空时
必须 `Fatal`，而非当作「无引用」静默通过。否则正则哪天失效，护栏会静默变成摆设——
而「摆设型的半截功能」正是本 change 要修的问题本身。

两端分别判定：

- **BE 为空** → Fatal（`router.go` 一定能解析出 17+ 个码，空即解析失效）
- **FE 为空** → Fatal（本 change 落地后前端必然有 5 个菜单码 + 若干按钮码，空即解析失效）

错误文案 MUST 点明「若权限码写法已变更，请同步更新本护栏的解析规则」。

## D8 · 前端过滤不是安全边界

菜单过滤与 `v-permission` **仅为交互体验**，去除用户不该看到的操作入口。
真正的访问控制由后端 `PermissionRequired` 中间件保证——用户可以绕过前端直接调接口，
后端 MUST 仍然 403。

本 change 不移除、不弱化任何后端权限校验。

## D9 · 普通用户角色的种子权限（否则登录后一片空白）

发现一个必须先解决的问题：

```
database.go:181-194   只为 admin 角色分配权限
database.go:100-103   「普通用户」角色被创建，但从未分配任何权限
```

因此本 change 落地后，非管理员用户登录 → `permissions: []` → 菜单全空 → 白屏感。

解法两步：

1. **给「普通用户」角色分配 5 个 `:view` 只读权限**（`database.go` 的 `initBaseData`）。
   这让 RBAC 可开箱演示：admin 看到全部，普通用户看到全部菜单但只有只读按钮。
2. **前端为「零可见菜单」提供空态兜底**（`Layout.vue` 侧边栏渲染 `el-empty` 或提示文案），
   防止将来有人配出零权限角色时再次出现白屏。

## D10 · 移动端不在本 change 范围内

`mobile/src` 只有 `Login` 与 `Home`，无菜单、无 RBAC 需求。
guard 的 FE 集合**只扫 `frontend/src`**，不扫 `mobile/src`。
若将来移动端引入 `v-permission`，需同步扩展 guard 的扫描范围（在 spec 中记录此边界）。

---

## 风险

| 风险 | 缓解 |
|---|---|
| 菜单派生顺序或层级出错 | tasks 4 要求以 admin 与普通用户两种身份分别人工验证菜单 |
| `/403` 守卫死循环 | D5 白名单 + tasks 2.2 显式断言 `/403` 可匿名访问 |
| guard 正则对引号风格敏感（单引号 vs 双引号、模板字符串） | tasks 4.4 要求故意写错一个码验证报红；解析规则需覆盖 `"`、`'`、`` ` `` 三种引号 |
| 五件套按钮挂码遗漏 | guard G5a 只保证「用了的码后端有」，不保证「该挂的都挂了」；由 tasks 3.3 逐页清单覆盖 |
