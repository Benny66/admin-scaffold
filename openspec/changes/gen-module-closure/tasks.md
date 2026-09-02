## 1. 基础设施：锚点与复数规则单一真相

- [x] 1.1 新增 `backend/scripts/pluralize.sh`：接受一个英文名词，输出复数形式（规则见 design D1：s/x/z/ch/sh → +es，辅音+y → ies，其余 +s）。该脚本 MUST 作为复数规则的单一真相，由 `gen-module.sh` 与 guard 测试共同调用，避免两处规则漂移。
- [x] 1.2 在 `backend/database/database.go` 的权限声明块前插入 `【gen:permissions】` 锚点注释，作为「生成器注入位置」与「guard 提取范围起点」的双重标记（design D4）。
- [x] 1.3 在 `frontend/src/router/index.js` 的 `children` 数组末尾插入 `【gen:route】` 锚点注释。

## 2. 权限初始化改为可增量补齐

- [x] 2.1 将 `initBaseData` 的权限初始化从「整批 `Count() == 0` 才 `Create`」改为遍历权限声明块、按 `code` 逐个查存后创建（不存在则建，存在则跳过），使新增权限码对已存在的数据库同样生效（design D2）。
- [x] 2.2 新增权限记录的 `Sort` 改为在运行时按当前 `permissions` 表最大值递增计算，MUST NOT 写死在注入的字面量中，避免与存量 `Sort: 1..17` 冲突（design D3）。
- [x] 2.3 将「普通用户」角色的只读权限分配（当前为 `WHERE code LIKE '%:view'` + `Count() == 0` 守卫）同样改为按 code 幂等 upsert，使新模块的 `<资源>:view` 自动进入普通用户权限集。
- [x] 2.4 保持角色、用户、字典三段的初始化语义不变（整批 `Count() == 0`），仅权限段改为增量——避免扩大改动面。（实施注：`role_permissions` 关联段同样改为按 `(role_id, permission_id)` 幂等，否则 `permissions` 表新增的码不会关联到任何角色，权限段改为增量将失去意义。）

## 3. 护栏：把静默失败变成红灯

- [x] 3.1 新增 guard 测试文件 `backend/internal/guard/permission_registry_test.go`：从 `router.go` 提取所有 `PermissionRequired("<code>")` 码，从 `database.go` 的 `【gen:permissions】` 块内提取所有 `Code: "<code>"`，校验前者是后者的子集。
- [x] 3.2 guard 的 `database.go` 提取范围 MUST 限定在 `【gen:permissions】` 锚点起始的权限声明块内，不得把 `Role{Code: "admin"}`、`DictType{Code: "user_status"}` 等其他模型的 `Code:` 字段误当作权限码（design D4）。
- [x] 3.3 guard 在任一侧解析结果为空时 MUST `t.Fatalf` 并提示解析规则可能已变更，不得当作「无引用」静默通过（沿用 `frontend_rbac_test.go` 的 G5c 原则）。
- [x] 3.4 guard 在 `database.go` 缺少 `【gen:permissions】` 锚点时 MUST `Fatal`，提示锚点缺失。
- [x] 3.5 新增 guard 测试：调用 `backend/scripts/pluralize.sh` 计算 `example` 的复数，校验 `_example/router/example.go` 注释中展示的路由路径与之相等——防止范例模板与生成器产出不同步。
- [x] 3.6 运行 `make test`，确认新 guard 在基座基线上首次运行即绿（存量 17 个码已核对一致，不产生误报）。（实施注：负向验证时发现护栏假阴性——被 `//` 注释掉的权限码仍被当作已注册，已修复为提取前先剔除行注释。）

## 4. 生成器改造

- [x] 4.1 `gen-module.sh` 引入 `pluralize.sh` 计算资源复数名，替换现有路由路径与权限码前缀的单数写法：`/asset` → `/assets`，`asset:view` → `assets:view`。（实测 `box → boxes`、`category → categories` 亦正确。）
- [x] 4.2 新增权限码注入：在 `database.go` 的 `【gen:permissions】` 锚点后注入 4 行纯数据字面量（`Name` / `Code` / `Type: "api"` / `Status: 1`），**不含 `Sort`**（design D3）。
- [x] 4.3 新增前端路由注入：在 `router/index.js` 的 `【gen:route】` 锚点前注入路由条目，含 `path`（复数）、`name`（PascalCase）、`component`（懒加载 `views/<name>/index.vue`）、`meta.permission`（`<复数>:view`）。
- [x] 4.4 注入的前端路由条目中，`title` 用模块名占位、`icon` 用中性图标，并附 `// TODO` 注释提示开发者替换为中文标题——中文标题无法从英文模块名推断（design D5）。
- [x] 4.5 修正完成提示第 3 条：去掉「与 `Layout.vue` 的 menus」，改为明确说明「菜单从路由派生，无需也禁止改动 `Layout.vue`」（design D6）。
- [x] 4.6 删除完成提示第 4 条（「如需复数请手动调整 router.go 与 api/index.js」）——复数化与前端路由已由生成器自动处理。
- [x] 4.7 更新完成提示，把权限码注册与前端路由注册从「手工步骤」中移除，使提示只列出真正需要人工介入的事项。

> **实施中修复的既有缺陷**：三处注入块（含既有的 `ROUTE_BLOCK`）都受命令替换 `$(cat <<EOF ...)` 剥离尾部换行的影响，
> 导致注入块末尾与下一行粘连（既有的路由注入已产出 `}\t// 【gen:routes】`，多模块时权限块会首尾粘连）。
> 已在三处 python 注入脚本中统一补回尾部换行（`if not block.endswith("\n"): block += "\n"`）。

## 5. 范例模板同步

- [x] 5.1 更新 `_example/router/example.go` 注释，使示例路径与权限码为复数形式（`/examples`、`examples:view`），与生成器实际产出一致。（原权限码为单数 `example:view`，已改。）
- [x] 5.2 更新 `_example/frontend/api.js` 注释，同步复数路径。（核查后原文件已是 `/examples` 复数，无需改动。）
- [x] 5.3 核对 `_example/` 中其余涉及路径或权限码的文本，确保无遗留的单数示例。（全目录检索后仅剩占位符说明文字，无路径/权限码残留。）

## 6. 文档同步

- [x] 6.1 在 `docs/map.md` 补充新模块命名约定：路由路径与权限码前缀使用资源名复数，由 `scripts/pluralize.sh` 统一产出。（顺带修正同文件里同样过时的「加路由/菜单 → `router/index.js` + `Layout.vue` 两处都要改」条目。）
- [x] 6.2 在 `backend/CLAUDE.md` 第 3 条（RBAC 权限接线）补充说明：新增模块的权限码由 `make gen` 自动注册进 `initBaseData`，手工新增受保护路由时 MUST 同步注册，否则 guard 测试失败。

## 7. 端到端验证

- [x] 7.1 运行 `make test`，确认全部 guard（含新增三项）通过。
- [x] 7.2 **负向验证**：手工从 `initBaseData` 移除一条权限码（如 `logs:view`），确认新 guard 变红并给出可操作的失败信息，随后还原——证明护栏真的有效而非形同虚设。
      **首次负向验证即发现假阴性**：注释掉 `logs:view` 后护栏仍绿，因正则不识别 Go 行注释、被注释的码仍算已注册。已修复（提取前剔除 `//` 之后内容），修复后正确变红。
- [x] 7.3 执行 `make gen name=asset`（另加 `box` 验证复数与多模块注入），确认生成的文件中包含：复数路由 `/assets`、`/boxes`、各 4 条 `assets:*` / `boxes:*` 权限码、前端路由条目，且 `make test` 与 `make build` 通过。
- [x] 7.4 **核心目标验证**：建一个仅 `user` 角色的普通用户，`GET /api/assets` → **HTTP 200**（修复前为 403）；对照组 `POST /api/assets` 仍 **403**，证明权限机制未被绕过。
- [x] 7.5 **幂等验证**：同一库连续启动 3 次，`permissions` 总数稳定在 25，各码均只有 1 条。
- [x] 7.6 **增量补齐验证**：用工作区已有旧库（17 条、无 `assets:*`）启动，自动补齐至 25 条，既有记录未被修改。
- [x] 7.7 运行 `make smoke`（通过）与 lint（frontend / mobile 均通过），并补跑 `npm run build` 确认注入的路由懒加载路径可解析。
- [x] 7.8 清理验证过程中产生的临时模块文件与数据库残留（`asset` / `box` 模块、运行时 db 及 `-wal` / `-shm`），确保工作区干净（对照 `.gitignore`）。
