# Tasks: frontend-rbac-enforcement

## 1. 菜单单一数据源

- [x] 1.1 `frontend/src/router/index.js`：为 5 个 children 的 `meta` 增加 `permission` 字段
      （`users:view` / `roles:view` / `permissions:view` / `dict:view` / `logs:view`，见 D3）
- [x] 1.2 `frontend/src/layout/Layout.vue`：删除 `:104-110` 的硬编码 `menus` 数组
- [x] 1.3 `Layout.vue`：新增 `menus` computed，从 `router.options.routes` 取 `path === '/'` 记录的
      `children` 派生（D2），并按 `appStore.hasPermission(item.meta.permission)` 过滤；
      缺省 `permission` 的路由视为登录即可见（D3）
- [x] 1.4 `backend/database/database.go` 的 `initBaseData`：为「普通用户」角色（`code: "user"`）
      分配 5 个 `:view` 只读权限，否则普通用户登录后菜单全空（D9）
- [x] 1.5 `Layout.vue`：侧边栏在「零可见菜单」时渲染 `el-empty` 空态兜底，避免白屏（D9）

## 2. 路由守卫与错误页

- [x] 2.1 新增 `frontend/src/views/ErrorPage.vue`，按 `route.meta.code` 渲染 403 / 404
- [x] 2.2 `router/index.js`：新增 `/403`（`meta.code = 403`）与 `/:pathMatch(.*)*`（`meta.code = 404`）
      两条顶层路由；守卫的**白名单判断 MUST 先于权限判断**，否则 `/403` 自身会触发无限重定向（D5）
- [x] 2.3 守卫：`to.meta.permission` 存在且 `hasPermission` 为假时 `next('/403')`
- [x] 2.4 守卫：保留并整理现有行为——已登录访问 `/login` 跳首页；未登录访问受保护路由跳 `/login`

## 3. 按钮级权限

- [x] 3.1 新增 `frontend/src/directives/permission.js`：无权限时 `el.parentNode?.removeChild(el)`
      （不用 `v-show` / `disabled`，理由见 D4）；支持 `string` 与 `string[]`（数组为「拥有任一即可」）
- [x] 3.2 `frontend/src/main.js`：**在 `app.use(pinia)` 之后**注册指令，否则指令内 `useAppStore()` 取不到实例（D4）
- [x] 3.3 为 `views/system/` 五个页面的操作按钮挂载权限码：

      | 页面 | 按钮 → 权限码 |
      |---|---|
      | `user` | 新增→`users:create`；编辑→`users:edit`；删除→`users:delete`；重置密码→`users:edit`；启/禁用→`users:edit` |
      | `role` | 新增→`roles:create`；编辑→`roles:edit`；删除→`roles:delete`；分配权限→`roles:edit` |
      | `permission` | 新增→`permissions:create`；编辑→`permissions:edit`；删除→`permissions:delete` |
      | `dict` | 字典类型与字典项的增/改/删→`dict:create` / `dict:edit` / `dict:delete` |
      | `log` | **例外**：「清空操作日志」后端走 `AdminRequired()` 而非权限码（`router.go:79`），
        前端须用 `v-if="appStore.isAdmin"`，不能挂 `v-permission` |

## 4. guard 护栏

- [x] 4.1 新建 `backend/internal/guard/frontend_rbac_test.go`，沿用 guard 包既有手法（标准库 `regexp`
      + 文件遍历 + `projectRoot()`），不引入 JS 解析器
- [x] 4.2 **G5a 权限码一致性**：断言 `FE ⊆ BE`（D6）。正则必须覆盖 `"`、`'`、`` ` `` 三种引号风格
      - `BE`：扫 `backend/router/router.go` 的 `PermissionRequired("...")`
      - `FE`：扫 `frontend/src` 下 `.vue` / `.js` 的 `v-permission="..."` 与 `permission: '...'`
      - 只做单向，**不做** `BE ⊆ FE`（理由见 D6）
- [x] 4.3 **G5b 菜单不得硬编码**：断言 `frontend/src/layout/Layout.vue` 中不含 `'/system/` 路径字面量
- [x] 4.4 **G5c 护栏失效自检**：`BE` 或 `FE` 任一解析为空即 `t.Fatalf`，文案点明「若权限码写法已变更，
      请同步更新本护栏的解析规则」（D7）
- [x] 4.5 失败信息 MUST 包含文件路径、权限码、缺失方向（前端多出 / 后端缺失），并列出对照集合

## 5. 文档定位对齐（顺带做掉 ①）

- [x] 5.1 `README.md:26`：五件套表述从「最佳实践示范」改为「开箱即用的能力实现，非代码范例；
      新增业务模块请参考 `backend/_example/`」
- [x] 5.2 `AGENTS.md §4`：保留「禁止作为模仿对象」，补一句与 README 的分工说明
      （README 讲「有什么能力」，本文件讲「照着谁写」）
- [x] 5.3 文档降级为**长期定位**，不在文档中预留「回改」TODO（五件套重写经核实已取消：
      五件套与 `_example` 实质对齐、漂移极小，文档降级本身即洞 3 的终态）

## 6. 端到端验证

- [x] 6.1 `make test` 全绿（含新增 guard）
- [x] 6.2 `make lint` 无新增告警
- [x] 6.3 `make smoke` 通过（1.4 改动了种子数据，需确认 admin 登录链路不受影响）
- [ ] 6.4 以 admin 登录：5 个菜单全部可见，所有操作按钮可见
- [ ] 6.5 以「普通用户」登录：5 个菜单可见，增/删/改按钮不渲染（DOM 中不存在，非隐藏）
- [ ] 6.6 直接访问无权限路由的完整 URL → 跳转 `/403`，**无死循环**
- [ ] 6.7 访问不存在的路径 → 渲染 404 页
- [x] 6.8 **后端直连验证安全边界**：以普通用户 token 直接 `DELETE /api/users/:id`，
      仍返回 403（证明 D8——前端过滤未替代后端校验）
- [x] 6.9 验证护栏真会红：临时在 `user/index.vue` 写一个未注册的码（如 `v-permission="'users:export'"`），
      跑 `go test ./internal/guard/` 确认报红并指明文件与码，然后回滚
