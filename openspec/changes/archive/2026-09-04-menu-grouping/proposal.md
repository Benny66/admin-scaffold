## Why

侧边栏菜单目前是一级平铺：用户管理、角色管理、权限管理、字典管理、操作日志五个系统项与 `make gen` 生成的业务模块（如 `asset` → `/assets`）挂在同一层，没有任何归属关系。新增业务模块越多，侧边栏越接近一条无组织的长列表——「新模块该归到哪个功能域」在产品上说不清，视觉上也看不出来。基座作为脚手架，需要为后续每个新项目叠加业务模块时给出可生长的菜单结构。

## What Changes

**菜单从一级平铺改为「分组 + 叶子」两级结构，URL 与权限码保持不变：**

- `frontend/src/router/index.js`：`path: '/'` 的 `children` 改为两层——外层是分组容器（有 `meta.title` + `meta.icon`，无 `component`，纯容器不可点），内层是叶子页面。现有五个系统菜单收进「系统管理」分组（分组 path `system` + 子 path `user`…，URL 仍为 `/system/user`）。
- `frontend/src/layout/Layout.vue`：菜单渲染从单层 `el-menu-item` 改为 `el-sub-menu` + `el-menu-item`。过滤顺序改为**先按权限过滤叶子、再丢弃空分组**——分组下所有叶子都无权限时整组不渲染，不留空壳。
- 新增 ESLint 自定义规则（AST 层，作用于 `frontend/src/router/index.js`），把「菜单必须归组」编译成红灯：叶子禁止裸挂顶层、分组必须有 `title` + `icon`、叶子必须有 `title` + `icon` + `permission`、分组不得为空、分组必须能直达（首个叶子 `path: ''` 或分组带 `redirect`）。
- `backend/scripts/gen-module.sh`：新增 `group=<分组 path>` 参数。指定已存在的分组时把新模块注入该分组；未指定（或分组不存在）时新建一个以模块复数名为 path 的分组（首个子项 `path: ''`，URL 与现状一致仍为 `/assets`）。
- 确认现有 guard（`frontend-rbac` 的权限码闭环、Layout 不得硬编码 `/system/`）在新结构下继续生效，必要时同步微调。

## Capabilities

### New Capabilities
- `menu-grouping`: 前端菜单两级分组能力——路由嵌套声明、Layout 分组渲染与空分组隐藏、gen 的 `group=` 注入、ESLint 结构护栏（裸叶子/缺字段/空分组一律红灯）。

### Modified Capabilities
- `frontend-rbac`: 菜单渲染从一级平铺改为「先过滤叶子再丢弃空分组」的两级派生；菜单项的可用性规则从「平铺可见」扩展为「分组 + 叶子两级可见」。
- `ai-scaffold-codegen`: `make gen` 从「注入平铺路由」改为「支持 `group=` 参数，注入已有分组或新建分组」。

## Impact

- `frontend/src/router/index.js`（嵌套 children，URL/权限码/图标不变）
- `frontend/src/layout/Layout.vue`（`menus` computed 递归化 + 模板 `el-sub-menu`）
- `eslint.config.js` 或新增局部规则文件（自定义规则，AST 层）
- `backend/scripts/gen-module.sh`（`group=` 参数解析 + 两路注入逻辑）
- `backend/internal/guard/frontend_rbac_test.go`（确认/微调，护栏不得误伤嵌套结构）
- 文档：`AGENTS.md` / `frontend/CLAUDE.md` / `README.md`（如有）补「新增菜单必须归入分组」与 gen 用法
- 无后端行为变更、无新依赖、无数据库变更；URL 与权限码全量向后兼容
