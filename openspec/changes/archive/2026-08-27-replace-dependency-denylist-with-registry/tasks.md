# Tasks: replace-dependency-denylist-with-registry

## 1. 依赖清单与登记文件

- [x] 1.1 新增根 `deps.yaml`，初始内容与当前三端实际直接依赖一一对应：后端 `go.mod` 主 require 块 8 项、`frontend/package.json` dependencies 7 项、`mobile/package.json` dependencies 5 项，每项附一句话理由
- [x] 1.2 用脚本/手工核对 deps.yaml 与 go.mod / 两个 package.json 的依赖清单完全一致（无遗漏、无多余）

## 2. guard 测试

- [x] 2.1 新增 `backend/internal/guard/deps_test.go`：解析 deps.yaml（用既有 `gopkg.in/yaml.v3`）+ go.mod 主 require 块 + 两个 package.json 的 dependencies
- [x] 2.2 正向校验：清单中每个直接依赖都出现在 deps.yaml，缺失即 `t.Errorf`（「依赖 X 未登记」）
- [x] 2.3 反向校验：deps.yaml 每个登记项都真实存在于对应清单，多余即 `t.Errorf`（「僵尸条目」）
- [x] 2.4 只校验直接依赖：跳过 go.mod `// indirect` 块与 package.json `devDependencies`
- [x] 2.5 确认不新增第三方依赖：guard 只用标准库 + 既有 yaml.v3

## 3. 宪法改写与文档

- [x] 3.1 改写根 `AGENTS.md` 第 2 条：删除「禁止引入的依赖」硬禁单，替换为「依赖登记制」段，指向 deps.yaml，保留「通用优先」作为登记理由提示而非硬判据
- [x] 3.2 新增 `docs/依赖管理.md`：讲清「加依赖 = deps.yaml 登记一行 + 理由，而非偷偷 import」，并说明登记制与 guard 的关系

## 4. 验证

- [x] 4.1 `make test`（含既有 5 个 guard + 新增 deps guard）通过
- [x] 4.2 反向验证：临时在 go.mod 加一个未登记的依赖，确认 deps guard 变红；移除后恢复绿
- [x] 4.3 确认基座基线 go.mod / package.json 未被改动（本 change 只加 deps.yaml 与 guard，不引入新依赖）
- [x] 4.4 确认根 AGENTS.md 中不再残留「禁止 excelize/echarts/qrcode」等禁单字样
