# Design: replace-dependency-denylist-with-registry

## Context

基座依赖管理现状：根 `AGENTS.md` 第 2 条硬编码「禁止引入」五个包（excelize/echarts/jsBarcode/qrcode/html5-qrcode），判据为「新增依赖 MUST 是可跨项目复用的通用依赖」。该规则无任何 guard 兜底（`internal/guard/` 只锁分层/协议/模型/锚点），且判据与禁单自相矛盾——被禁的通用库其实都满足「跨项目可复用」。

用户场景：copy 基座开发真实项目时，`excelize`/`echarts`/`qrcode` 是刚需。硬禁单从「保基座精简」异化为「卡真实项目」，且因无执行力，实际效果只是「AI 读到后拒绝加依赖」而非「CI 变红」。

## Goals / Non-Goals

**Goals:**

1. 消除「硬禁单挡路」：真实项目可引入业务所需依赖（excelize/echarts/qrcode 等）。
2. 保留「约束即代码」的护栏价值：新增依赖须登记 + 写理由，未登记由 guard 测试拦截。
3. 让规则自洽：判据从「是否通用」改为「是否已登记」，不再出现「通用库反被禁」的矛盾。
4. 零新第三方依赖：guard 复用既有 yaml.v3，沿用 guard「不引入新依赖」约定。

**Non-Goals:**

- 不迁移 asset-admin 的业务代码，不改变 `base-backend` 模块名。
- 不在本次 change 中实际引入 excelize/echarts/qrcode（那是使用方项目的事，基座基线保持不含）。
- 不做「依赖安全审计」（CVE/许可证扫描），登记制只解决「可见性 + 有意决策」，不解决「安全」。

## Decisions

### D1：用 `deps.yaml` 作为依赖单一真相，而非继续用 AGENTS.md 散文

新增项目根 `deps.yaml`，结构：

```yaml
backend:
  - package: gorm.io/gorm
    reason: ORM 基座
  - package: github.com/gin-gonic/gin
    reason: HTTP 框架基座
  # ... 其余既有依赖
frontend:
  - package: element-plus
    reason: UI 组件库
  # ...
mobile:
  - package: vant
    reason: 移动端 UI 组件库
  # ...
```

**Why：** 依赖清单是机器可读的单一真相，可被 guard 测试精确比对，不再依赖散文里「禁止 X」的模糊表述。文件随项目走（copy 基座后即可改），是「项目级」而非「基座级」约束。

**Alternatives considered：**
- 继续用 AGENTS.md 散文 + 改判据 → 拒绝，仍是软约束，无执行力，问题没解决。
- 用 JSON 而非 YAML → 可行，但项目已用 yaml.v3（config.yaml 解析），YAML 与既有技术栈一致、可读性好。

### D2：判据从「是否通用」改为「是否已登记」

新规则核心：「新增依赖 MUST 登记到 `deps.yaml` 并附一句理由」。不再有「禁止引入」清单，只有「未登记的依赖」会被 guard 拦下。

**Why：** 判据「是否通用」无法自动判定、且与禁单矛盾。判据「是否已登记」可被 guard 精确校验，且把决策权交给使用方项目——基座基线默认不含业务库，但项目要加，登记即可。

### D3：guard 测试 `Test_UnregisteredDependency` 双向校验

`backend/internal/guard/deps_test.go` 解析 `deps.yaml` + `go.mod`（require 直接依赖）+ `frontend/package.json` + `mobile/package.json`（dependencies），双向断言：

1. **正向**：go.mod/package.json 里每个**直接依赖**都出现在 deps.yaml 中，否则失败（「新增依赖未登记」）。
2. **反向**：deps.yaml 里登记的每个包都真实存在于对应清单，否则失败（防登记了没装 / typo）。

**Why：** 双向校验与既有 `Test_AllModelsRegisteredInAutoMigrate` 的「正向注册 + 反向防 typo」完全同构，是基座 guard 的既有模式。正向拦截「偷偷 import」，反向拦截「登记了但没实际引入的僵尸条目」。

**Alternatives considered：**
- 只做正向 → 拒绝，会放过登记了没装的 typo。
- 解析 import 语句而非 go.mod/package.json → 更精确（能抓到「import 了但没写进 go.mod」的中间态），但复杂得多，且 go.mod/package.json 本就是依赖的权威清单，够用。

### D4：区分「直接依赖」与「间接依赖」

只校验 `go.mod` 的 `require (...)` 主块（直接依赖），跳过 `// indirect` 块；前端只校验 `dependencies`，跳过 `devDependencies`。

**Why：** indirect 依赖与 devDependencies 是传递依赖/工具链，登记它们会产生海量噪音且无决策价值。登记制的对象是「有意识的、可跨项目复用的**直接**依赖」。这也让 deps.yaml 初始条目可控（go.mod 直接依赖 8 项、frontend 7 项、mobile 5 项）。

### D5：guard 继续「不引入新第三方依赖」

guard 测试用标准库 `os`/`path/filepath`/`strings` + 既有 `gopkg.in/yaml.v3`（`go.mod` 已有）解析 yaml，不新增依赖。

**Why：** 与 `guard_test.go` 开头注释「不引入任何第三方依赖」的既有约定一致（yaml.v3 已是后端既有直接依赖，不算新增）。go.mod/package.json 解析用简单的正则或字符串匹配即可，无需 JSON 解析库（package.json 是 JSON，但提取 dependencies 字段用轻量字符串处理足够，或引入 encoding/json 标准库，二者皆零新依赖）。

### D6：AGENTS.md 第 2 条改写为「依赖登记制」

原文「禁止引入的依赖」段替换为「依赖登记制」段，指向 `deps.yaml`，并删除硬编码禁单。保留「新增依赖 MUST 是可跨项目复用的通用依赖」作为**登记理由的提示**（而非硬判据），因为它的本意（别塞业务专属包）依然有价值，只是不再作为「禁止」依据。

**Why：** AGENTS.md 是 AI 读的宪法，须与 guard 的硬门一致——guard 校验「已登记」，宪法就说「必须登记」，不再说「禁止」。保留「通用优先」作为理由提示，因为它仍是好品味，只是从「硬禁」降级为「建议」。

## Risks / Trade-offs

- [登记制 vs 硬禁单的执行力差异] → 登记制仍有执行力（guard 拦未登记），只是「允许登记」而非「永远禁止」，这恰恰是目标，非风险。
- [yaml 解析在 guard 里新增 import] → yaml.v3 已是后端直接依赖，guard 新增 `gopkg.in/yaml.v3` import 不违反「不新增依赖」约定；若担心，可退回「手写极简 yaml 行解析」规避，但代价是脆弱。倾向直接用 yaml.v3。
- [go.mod 直接依赖 vs 实际 import 的缝隙] → go.mod 是依赖权威，`go mod tidy` 会清理未用依赖，故「go.mod 有但代码没 import」的缝隙会被 tidy 收敛；反向「代码 import 但 go.mod 没写」无法编译，天然暴露。可接受。
- [deps.yaml 与三端清单漂移] → 由 D3 反向校验兜底，任何一端加了依赖忘了登记，`make test` 即红。

## Open Questions

- **devDependencies 是否也要登记**：当前决定不登记（D4），但若未来想把工具链（vite/eslint 版本）也纳入治理，需另立规则。本次范围外。
- **移动端是否纳入 guard 校验**：guard 在 `backend/internal/guard/`，用相对路径 `../../frontend` 访问前端清单，与既有「guard 访问 backend 各层」不同但可行；若觉得跨端解析放后端 guard 别扭，可后续拆出独立校验脚本。本次先放 guard，保持「make test 一处聚合」。
