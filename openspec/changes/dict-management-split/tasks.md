## 1. 依赖登记

- [x] 1.1 在根 `deps.yaml` 后端段登记 `github.com/xuri/excelize/v2`（理由：字典项导入导出，通用可复用依赖）

- [x] 1.2 `go get github.com/xuri/excelize/v2`（在 backend/ 下），确认 guard 依赖护栏（deps\_test.go）通过

## 2. 后端：service 层导入导出数据组装

- [x] 2.1 `dict_service.go` 新增 `GetAllDictTypesWithItems()`：一次取出全量字典类型及其字典项（供全量导出）

- [x] 2.2 `dict_service.go` 新增 `GetDictItemsAll(typeID)`：取单个类型的全部字典项（供单类型导出/模板）

- [x] 2.3 `dict_service.go` 新增 `ImportDictItems(typeID, rows)`：按 value 覆盖合并（value 存在→更新 label/sort/status/remark；不存在→新增），返回新增/更新统计

- [x] 2.4 导入前校验行数上限（如 >500 拒绝）与必填字段（label/value 缺失的行跳过或报错）

## 3. 后端：控制器与路由

- [x] 3.1 新增 `controllers/dict.go` 的 `ExportDictItems`（单类型）：用 excelize 生成 xlsx，`Content-Disposition` 下载，鉴权 `dict:view`

- [x] 3.2 新增 `controllers/dict.go` 的 `ExportAllDictTypes`（全量）：多 sheet 导出全部类型，sheet 名用 type.name，冲突/非法回退 code，鉴权 `dict:view`

- [x] 3.3 新增 `controllers/dict.go` 的 `ImportDictItems`：解析上传 xlsx → 调 service 覆盖合并 → 返回新增/更新统计，鉴权 `dict:create`

- [x] 3.4 `backend/router/router.go` 的 `/dict` 组登记 3 条路由：
  - `GET /types/:id/items/export`（dict:view）

  - `GET /types/export`（dict:view，注意与 `/types/:id` 顺序避免抢占）

  - `POST /types/:id/items/import`（dict:create）

- [x] 3.5 `contracts/openapi.yaml` 补登这 3 条 dict 接口（若契约有对应段则更新）

## 4. 前端：API 定义

- [x] 4.1 `frontend/src/api/index.js` 新增：
  - `exportDictItems(id, params)` → blob GET `/dict/types/:id/items/export`

  - `exportAllDictTypes(params)` → blob GET `/dict/types/export`

  - `importDictItems(id, formData)` → POST `/dict/types/:id/items/import`

- [x] 4.2 确认用 axios `responseType: 'blob'` 处理文件下载

## 5. 前端：`dict/index.vue` 左右分栏重构

- [x] 5.1 页面根布局改为左右两栏（如 `el-row`/`el-col` 或 flex），左栏字典类型、右栏字典项

- [x] 5.2 左栏：分页表格（10 条/页，复用 `getDictTypeList`）+ 关键字搜索 +「新增字典类型」；行点击把 id 写入 `currentTypeId` 并触发右栏加载

- [x] 5.3 左栏行内 hover 编辑/删除图标：hover 显隐编辑/删除，编辑复用字典类型弹窗，删除走确认并级联清除后可选的字典项

- [x] 5.4 右栏：当前类型名标题 + 字典项分页表格 + 「新增」「导入」「导出」「导出全部」工具条 + 行内编辑/删除

- [x] 5.5 移除原「字典项弹窗」层，右栏直接作为常驻面板；保留类型新增/编辑与字典项新增/编辑弹窗（现仅 2 层）

- [x] 5.6 删除字典类型后若其为当前选中类型，清空 `currentTypeId` 并复位右栏为空态

## 6. 前端：导入导出交互

- [x] 6.1 「导出」：拉 `exportDictItems` 的 blob 并触发下载（当前选中类型）

- [x] 6.2 「导出全部」：拉 `exportAllDictTypes` 的 blob 下载（全局操作，放左栏顶部或页面级工具栏）

- [x] 6.3 「导入」：`el-upload`（`auto-upload=false` + 自定义 `http-request`）携带 `dict_type_id`，成功后可弹结果（新增 N/更新 N）并刷新右栏

- [x] 6.4 提供「下载模板」：复用单类型导出接口生成空表头文件，或前端用导入弹窗内的模板下载按钮

## 7. 护栏与验收

- [x] 7.1 确认前端新用权限码仍为 `dict:view` / `dict:create`，`frontend_rbac_test.go`（前端 ⊆ 后端）通过，`permission_registry_test.go` 通过

- [x] 7.2 确认菜单结构未改，`eslint-rules/menu-group.js` 不误报；`make lint` 全绿（含四端 ESLint）

- [x] 7.3 `make test` 全绿（含 guard，含 deps.yaml 对新依赖的校验）

- [x] 7.4 `make smoke` 全绿

- [ ] 7.5 手测：左栏选中切换右栏、左栏翻页不重置右栏、类型 CRUD + hover 操作、字典项 CRUD 均正常（交付给用户）

- [ ] 7.6 手测：单类型导出、全量多 sheet 导出、导入覆盖合并（新增/更新统计）、下载模板均正常；无权限用户导出/导入被 403（交付给用户）

