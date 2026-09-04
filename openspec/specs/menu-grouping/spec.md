# menu-grouping Specification

## Purpose

前端侧边栏菜单从一级平铺升级为「分组容器 + 叶子页面」两级的能力：路由嵌套声明、Layout 分组渲染与空分组隐藏、`make gen` 的分组注入，以及把「菜单必须归组」编译成 ESLint 红灯的结构护栏。

## Requirements

### Requirement: 菜单呈两级分组结构且 URL 保持不变

`path === '/'` 路由的 `children` MUST 采用两层结构：外层为分组容器（有 `path` + `meta.title` + `meta.icon`，无 `component`、无 `name`），内层为叶子页面。存量「系统管理」五件套（用户/角色/权限/字典/日志）MUST 收进 `path: 'system'` 的分组，其访问 URL（如 `/system/user`）与权限码（如 `users:view`）MUST 保持不变。

每个分组的首个叶子 MUST 使用空串 path（Vue Router 默认子路由），使直接访问分组路径（如 `/assets`）渲染首个叶子而非 404。

#### Scenario: 访问迁移后的系统管理页面

- **WHEN** 用户访问 `/system/user`

- **THEN** URL、路由名与权限校验结果与改动前完全一致（分组只改变了菜单结构，不改变寻址）

#### Scenario: 直接访问分组路径

- **WHEN** 用户直接输入 `/assets`（`assets` 分组的首叶子 `path: ''`）

- **THEN** 渲染该分组的首叶子页面，而非 404

#### Scenario: 分组容器不带组件与路由名

- **WHEN** 检查 `path === '/'` 下的分组节点声明

- **THEN** 分组节点只有 `path` 与 `meta`，没有 `component`、没有 `name`

### Requirement: 侧边栏按分组渲染叶子并丢弃空分组

Layout.vue 的菜单 MUST 先按 `meta.permission` 过滤每组叶子（缺省可见、`isAdmin` 直通），再丢弃叶子数为零的分组，最后渲染为 `el-sub-menu`（分组，仅展开/收起）与 `el-menu-item`（叶子）两级。

分组 MUST 不可被点击导航（纯容器）。过滤后总可见菜单数为 0 时 MUST 渲染空态提示（沿用 `frontend-rbac` 的「零可见菜单不得白屏」）。

#### Scenario: 只被授予部分叶子

- **WHEN** 某用户仅拥有 `logs:view`

- **THEN** 「系统管理」分组渲染且内部只有「操作日志」一项

#### Scenario: 分组下叶子全无权限

- **WHEN** 某用户对「系统管理」下全部叶子（`users:view`/`roles:view`/…）均无权限

- **THEN** 「系统管理」分组整体不渲染，不留空壳标题

#### Scenario: 管理员看到全部分组与叶子

- **WHEN** 以 `isAdmin` 为 true 的用户登录

- **THEN** 所有分组与其全部叶子渲染，管理员直通不依赖 `permissions` 数组

#### Scenario: 点击分组不导航

- **WHEN** 用户点击分组标题（`el-sub-menu`）

- **THEN** 仅展开/收起子项，不触发路由跳转

#### Scenario: 全部菜单不可见时

- **WHEN** 过滤后可见分组数为 0

- **THEN** 侧边栏渲染空态提示，不呈现空白

### Requirement: 菜单结构由 ESLint 规则强制

项目 MUST 提供作用于 `frontend/src/router/index.js` 的 ESLint 自定义规则（AST 层，非 Go guard），把下列「菜单卫生」要求编译为红灯；规则解析不到根路由时必须报错而非静默放行：

- `path === '/'` 的 children 中，叶子（无 `children` 属性）禁止裸挂顶层，除非 `meta.standalone === true`

- 分组节点必须有非空 `meta.title` 与 `meta.icon`

- 分组的 `children` 不得为空

- 叶子必须有 `meta.title`、`meta.icon`、`meta.permission`

- 分组必须可达：存在 `path: ''` 的子项或分组自身声明了 `redirect`

#### Scenario: AI 把新模块裸挂在顶层

- **WHEN** 某 AI 在 `path === '/'` 的 children 顶层直接加一条无 `children` 的路由

- **THEN** `make lint` 报错，提示「菜单项必须归属于分组」，并指向报错行

#### Scenario: 分组缺图标或标题

- **WHEN** 分组节点声明了 `children` 但 `meta` 缺 `icon` 或 `title` 为空串

- **THEN** `make lint` 报错，指明分组位置与缺失字段

#### Scenario: 空分组

- **WHEN** 分组节点的 `children` 为空数组

- **THEN** `make lint` 报错，提示移除空分组

#### Scenario: 叶子缺权限码

- **WHEN** 叶子页面缺少 `meta.permission`

- **THEN** `make lint` 报错，提示补充权限码

#### Scenario: 规则解析失效时不得静默通过

- **WHEN** `router/index.js` 的路由声明写法变化，导致规则找不到根路由对象

- **THEN** 规则报错提示「路由结构写法可能已变更」，而非当作「无违规」放行

### Requirement: make gen 支持把新模块注入分组

`make gen` MUST 支持 `group=<分组 path>` 参数：

- 指定一个**已存在**的分组时，把新模块作为叶子注入该分组的 `children`

- **未指定**或分组不存在时，新建一个分组（分组 `path` 取模块复数名，首叶子 `path: ''`），使新模块 URL 保持为 `/assets` 形态

- 生成完成后提示中 MUST 输出最终 URL 与「把 `title` 换成中文」的 TODO

#### Scenario: 注入已有分组

- **WHEN** 执行 `make gen name=asset_category group=asset`

- **THEN** 新叶子注入 `path: 'asset'` 分组的 `children`，URL 为 `/asset/asset_category`（子 path 取模块复数名），已有分组其他叶子不受影响

#### Scenario: 未指定 group 时自建分组

- **WHEN** 执行 `make gen name=asset`（不带 `group`）

- **THEN** 生成 `path: 'assets'` 的新分组 + 首叶子 `path: ''`，URL 为 `/assets`，与改动前 `make gen` 的寻址一致

#### Scenario: 自建分组的菜单合法可过 lint

- **WHEN** 执行上述自建分组后运行 `make lint`

- **THEN** 不触发「裸挂顶层/缺字段」等结构违规

#### Scenario: 注入已有分组不破坏结构

- **WHEN** 注入目标分组含多个既有叶子

- **THEN** 生成器在数组边界内插入，既有叶子与注释不被破坏，`make lint` 与 `make test` 均通过

