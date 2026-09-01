# Tasks: brand-config-guard

## 1. 修复既有的死配置（让护栏基线变绿）

- [x] 1.1 `frontend/src/stores/app.js`：`fetchSystemInfo` 解构补 `subtitle`，新增 `setSubtitle` action + localStorage 持久化
- [x] 1.2 `mobile/src/stores/app.js`：补 `subtitle` 与 `favicon` 的解构、持久化，并设置 `link[rel="icon"]`（与前端对齐）
- [x] 1.3 人工验证：`/api/system/info` 返回的 5 个字段现在都被两端 store 接收

## 2. 新增 guard 测试

- [x] 2.1 新建 `backend/internal/guard/brand_config_test.go`，package guard，复用既有 `backendRoot()` / `projectRoot()` 辅助函数
- [x] 2.2 实现 G1：`reflect` 遍历 `config.AppConfig` 与 `config.yamlFile` 的 `App` 字段，断言 yaml tag 集合相等（双向差集都要报错）
- [x] 2.3 实现 G2：`go/parser` 解析 `backend/config/config.go`，断言每个 yaml tag 字段都有对应的 `if yf.App.<F> != ""` 覆盖分支
- [x] 2.4 实现 G3：go/ast 取 `GetSystemInfo` 中 `gin.H{}` 的 key 集合，正则扫描两端 `stores/app.js` 的 `const { ... } = res.data`，断言后端 key 是前端解构的子集
- [x] 2.5 每条断言的失败信息 MUST 指明「漏改了哪一处」（如「字段 login_bg 在 yamlFile 中缺失」），而非只报 fail

> 实现备注：`yamlFile` 是 config 包的非导出类型，guard 包无法用 `reflect` 直接取它的字段，
> 故 G1 改为与 G2/G3 一致的 go/ast 解析（yaml tag 本身仍用 `strconv.Unquote` +
> `reflect.StructTag.Get` 解析）。G1 的断言语义不变。

## 3. 修正 spec 漂移

- [x] 3.1 `openspec/specs/brand-config/spec.md` 第 8 行：把「默认值 + YAML 覆盖 + 环境变量（沿用既有三层配置范式）」改为「默认值 + YAML 覆盖」
- [x] 3.2 在同处补一句：三层配置范式（含环境变量）适用于 server / database / jwt，品牌段为两层

## 4. 验证护栏真的会红

- [x] 4.1 临时在 `AppConfig` 加一个字段但不动 `yamlFile`，跑 `go test ./internal/guard/`，确认 G1 报红并指明缺失位置，然后回滚
- [x] 4.2 同样方式临时验证 G2（加字段但不加覆盖分支）与 G3（后端加返回 key 但前端不解构）
- [x] 4.3 回滚后跑 `make test` 确认全绿，`make lint` 无新增告警
