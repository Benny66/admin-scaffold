## 1. router/index.js：两级分组结构

- [x] 1.1 把现有五个系统菜单（`system/user|role|permission|dict|log`）收进 `path: 'system'` 分组容器（`meta: { title: '系统管理', icon: 'Setting' }`，无 component/name），叶子 path 去掉 `system/` 前缀，URL 与权限码不变
- [x] 1.2 确认 `redirect: '/system/user'` 与 `meta.standalone` 白名单机制不受嵌套影响；`【gen:route】` 锚点保留在顶层 children 末尾
- [x] 1.3 手测：直接访问 `/system/user`、刷新页面、登录后重定向三者正常，无 Vue Router「无 component 父路由」警告

## 2. Layout.vue：两级渲染与空分组隐藏

- [x] 2.1 `menus` computed 改为两级派生：遍历 '/' 的 children → 仅保留有非空 children 的分组 → 组内叶子按 `meta.permission` 过滤（缺省可见、isAdmin 直通）→ 丢弃空分组
- [x] 2.2 模板改为 `el-sub-menu`（分组标题 + icon，仅展开/收起）内嵌 `el-menu-item`（叶子，index 为拼好的完整路径）；`menus.length === 0` 空态沿用 el-empty
- [x] 2.3 手测：普通用户只看到有权限的分组与叶子；分组下叶子全无权限时整组消失；折叠态下 sub-menu 弹出子项正常

## 3. ESLint 结构护栏（menu-grouping 核心）

- [x] 3.1 新建 `eslint-rules/menu-group.js`：AST 规则检查 `path === '/'` 的 children——叶子禁止裸挂顶层（`meta.standalone` 豁免）、分组必有非空 `meta.title`+`meta.icon`、分组 `children` 非空、叶子必有 `meta.title`+`meta.icon`+`meta.permission`、分组可达（`path: ''` 子项或 `redirect`）
- [x] 3.2 规则在解析不到根路由对象时必须报错（护栏能感知自己瞎了），不得静默放行
- [x] 3.3 `eslint.config.js` 注册为局部插件规则（files 限定 `frontend/src/router/index.js`），报错文案带修法
- [x] 3.4 跑 `make lint`：对改造后合法结构零报红；临时把一条叶子拖到顶层验证规则确实报错

## 4. guard 兼容性确认

- [x] 4.1 确认 `frontend_rbac_test.go` 的权限码扫描（嵌套 meta 仍在 `.js` 文本中）与 `Test_LayoutMustNotHardcodeMenu`（Layout 仍无路径字面量）在嵌套结构下继续生效
- [x] 4.2 `make test` 全绿（含 guard）

## 5. gen-module.sh：group= 参数与两路注入

- [x] 5.1 解析 `group=<path>`（Makefile `make gen name=x group=y` 透传），无 group 时置空
- [x] 5.2 自建分组注入：无 group 或分组不存在时，在顶层 `【gen:route】` 锚点前插入完整分组块（分组 path=模块复数、`meta.title` 为占位 TODO、首叶子 `path: ''`），URL 为 `/assets` 形态
- [x] 5.3 注入已有分组：python 轻量结构定位 `path: '<group>'` 的 `children: [...]` 数组边界（字符串/注释状态机 + 括号配对，不闭合即报错退出），在 `]` 前插入叶子（子 path=模块复数，URL 如 `/assets/asset_categories`）
- [x] 5.4 完成提示输出最终菜单 URL、归属分组、「把 title 换成中文」TODO，且不出现要求改 Layout.vue 的措辞
- [x] 5.5 临时目录验证：`make gen name=asset`（自建组）与 `make gen name=asset_category group=asset`（注入组）两种注入后 `make lint` + `make test` 全绿、菜单结构合法

## 6. 文档

- [x] 6.1 `AGENTS.md` / `frontend/CLAUDE.md` 补充：新增菜单 MUST 归入分组（用 `make gen group=` 或按 `menu-grouping` 结构手工写 router），禁止裸挂顶层；gen 用法示例更新
- [x] 6.2 检索并修正现有文档中「菜单是一级/平铺」的描述（如有），与两级结构自洽

## 7. 整体验收

- [x] 7.1 `make test` 全绿（含 guard）
- [x] 7.2 `make lint` 全绿（含新菜单结构规则；backend vet + 四端 ESLint 按现状范围）
- [x] 7.3 `make smoke` 全绿
- [x] 7.4 手测：管理员登录侧边栏为「系统管理 ▼」+ 全部叶子；仅授 `logs:view` 的普通用户只看到操作日志；无权限用户看到空态提示
- [x] 7.5 手测：直接输入 `/assets`（自建分组）渲染首叶子非 404；`el-menu` 折叠态 sub-menu 展开正常
