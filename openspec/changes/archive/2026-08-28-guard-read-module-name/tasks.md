# Tasks: guard-read-module-name

## 1. guard 去硬编码

- [x] 1.1 在 `backend/internal/guard/guard_test.go` 新增 `readModuleName(t)`，从 `backend/go.mod` 解析 module 名
- [x] 1.2 修改 `Test_ControllerLayerMustNotTouchGORM`：`imp == "base-backend/database"` → `imp == moduleName+"/database"`，错误信息动态拼接模块名
- [x] 1.3 确认 `gorm.io/gorm` 判断保持不变（外部依赖固定路径，无需动态化）

## 2. 验证

- [x] 2.1 `make test`（含 guard）通过，`base-backend` 场景下判断结果与改前一致
- [x] 2.2 反向验证：临时把 go.mod module 改为 `my-system`，确认 controller 直连 `<my-system>/database` 的 import 会被 guard 拦截；恢复后仍绿
- [x] 2.3 确认 guard 未引入新第三方依赖（只用 os/strings 标准库）
