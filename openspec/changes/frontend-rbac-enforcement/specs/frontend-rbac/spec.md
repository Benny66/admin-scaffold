# Spec: frontend-rbac

把「后端 RBAC 已在 `router.go` 逐条接线、权限码已随登录返回，但前端菜单不过滤、按钮不隐藏、
无权限路由无兜底」这一断链补上，并把「权限码前后端一致」编译成会失败的 guard 测试。

## ADDED Requirements

### Requirement: 菜单必须由路由声明派生并按权限过滤

侧边栏菜单 MUST 从 `router.options.routes` 中 `path === '/'` 记录的 `children` 派生，
并按每项 `meta.permission` 用 `hasPermission()` 过滤。`meta.permission` 缺省的路由 MUST 视为
「登录即可见」并保留。管理员（`isAdmin`）MUST 看到全部菜单。

#### Scenario: 只被授予日志查看权限的用户

- **WHEN** 某用户仅拥有 `logs:view` 权限并登录
- **THEN** 侧边栏只渲染「操作日志」一项，其余四项不渲染

#### Scenario: 管理员不受过滤影响

- **WHEN** 以 `isAdmin` 为 true 的用户登录
- **THEN** 全部 5 个菜单项渲染（管理员直通，不依赖 `permissions` 数组内容）

#### Scenario: 未声明权限码的公共页面

- **WHEN** 某路由的 `meta` 未声明 `permission`
- **THEN** 该路由对所有已登录用户可见

### Requirement: 菜单数据源唯一，禁止硬编码副本

`frontend/src/layout/Layout.vue` MUST NOT 包含菜单路径字面量（如 `'/system/'`）。
添加新菜单项 MUST 只需修改 `frontend/src/router/index.js` 一处。

#### Scenario: AI 在 Layout 中新增一份菜单数组

- **WHEN** `Layout.vue` 出现 `'/system/'` 开头的路径字面量
- **THEN** guard 测试失败，提示菜单必须从路由声明派生

#### Scenario: 新增业务模块

- **WHEN** 通过 `make gen` 生成模块并在 `router/index.js` 注册一条带 `meta.permission` 的路由
- **THEN** 菜单自动出现并按权限过滤，无需改动 `Layout.vue`

### Requirement: 无权限路由必须跳转 403 页面且不得死循环

路由守卫 MUST 在 `to.meta.permission` 存在且 `hasPermission()` 为假时跳转 `/403`。
守卫的**白名单判断 MUST 先于权限判断**，`/login` 与 `/403` MUST 被无条件放行。

#### Scenario: 直接输入无权限页面的 URL

- **WHEN** 无 `users:view` 的用户直接访问 `/system/user`
- **THEN** 跳转 `/403` 并渲染 403 页面

#### Scenario: 403 页面自身不得触发重定向

- **WHEN** 用户被重定向到 `/403`，守卫再次执行
- **THEN** `/403` 命中白名单并被放行，**不得**再次跳转到 `/403`（无死循环）

#### Scenario: 未登录访问受保护路由

- **WHEN** 无 token 的用户访问 `/system/user`
- **THEN** 跳转 `/login`

### Requirement: 未匹配路由必须渲染 404 页面

路由表 MUST 包含 catch-all 兜底路由，渲染 404 页面。

#### Scenario: 访问不存在的路径

- **WHEN** 用户访问 `/no-such-page`
- **THEN** 渲染 404 页面，而非空白内容区

### Requirement: 无权限的操作按钮必须从 DOM 移除

`v-permission` 指令 MUST 在用户无对应权限时将该元素**从 DOM 中移除**
（`el.parentNode?.removeChild(el)`），而非使用 `v-show` 隐藏或 `disabled` 禁用。
指令 MUST 支持 `string` 与 `string[]` 两种取值，数组语义为「拥有任一即可」。

#### Scenario: 无删除权限的用户查看用户列表

- **WHEN** 用户缺少 `users:delete`
- **THEN** 每行的「删除」按钮**不存在于 DOM**（不可被 Tab 聚焦，不占布局空间）

#### Scenario: 多码任一满足

- **WHEN** 元素声明 `v-permission="['users:edit', 'users:create']"` 且用户拥有其中任一
- **THEN** 元素正常渲染

#### Scenario: 一个码都没有

- **WHEN** 用户对数组中所有码均无权限
- **THEN** 元素从 DOM 移除

### Requirement: 超管专属操作必须按 isAdmin 判断

后端以 `AdminRequired()`（而非权限码）保护的操作，前端 MUST 用 `appStore.isAdmin` 判断显隐，
MUST NOT 对其挂 `v-permission`（这类操作在权限表中没有对应码，挂码会被 guard 报红）。

#### Scenario: 清空操作日志按钮

- **WHEN** 渲染「清空操作日志」按钮（后端 `router.go` 以 `AdminRequired()` 保护）
- **THEN** 该按钮以 `v-if="appStore.isAdmin"` 控制显隐，非管理员看不到

### Requirement: 前端使用的权限码必须在后端注册过

前端使用的每个权限码 MUST 能在后端 `router.go` 的 `PermissionRequired("...")` 集合中找到。

扫描范围为 `frontend/src` 下 `.vue` / `.js` 中出现的 `v-permission="..."` 与 `meta.permission`。
本校验 MUST 为单向（前端 ⊆ 后端），MUST NOT 断言反向包含。

#### Scenario: 前端用了后端未注册的码

- **WHEN** 某 `.vue` 出现 `v-permission="'users:export'"`，而 `router.go` 中无 `PermissionRequired("users:export")`
- **THEN** guard 测试失败，指明文件、权限码与「后端缺失」方向

#### Scenario: 后端有码而前端未使用（不得误报）

- **WHEN** `router.go` 中存在 `PermissionRequired("roles:view")`，但前端没有对应的 `v-permission` 用法
- **THEN** guard 测试通过（反向不做断言）

#### Scenario: 移动端不在扫描范围内

- **WHEN** guard 收集前端权限码集合
- **THEN** 只扫描 `frontend/src`，不扫描 `mobile/src`（移动端无菜单与 RBAC 需求）

### Requirement: 前端权限过滤不得成为安全边界

菜单过滤与 `v-permission` MUST 仅被视为交互体验优化。后端 MUST 保留全部既有
`PermissionRequired` / `AdminRequired` 校验，前端过滤 MUST NOT 替代或弱化任何后端校验。

#### Scenario: 绕过前端直接调用接口

- **WHEN** 无 `users:delete` 的用户以有效 token 直接请求 `DELETE /api/users/:id`
- **THEN** 后端仍返回 403

### Requirement: 零可见菜单时不得白屏

当过滤后可见菜单数为 0 时，侧边栏 MUST 渲染空状态提示（如 `el-empty`），MUST NOT 呈现空白。

#### Scenario: 用户未被授予任何权限

- **WHEN** 某用户 `permissions` 为空数组并登录
- **THEN** 侧边栏显示空状态提示文案，而非无内容

### Requirement: 默认角色必须可开箱演示权限差异

「普通用户」角色 MUST 在种子数据中被授予 5 个 `:view` 只读权限，使 clone 后可直接对比
管理员与普通用户的菜单与按钮差异。

#### Scenario: 全新数据库初始化后以普通用户登录

- **WHEN** 数据库首次初始化，创建一个仅绑定「普通用户」角色的账号并登录
- **THEN** 该用户看到全部 5 个菜单，但增/删/改类按钮均不渲染

### Requirement: 护栏解析失效时必须报红而非静默通过

guard 测试 MUST 在权限码解析结果为空时以 `t.Fatalf` 失败，MUST NOT 当作「无引用」静默通过。

判定范围：从 `router.go` 解析出的后端码集合，或从 `frontend/src` 解析出的前端码集合，任一为空即 Fatal。

#### Scenario: 权限码写法变更导致正则失效

- **WHEN** `PermissionRequired` 被重构为变量传参（如 `PermissionRequired(consts.PermUserView)`），
  导致正则解析不到任何码
- **THEN** guard 测试 Fatal，提示「若权限码写法已变更，请同步更新本护栏的解析规则」

### Requirement: 文档对五件套的定位必须自洽

`README.md` 与 `AGENTS.md` 对 `views/system/` 五件套的定位描述 MUST 一致：
五件套是**开箱即用的能力实现**，但 MUST NOT 被描述为代码范例；新增业务模块的代码范例
MUST 唯一指向 `backend/_example/`。

#### Scenario: AI 在两份文档中读到矛盾指引

- **WHEN** 查阅 README 与 AGENTS.md 关于「新增模块参考谁」
- **THEN** 两份文档给出一致答案：`backend/_example/`

### Requirement: 零新依赖且纳入统一验证入口

本能力 MUST NOT 引入任何第三方前端或 Go 依赖（`v-permission` 为 Vue 原生指令，
guard 使用标准库 `regexp`）。新增 guard MUST 由 `make test` 一并执行。

#### Scenario: 依赖清单不变

- **WHEN** 本 change 落地后检查 `deps.yaml`
- **THEN** 无新增登记项

#### Scenario: 纳入统一验证入口

- **WHEN** 执行 `make test`
- **THEN** 新增的 `frontend_rbac_test.go` 随 guard 包一同执行
