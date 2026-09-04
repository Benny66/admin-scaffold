## Context

侧边栏菜单当前由 `frontend/src/router/index.js` 的 `path === '/'` 记录下**扁平 children** 派生（`frontend-rbac` D1 确立「前端 router 单一源」，`Test_LayoutMustNotHardcodeMenu` guard 强制 Layout.vue 不得出现 `/system/` 字面量），Layout.vue 渲染单层 `el-menu-item`。现状五个系统项与 `make gen` 注入的业务模块（`path: '<plural>'`）同层平铺，无归属。

数据模型侧无需任何改动：本次是纯前端信息架构改造，不动后端、不新增字段、不新增依赖。

## Goals / Non-Goals

**Goals:**
- 菜单呈「分组（容器）→ 叶子（页面）」两级；现有五件套收进「系统管理」分组
- `make gen` 生成的模块有归属：要么进已有分组，要么自建分组
- 「菜单必须归组、分组必须有完整元数据、不允许空分组」被编译成 ESLint 红灯，而非靠 review
- URL、权限码、图标、菜单顺序与改动前**全量向后兼容**（`/system/user`、`/assets` 不变）
- 空分组整组隐藏（叶子被权限过滤光的分组不留空壳）

**Non-Goals:**
- 不做后端下发菜单 / 运行时可配置（延续 `frontend-rbac` D1 的否决；见 design 里被否方案）
- 不做分组可点击直达（分组是纯容器）
- 不做 mobile / miniapp 端改造（它们无侧边栏菜单概念）
- 不做三、四级菜单（两级足够，多于两级引入导航复杂度而无当前收益）

## Decisions

### D1: 菜单数据源继续是前端 router，用嵌套 children 表达分组

沿用 `frontend-rbac` D1（前端 router 单一源）不变，`path === '/'` 的 children 升级为两层：

```js
{
  path: '/', component: Layout, redirect: '/system/user',
  children: [
    {
      path: 'system',                       // 分组容器：无 component
      meta: { title: '系统管理', icon: 'Setting' },
      children: [
        { path: 'user', name: 'SystemUser', component: () => import('@/views/system/user/index.vue'),
          meta: { title: '用户管理', icon: 'User', permission: 'users:view' } },
        // ...role / permission / dict / log
      ],
    },
    // 【gen:route】
  ],
}
```

**理由**：
- 与 `frontend-rbac` 已有护栏体系完全兼容——权限码扫描、Layout 禁止硬编码路径全部沿用
- 不改 URL：Vue Router 拼接父子 path，`system` + `user` = `/system/user`
- 不引入第二真相源（对比「抽独立 menu.js」与「后端下发」，均被否，见下）

**被否方案 A——抽独立 `menu.js` 纯数据，router 与菜单都从它派生**：检查更干净，但页面组件需要 `() => import('@/views/...')` 字面量才能被 Vite 静态分析；纯数据化后要么改用 `import.meta.glob`（改动面大、丢类型提示），要么维护「菜单数据」与「路由声明」两份需互相同步的清单（回到两真相源）。收益（更易检查）由 ESLint AST 规则同等获得，故不值此代价。

**被否方案 B——后端下发菜单树**：`frontend-rbac` D1 已否决（需菜单管理模块 + 第二套权限来源），且后端 `Permission` 表虽有 `Type:"menu"/ParentID/Path/Icon` 字段但从未接线，本次不改动。若未来需要「管理员运行时可配置菜单」，应单开 change 评估，不在此夹带。

### D2: 分组是纯容器，首叶子 path 为空串保证 URL 直达

- 分组节点：只有 `path` + `meta`，**无 `component`**（避免与 Layout 冲突渲染）、**无 `name`**（避免命名路由歧义），UI 用 `el-sub-menu`（点击仅展开/收起，不可导航）。
- 每组**首个叶子 `path: ''`**（Vue Router 默认子路由）：直接访问分组 URL（如 `/assets`）渲染该叶子，避免 404；`make gen` 自建分组时 URL 恰好等于现状的 `/assets`，向后兼容。
- `el-menu` 的 `:default-active="$route.path"` 对子项高亮自动生效；折叠态 `el-sub-menu` 需额外确认弹出行为正常（手测项）。

**理由**：分组即容器、URL 与后端资源复数名一致的组合，使新模块菜单路径天然落在 `/assets`、`/assets/categories` 这种可读且可预言的形态。

### D3: Layout 派生顺序——先滤叶子、再丢空组

`menus` computed 必须按此顺序，反了就会出现「空分组壳」：

```
遍历 '/' 的 children
  └─ 仅保留「有非空 children」的分组
       └─ 叶子按 meta.permission 过滤（缺省可见；isAdmin 直通——沿用现逻辑）
       └─ 过滤后叶子数为 0 → 丢弃该分组
```

模板改为 `el-sub-menu` 外层遍历分组、`el-menu-item` 内层遍历叶子。`menus.length === 0` 的空态 `el-empty` 逻辑不变（现有 `frontend-rbac`「零可见菜单」Requirement 仍覆盖）。

### D4: 结构约束用 ESLint 自定义规则，不用 guard regexp

「菜单必须归组」这类**嵌套结构**约束，用 Go + regexp 解析 `children` 递归会脆弱到不可维护（一眼就撞上 `component: () => import(...)`）。ESLint 走 AST，父子关系是真解析的。

- 位置：新增 `eslint-rules/menu-group.js`（根目录），`eslint.config.js` import 并注册为局部插件规则，`files` 限定 `frontend/src/router/index.js`。
- 规则检查清单（任一违反即 error，报错文案带修法）：
  1. `path: '/'` 的 children 中，叶子（无 `children`）禁止裸挂顶层——除非 `meta.standalone === true`（为将来「首页/工作台」预留，本次无使用方）
  2. 分组节点必须有 `meta.title`（非空字符串）与 `meta.icon`
  3. 分组 `children` 不得为空
  4. 叶子必须有 `meta.title`、`meta.icon`、`meta.permission`
  5. 分组必须可达：存在 `path: ''` 的子项 **或** 分组自身声明了 `redirect`
- 规则解析不到根路由（`path: '/'` 结构写法变化）时必须报错而非静默放行——延续 guard「护栏要能感知自己瞎了」的铁律。

**为何不是 guard 测试**：见上。**为何不放进既有 lint 配置段**：规则体量 > 20 行且含 AST 遍历，放独立文件可单测、可读。

### D5: `make gen` 增加 `group=<path>` 参数，两路注入

调用形态 `make gen name=asset_category group=asset`（`--group` 与位置参数二选一，Makefile 透传）。逻辑：

- **分组不存在（或未传 group）**：在顶层 `【gen:route】` 锚点前插入**完整分组块**——分组 `path` 取模块复数（`assets`）、首叶子 `path: ''`，URL 与现状 `/assets` 一致。
- **分组已存在**：用 python 轻量结构定位——找 `path: '<group>'` 对象，括号配对定位其 `children: [ ... ]` 数组边界，在 `]` 前插入新叶子（子 path 取模块复数，如 `asset_categories`，URL 为 `/assets/asset_categories`；提示文案建议 AI 如需更简洁可改为 `categories`）。
- 前置校验沿用现有「路径/权限码占用检查」，并**新增**「目标分组不存在时报错清单」——不，已存在的才插，不存在的建组，无需新检查；但 `name=asset group=assets` 这类「模块复数撞分组」需提示（分组 `assets` 已有默认叶子 `path:''` 占 `/assets`，新模块应以非空子 path 注入）。
- 完成提示里打印最终 URL 与「把 title 换成中文」TODO，延续 gen「生成骨架 + TODO 锚点」哲学。

**理由**：分组归属是产品决策，故 gen 只提供「显式指定 / 自建」两条路，不静默猜测「往哪个已有组塞」。复杂数组注入用括号配对而非字符串锚点，是因为 router 结构升级后每个分组各自持有子数组，无法再靠单一 `【gen:route】` 注释锚点区分「注入到哪个分组」。

### D6: 现有护栏的兼容性确认

- `Test_LayoutMustNotHardcodeMenu`：改造后 Layout.vue 仍无路径字面量 ✓（嵌套结构只存在于 router/index.js）
- `Test_FrontendPermissionCodesRegisteredInBackend`：`meta.permission` 仍在 `.js` 中、权限码不变 ✓
- `frontendPermissionCodes` 的扫描是 `frontend/src` 下所有 `.vue/.js` 按引号正则抓 `<res>:<act>`——router/index.js 的嵌套 meta 同样命中 ✓
- 需要微调点：若有 guard 断言「children 均为叶子」之类的一级假设需排查（预期没有，实现时全量 `make test` 验证）

## Risks / Trade-offs

- **[Vue Router 对无 component 的父路由有警告？]** → Vue Router 4 允许中间层路由无 component（渲染到上一级 router-view）。实现时手测 `/system/user` 直进与刷新，若遇警告改用分组加 `redirect` 兜底。→ 低风险。
- **[`el-menu` 折叠态 sub-menu 弹出层样式]** → Element Plus 对折叠 sub-menu 用 popup 展示子项，需在折叠态手测。→ 手测项（见 tasks）。
- **[gen 的 python 括号配对被注释里 `[`/`]` 干扰]** → router/index.js 注释是中文散句，风险低；实现用「逐字符 + 字符串/注释状态机」提高鲁棒性，并加「配对不闭合即报错退出」。
- **[ESLint 自定义规则实现细节（flat config 内联插件）]** → 基座 eslint.config.js 已是 flat config；自定义规则以 `plugins: { local: { rules: {...} } }` 内联注册，属标准写法。→ 低风险。
- **[新增模块若「自建单页分组」会堆出一排只有一项的分组]** → 可接受：结构合法、归属清晰，比裸平铺强；文档引导优先把第二个及以后的同域页面收进已有分组。

## Migration Plan

无数据迁移、无后端改动、无 URL 变更，纯前端结构与生成器改造：

1. router/index.js 迁移五件套进 `system` 分组（URL 不变）→ 前端可独立发布
2. Layout.vue 两级渲染 + ESLint 规则落地 → `make lint` 立刻对「裸叶子」报红
3. gen-module.sh `group=` 参数落地
4. 全量 `make test` / `make lint` / `make smoke` / 手测清单

回滚策略：单项均可单独 revert，无级联。唯一持久副作用是 gen 已生成的项目骨架，无影响。

## Open Questions

无阻塞性未决项。可选前瞻：若未来要「运行时菜单配置」，单开 change 评估后端下发（复用 `Permission` 表已预埋的 `Type:"menu"/ParentID` 字段）。
