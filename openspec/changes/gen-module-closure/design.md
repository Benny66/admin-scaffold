## Context

`make gen` 目前是「后端闭环、前端半手工」：后端 `models`/`services`/`controllers` 全生成，`router.go` 与 `database.go` 通过 `【gen:routes】`/`【gen:migrate】` 锚点自动注入，但**权限码注册、前端路由注册这两步要手工补**，且完成提示本身已经过时。

由此产生的不对等是本次要修的核心：

```
自动注入                     手工补                      后果（漏做时）
────────────────────────    ────────────────────────    ────────────────────────
router.go                    initBaseData               非 admin 403
  PermissionRequired(...)       Permission{Code:...}       且 guard / smoke / CI 全绿
```

既有护栏 `Test_FrontendPermissionCodesRegisteredInBackend` 守的是「前端码 ⊆ `router.go` 的 `PermissionRequired`」，它**不校验 `initBaseData`**——这一环没有任何红灯。

已核对基线：`router.go` 现有 17 个 `PermissionRequired` 码与 `initBaseData` 的 17 条 Permission 记录完全一致，新增护栏在基座首次运行即绿。

相关既有约定：
- `ai-scaffold-guardrails` 已有「模型必须注册进 AutoMigrate」（防漏建表）——本 change 要加的是它的同构对偶。
- `frontend-rbac` 已将菜单改为从路由派生，`Layout.vue` 由 `Test_LayoutMustNotHardcodeMenu` 守护，禁止硬编码菜单。

## Goals / Non-Goals

**Goals:**

- 让 `make gen` 产出的模块**开箱即可被非管理员用户正常访问**（权限码自动注册）。
- 让「漏注册权限码」从静默失败变为 `make test` 红灯。
- 消除前端路由的手工编辑，与后端锚点对称。
- 修正已过时的生成器完成提示，避免误导 AI 去改 `Layout.vue` 撞护栏。
- 统一新模块的 URL 与权限码前缀为复数。

**Non-Goals:**

- 不重构存量五件套的命名不一致（`dict:view` 单数、`/dict/types` 嵌套路径）。
- 不改动 `Permission.Type` 的取值约定（存量全为 `api`，而 permission 页面选项含 `menu`/`button`/`api`）。
- 不做前端 CRUD 抽象封装（已明确推迟）。
- 不把角色/用户/字典的初始化也改成 upsert——仅权限段需要「随模块增长」的能力，其余保持整批 `Count==0` 语义。
- 不引入任何新第三方依赖。

## Decisions

### D1：复数化用内置规则表，不引入依赖

生成器已有 `replace()` 用 `sed` 做占位符替换，为零依赖、跨平台的 bash 脚本。复数化沿用同一约束，用 `case` 实现：

- 以 `s` / `x` / `z` / `ch` / `sh` 结尾 → 加 `es`
- 以「辅音 + `y`」结尾 → `y` → `ies`
- 其余 → 加 `s`

由此：`asset → assets`、`box → boxes`、`category → categories`。

**替代方案**：引入 Python（脚本里已经用 `python3` 做按行注入）。否决——复数化是纯字符串变换，为此启动一次 Python 解释器得不偿失，且 `case` 的规则表对阅读者更直观。

**已知限制**：不处理不规则名词（`person`、`datum`）。模块名由开发者自取，遇到不规则词时开发者可在生成后手工微调，不做过度设计。

### D2：`initBaseData` 权限段改为「按 code 幂等 upsert」

现状是整批守卫，一旦库里已有任意一条权限，新增的码**永远不会被写入**：

```go
DB.Model(&models.Permission{}).Count(&permCount)
if permCount == 0 {
    permissions := []models.Permission{ /* 17 条 */ }
    DB.Create(&permissions)
}
```

改为遍历声明列表、按 `code` 逐个查存：不存在则创建，存在则跳过。`Sort` 在运行时按当前最大值递增计算，而非写死在字面量里——这样生成器注入的新码不会与存量 `Sort: 1..17` 冲突。

这是「老库也能自动补齐」的关键，也是本漏洞的放大器所在：当前实现下，即使是手工加了码的开发者，在已有库上也必须删库重建才能生效。

**Why not 全量改**：角色/用户/字典的初始化不需要「随模块增长」的语义，改它们会扩大风险面且无收益。仅权限段需要。

### D3：生成器只注入「数据字面量」，逻辑集中在 `initBaseData`

生成器往 `database.go` 的 `【gen:permissions】` 锚点注入的，是 4 行纯数据：

```go
{Name: "查看资产", Code: "assets:view", Type: "api", Status: 1},
```

**不含 `Sort`**。`Sort` 由 `initBaseData` 在运行时计算赋值。

**替代方案**：生成器注入时写死 Sort 序号。否决——生成器无法感知目标库里已有的 Sort 分布，写死必然冲突。把「分配 Sort」这个运行时职责留给 `initBaseData`，生成器只负责声明「有哪些码」，职责更干净。

### D4：新 guard 的提取范围必须限定在权限块内

`initBaseData` 里 `Role` 也有 `Code:` 字段：

```go
{Name: "超级管理员", Code: "admin", ...},
```

若用 `Code: "\w+"` 无脑扫描整个文件，会把 `admin` / `user` 也收进「已注册码」集合。本 change 只做正向校验（`router.go ⊆ initBaseData`），因此不会误报——但会污染集合、掩盖真实缺失。

**决策**：以 `【gen:permissions】` 为起点提取，到该字面量块的结束 `}` 为止。锚点同时服务于「生成器注入位置」与「guard 识别范围」，一处定义两处消费，与 `【gen:routes】` 的用法一致。

### D5：前端锚点注入时，`title` / `icon` 生成占位值并留 TODO

路由条目的 `path` / `name` / `component` / `meta.permission` 均可从模块名机械推导，但**中文 `title` 无法从英文模块名推断**，`icon` 也无合理默认值。

**决策**：`title` 用模块名占位、`icon` 用中性图标，并在注入内容旁留 `// TODO` 注释提示开发者替换。宁可生成一个「能用但标题待改」的条目，也不生成一个需要开发者从零手写的空位。

菜单无需任何处理——`Layout.vue` 已从路由派生并按 `meta.permission` 过滤，加了路由条目即自动出现在侧边栏。

### D6：完成提示改为「单一真相」式表述

原第 3 条要求同时改 `router/index.js` 与 `Layout.vue`，而后者已被护栏禁止。修正后明确写出「菜单自动出现，无需改动 `Layout.vue`」，把这条隐性知识变成显性提示——对本仓库尤其重要，因为主要使用者是 AI，`Layout.vue` 的派生机制它无法从代码里直接推断。

### D7：与两个 in-progress change 可并行

`auth-token-lifecycle`（0/43）与 `login-brand-visual`（35/39）改动的文件为 `models/user.go`、`utils/jwt.go`、`middleware/`、`controllers/auth.go`、`router/router.go`、`frontend/{utils/request.js,stores/app.js,views/Login.vue,layout/Layout.vue}` 及 mobile 三文件。

本 change 改动 `scripts/gen-module.sh`、`database/database.go`、`frontend/src/router/index.js`、`_example/` 两个模板、`guard/` 新增文件、`docs/map.md`。

**重叠面**：仅 `frontend/src/router/index.js` 与两者都无交集（两者均未列入该文件）。因此本 change 可独立实施，无需排队。

## Risks / Trade-offs

- **【`initBaseData` 改为 upsert 后，已有库会在重启时写入新权限码】** → 这正是目的。风险在于若某条既有 Permission 的 `code` 被人为改过，会出现重复条目。缓解：upsert 以 `code` 为唯一键做存在性检查，不做更新（只补不覆盖），不会破坏既有数据。
- **【复数化规则对不规则名词失效】** → 生成后开发者手工微调 URL 与权限码即可，生成器不做英语词典。已在 D1 标注为已知限制。
- **【新 guard 依赖 `【gen:permissions】` 锚点存在】** → 锚点缺失时 guard MUST `Fatal`（而非静默通过），沿用 `frontend_rbac_test.go` 中 G5c 的「护栏瞎了必须报警」原则。
- **【生成器行为变更，已生成的单数模块与新模块风格割裂】** → 存量模块不受影响（`require_new` 拒绝覆盖）。若用户希望迁移旧模块，属独立工作，已在 Non-Goals 排除。
- **【前端路由注入的 `title` 是占位值，可能被忘记替换】** → 侧边栏会显示模块名而非中文标题，属可见但非阻断的瑕疵。TODO 注释 + 完成提示第 3 条双重提醒。

## Migration Plan

1. 改 `database/database.go`：新增 `【gen:permissions】` 锚点，权限初始化改为按 code 幂等 upsert，`Sort` 运行时递增。
2. 改 `backend/scripts/gen-module.sh`：增加 `pluralize()`、权限码注入块、前端路由注入块，修正完成提示。
3. 改 `backend/_example/` 两个注释模板：对齐复数约定。
4. 改 `frontend/src/router/index.js`：增加 `【gen:route】` 锚点。
5. 新增 guard 测试文件。
6. 同步 `docs/map.md` 的命名约定说明。
7. 验证：`make test`（新 guard 应全绿）、`make smoke`、手工 `make gen name=asset` 后 `make test` 应仍全绿（证明闭环成立）。

**回滚**：全部改动为文本文件 + 新增测试，`git revert` 即可。数据库侧新增的权限码为纯增量，回滚代码后残留的码不影响功能（`PermissionRequired` 只查已授权的码），无需数据迁移。

**升级已有库**：无需手工步骤，服务启动时 `initBaseData` 自动补齐新增权限码。

## Open Questions

> 以下两项在实施时已按标注的取向落地，均可事后调整。

1. **普通用户角色的 `:view` 权限自动分配** —— **实施取向：保持自动分配。**
   本 change 将其改为幂等 upsert 后，新模块的 `assets:view` 会自动进入「普通用户」权限集，即
   **任何新建模块默认对普通用户只读可见**。选定此取向是为了保持「非管理员登录后菜单不空」的
   开箱演示效果（原代码注释 design D9 亦以此为理由）。
   若你更倾向于「默认不分配、由管理员显式授权」，改动点很集中：删掉 `initBaseData` 末尾的
   `syncRolePermissions(userRole.ID, viewPermissions)` 即可，但需注意非管理员登录后侧边栏会为空。

2. **`Permission.Type` 取值** —— **实施取向：沿用 `api`，不改。**
   存量全为 `api`，生成器注入的 4 条同样为 `api`。但 permission 页面的选项含 `menu`/`button`/`api`，
   新增模块若需要菜单级权限是否应支持 `menu`，留待后续统一，本 change 不涉及。

## 实施中附带修复的既有缺陷

**注入块尾部换行被命令替换剥离。** 三处注入（含 change 之前就存在的 `ROUTE_BLOCK`）都采用
`BLOCK="$(cat <<EOF ... EOF)"` 构造内容，而**命令替换会剥离尾部换行**，导致注入块末尾与下一行粘连：

- 既有的后端路由注入已产出 `}\t// 【gen:routes】` 这种畸形行（不影响 Go 编译，故一直未被发现）；
- 新增的权限块在多模块场景下会出现 `...Status: 1},\t\t// 下一模块的注释` 首尾粘连。

已在三处 python 注入脚本中统一补回尾部换行。这不属于范围蔓延——它是同一生成器的同一缺陷，
且本次新增的两处注入会放大它。
