## Context

当前字典模块为单栏形态：`frontend/src/views/system/dict/index.vue` 以「字典类型表格 + 点击行/按钮弹窗显示字典项 + 字典项增删改弹窗（append-to-body 叠层）」实现。数据关系为 `DictType` (id, name, code, description, status, sort) 1—N `DictItem` (dict_type_id, label, value, sort, status, remark)，`code` 唯一。后端已有分页接口与 `/dict/code/:code` 公开查询。项目此前无任何导入导出设施（`deps.yaml` 无 excelize，backend 无 CSV/Excel 工具）。`contracts/openapi.yaml` 目前只登记了 `GET /dict/types` 一条极简 dict 路径。

## Goals / Non-Goals

**Goals:**
- 将字典管理改为左右分栏：左侧字典类型列表，右侧当前类型字典项列表；无弹窗即切换。
- 字典项支持 Excel（xlsx）导入导出，含单类型与全量两层次。
- 全程保持既有数据模型、`/system/dict` 路由、`dict:view/create/edit/delete` 权限码不变。

**Non-Goals:**
- 不实现字典类型的导入导出（本次只做字典项）。
- 不引入前端 xlsx 解析依赖；导入文件解析统一在后端完成。
- 不改变字典项/字典类型的数据结构，不做数据库迁移。
- 不做导入的并发/异步任务（同步处理，字典项规模小）。

## Decisions

**D1. 左栏用分页列表（10 条/页），右栏选中类型用 `currentTypeId` 记录**
- 左栏复用现有 `GET /dict/types`（page_size=10）与关键字搜索，分页状态独立。
- 右栏加载依据是 `currentTypeId`（点击左栏行时写入），与左栏页码解耦——翻页或搜索不重置已选类型。
- 备选：全量加载左栏。被否——用户明确要分页，且分页与后端现有接口完全一致、无新接口。

**D2. 导入导出文件格式统一为 Excel xlsx，后端新增 excelize 依赖**
- 单类型导出、全量导出、导入模板、导入解析共用 `github.com/xuri/excelize/v2`。
- 列头固定为：`label | value | sort | status | remark`。`status` 约定 `1/0`（与模型一致），`sort` 缺省 `0`、`remark` 可空。
- 备选 CSV（零依赖）：被否——用户体验与编辑友好度差，且 AGENTS.md 明示 excelize 为鼓励的通用可复用依赖。excelize 需登记 `deps.yaml` 并附理由。

**D3. 三个后端接口，全部复用既有权限码，不新增**
- `GET /dict/types/:id/items/export` → 单类型导出（`dict:view`）
- `GET /dict/types/export` → 全量导出，多 sheet，sheet 名用类型中文名或 code（`dict:view`）
- `POST /dict/types/:id/items/import` → 上传 xlsx，按 value 覆盖合并（`dict:create`）
- 理由：导出是读操作归 `dict:view`，导入写字典项归 `dict:create`；不引入 `dict:export`/`dict:import` 新码，避免扩 guard 白名单与菜单/后端权限集合。guard「前端 ⊆ 后端」保持通过。
- 全量导出的 sheet 命名冲突（同名类型）用 code 兜底。

**D4. 导入合并语义 = 按 value 覆盖（upsert）**
- 逐行读取：`value` 已存在于当前 `dict_type_id` 下 → 更新 label/sort/status/remark；不存在 → 新增。
- 文件未涉及的既有项保留（非全量替换）。返回统计：新增 N 条、更新 N 条、跳过/失败详情。

**D5. 前端交互**
- 左栏：`el-table` 行点击选中（highlight-current-row）+ 行内 `hover` 显示 ✎/🗑（借助 `el-popover`/自定义 CSS 控制显隐），顶部「新增字典类型」，底部左栏分页。
- 右栏：工具条 `[新增][导入][导出][导出全部]`，字典项表格 + 分页；`[导入]` 用 `el-upload`（`auto-upload=false` + `http-request` 自定义，携带 `dict_type_id`）；`[导出]`/`[导出全部]` 用 blob 下载。
- 「导出全部」放左栏或顶部工具栏（全局性操作），不随右栏选中类型变化。

**D6. 模板下载**：`GET /dict/types/:id/items/export` 在导出数据的同时，也用于「下载导入模板」（前端另给一个「下载模板」按钮，或导入弹窗内提供）。表头与导入约定一致，避免导入列名填错。

## Risks / Trade-offs

- [多 sheet 全量导出时 sheet 名可能重复/含非法字符] → 优先用 `type.name`，冲突或非法时回退 `type.code`，仍冲突则追加序号；sheet 名在 excelize 里有长度/字符限制，需 sanitize。
- [导入覆盖语义可能覆盖用户手改的 label] → 导入后返回「新增/更新」明细，前端提示用户确认结果；导入接口按 value 匹配是明确约定，写入前可在 controller 做行数上限校验（如 >500 拒绝）。
- [大字典项导入阻塞请求] → 字典项规模小（单类型一般 < 100 条），同步处理可接受；不做异步任务，复杂度不划算。
- [excelize 增加后端依赖体积] → 通用可复用依赖，AGENTS.md 明示鼓励；登记 deps.yaml 即可，无 guard 冲突。
- [左侧分页 + 右侧选中解耦导致「当前类型不在可视页」] → 高亮仍持在右栏标题显示类型名，左栏翻页可重新定位；属于可接受的小取舍。

## Migration Plan

- 纯增量，无数据迁移。前后端改动可一次性合入，不改表结构。
- 回滚：还原 `dict/index.vue` 与三个接口即可，无残留数据副作用（导入产生的字典项如需撤销，走现有删除字典项功能）。

## Open Questions

- 是否需要「下载导入模板」作为独立按钮（vs 仅靠导出文件当模板）——默认提供，实现为前端按钮复用导出接口，生成空表头文件。
- 全量导出的 sheet 标题用中文 type.name 还是 code——默认 type.name，冲突回退 code。