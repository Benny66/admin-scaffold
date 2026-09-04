## Why

当前字典管理是「字典类型列表 + 点击弹窗查看字典项」的单栏形态：切换字典类型必须关闭弹窗再点下一个，编辑字典项时弹窗叠弹窗遮挡上下文，且字典项不支持导入导出，只能一条条手填。对把字典项作为配置数据高频维护的场景，这种交互与批量能力都不够顺。

## What Changes

- **交互重构为左右分栏**：左侧为字典类型列表，右侧为当前选中类型的字典项列表；点左侧行即切换右侧，无需弹窗。

  - 左栏：分页（10 条/页）+ 关键字搜索 + 行内 hover 展示编辑/删除图标 +「新增字典类型」。

  - 右栏：当前类型的字典项分页列表 + 新增/编辑/删除 + 导入/导出工具条。

  - 右侧选中类型由 `currentTypeId` 记录，独立于左栏分页状态。

- **字典项导入导出（Excel xlsx）**，覆盖两个层次：

  - 单类型：导出当前选中类型、导入文件到当前选中类型。

  - 全量：一次导出全部字典类型，Excel 每类型一个 sheet。

- **导入合并语义**：按 `value` 覆盖合并——已存在的 value 更新 label/sort/status/remark，不存在的 value 新增；文件未涉及的项保留。

- **新增依赖**：后端注册 `xuri/excelize`（Excel 读写）。

- **不改变数据模型、路由寻址、权限码**：仅重构前端页面形态并新增 3 个字典项 import/export 接口。`/system/dict`、`dict:view/create/edit/delete` 均保持不变。

## Capabilities

### New Capabilities

- `dict-management-split`: 字典管理页面改为左右分栏交互，并为字典项新增 Excel 导入导出（单类型/全量两层次）、按 value 覆盖合并的导入语义。

### Modified Capabilities

<!-- 无改动既有 spec；dict 模块此前无独立 spec -->

## Impact

- **后端**：`backend/controllers/dict.go`（+导入/导出控制器）、`backend/services/dict_service.go`（+导出数据组装、导入按 value upsert）、`backend/router/router.go`（+3 条路由）、`backend/models`（不变）。

- **前端**：`frontend/src/views/system/dict/index.vue`（整体重构为左右分栏）、`frontend/src/api/index.js`（+3 个导入导出 API 定义）。

- **依赖**：`deps.yaml` 后端登记 `github.com/xuri/excelize/v2`（理由：字典项导入导出，通用可复用依赖）。

- **System of record**：`contracts/openapi.yaml` 若已有 dict 契约需同步补 3 个接口（若存在）。

- 无数据库迁移、无 `.db` 文件变更。

