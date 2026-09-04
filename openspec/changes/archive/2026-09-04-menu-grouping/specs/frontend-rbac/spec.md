# Spec: frontend-rbac

本 delta 覆盖 `menu-grouping` 对 `frontend-rbac` 的修改：菜单从一级平铺派生成两级（分组 + 叶子），并保证既有 RBAC 语义（权限过滤、超管直通、零可见空态）在新结构下延续。

## MODIFIED Requirements

### Requirement: 菜单必须由路由声明派生并按权限过滤

侧边栏菜单 MUST 从 `router.options.routes` 中 `path === '/'` 记录的 `children` 派生，呈现「分组容器 → 叶子页面」两级。过滤 MUST 分两步：先按每项 `meta.permission` 用 `hasPermission()` 过滤每组叶子，再丢弃叶子数为零的分组（空分组不得以空壳标题残留）。`meta.permission` 缺省的路由 MUST 视为「登录即可见」并保留。管理员（`isAdmin`）MUST 看到全部菜单。分组的层级结构 MUST 由 `menu-grouping` 能力约束（叶子禁止裸挂顶层）。

#### Scenario: 只被授予日志查看权限的用户

- **WHEN** 某用户仅拥有 `logs:view` 权限并登录
- **THEN** 侧边栏渲染「系统管理」分组，组内只有「操作日志」一项，其余叶子不渲染

#### Scenario: 分组下所有叶子均无权限

- **WHEN** 某用户对「系统管理」分组内全部叶子（`users:view`/`roles:view`/`permissions:view`/`dict:view`/`logs:view`）均无权限
- **THEN** 该分组整体不渲染，不留空分组标题

#### Scenario: 管理员不受过滤影响

- **WHEN** 以 `isAdmin` 为 true 的用户登录
- **THEN** 全部分组与其全部叶子渲染（管理员直通，不依赖 `permissions` 数组内容）

#### Scenario: 未声明权限码的公共叶子

- **WHEN** 某叶子路由的 `meta` 未声明 `permission`
- **THEN** 该叶子对所有已登录用户可见

### Requirement: 菜单数据源唯一，禁止硬编码副本

`frontend/src/layout/Layout.vue` MUST NOT 包含菜单路径字面量（如 `'/system/'`）。
添加新菜单项 MUST 只需修改 `frontend/src/router/index.js` 一处（用 `menu-grouping` 的两级结构：先定位归属分组，或新建分组）。

#### Scenario: AI 在 Layout 中新增一份菜单数组

- **WHEN** `Layout.vue` 出现 `'/system/'` 开头的路径字面量
- **THEN** guard 测试失败，提示菜单必须从路由声明派生

#### Scenario: 新增业务模块

- **WHEN** 通过 `make gen name=<模块> group=<分组>` 生成模块并注入路由
- **THEN** 菜单在该分组下自动出现并按权限过滤，无需改动 `Layout.vue`；自建分组时 URL 为 `/assets` 形态（见 `menu-grouping` 与 `ai-scaffold-codegen`）
