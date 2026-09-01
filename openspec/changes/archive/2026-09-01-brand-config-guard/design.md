# Design: brand-config-guard

## Context

品牌配置在 `backend/config/config.go` 中有四份副本：

```
① AppConfig struct     (L23-29)    定义字段与 yaml tag
② init() 默认值         (L80-87)    每个字段的默认值
③ yamlFile 影子结构     (L152-159)  仅用于 YAML 反序列化，字段与 ① 重复
④ if 非空覆盖链         (L208-223)  逐字段手写 if 覆盖
```

①②③ 靠人眼同步，④ 靠人肉记忆。这是「配置字段加一半」的温床。

同时有一条已经断掉的链：`GetSystemInfo`（`controllers/system.go:29-35`）返回 5 个字段，`frontend/src/stores/app.js:45` 只解构了 4 个（`subtitle` 落单），`mobile/src/stores/app.js:40` 只解构了 3 个（`subtitle`、`favicon` 落单）。后端加了字段、前端没接，静默变成死配置。

`internal/guard/` 已有同构先例：`deps_test.go` 双向校验 `deps.yaml` 与 go.mod / package.json，`guard_test.go` 用 go/ast 断言分层铁律。本 change 沿用同一手法，不引入新依赖。

## Goals / Non-Goals

**Goals:**

1. 加一个品牌字段时，漏改 ①②③④ 任一处 → `make test` 变红。
2. 后端 `GetSystemInfo` 新增返回字段但两端 store 未消费 → `make test` 变红。
3. 修复当前已存在的 `subtitle` / `favicon` 死配置，让护栏在绿的基线上生效。
4. 消解 `brand-config/spec.md:8` 声称支持环境变量、代码却不支持的漂移。

**Non-Goals:**

- 不重构成单一结构体 + 通用合并函数（见 D3，那是更大的重构）。
- 不新增 `app.*` 的环境变量支持（见 D4）。
- 不改动 `GetSystemInfo` 现有字段的语义与取值逻辑。
- 不校验非品牌段（server / database / jwt）的配置一致性。

## Decisions

### D1：G1/G2 用反射 + go/ast，G3 用 gin.H AST + JS 文本扫描

- **G1**：`reflect.TypeOf(config.AppConfig{})` 遍历字段取 yaml tag；`reflect.TypeOf(config.yamlFile{})` 取 `App` 字段的结构体，同样遍历。两者集合必须相等（双向差集都要报错）。
- **G2**：`go/parser` 解析 `config.go`，在 AST 中搜索 `if yf.App.<GoField> != ""` 形式的 `IfStmt`（`BinaryExpr`，`Op` 为 `!=`，`Y` 为空串字面量）。① 中的每个字段都必须命中。
- **G3**：解析 `controllers/system.go`，定位 `GetSystemInfo` 函数体内的 `gin.H{...}` 复合字面量，取所有 `KeyValueExpr` 的 key；再正则扫描两端 `stores/app.js` 中 `fetchSystemInfo` 里 `const { ... } = res.data` 的解构列表。前者必须是后者的子集。

**Why：** G1/G2 用反射与 AST，与 `guard_test.go` 既有手法同构，无需新增依赖；G3 跨语言（Go→JS），JS 侧无 AST 解析器可用，正则扫描解构行是最小可行手段，且该行是高度稳定的固定写法。

**Alternatives considered：**
- 全部用文本正则 → 拒绝，对结构体改名、tag 顺序变化过于脆弱，与 guard 包既有的 AST 手法不一致。
- G3 引入 JS 解析器（如 goja）→ 拒绝，为一条断言引入一个运行时依赖，违反 deps.yaml 的「通用依赖才进基座」判据。
- G3 只在前端做、移动端豁免 → 拒绝，AGENTS.md 第 1 条要求三端统一，且移动端正是漏得更多的那一端。

### D2：G3 落地前先修复死配置，保证基线是绿的

G3 一旦生效，当前就会红（`subtitle` 未被消费）。故本 change 顺带补上两端 store 对 `subtitle` 的解构与持久化，并为移动端补上 `favicon` 消费（设置 `link[rel="icon"]`，与前端对齐）。

**Why：** 护栏的正确姿态是「锁住已经正确的状态」。带着一个已知红灯落地护栏，会让团队习惯性忽略红灯，护栏就废了。

**Alternatives considered：**
- 给 G3 加一个「已知死配置」豁免名单 → 拒绝，豁免名单会变成新的债务堆积点，正是 `subtitle` 沦为死配置的原因。
- 先落地护栏、把修复推给后续 change → 拒绝，红灯期不可控，且违反「make test 必须绿」的验证入口要求。

### D3：不重构四处副本，只加护栏

保留 ①②③④ 的四份结构，用 guard 测试强制同步。

**Why：** 重构（合并 ①③、或用反射自动遍历覆盖 ④）会改变配置读取的既有行为面，风险与收益不成比例；而护栏用几十行测试就拿到了「漏改即红」的绝大部分收益，且零行为变更。

**Alternatives considered：**
- 合并 ①③ 为单一结构体（给 AppConfig 加 yaml tag，直接反序列化进 `GlobalConfig.App`）→ 收益大但需重写默认值/覆盖语义，属于独立重构，超出本 change 范围。记录为 Open Question。
- 让 ④ 用反射自动遍历覆盖 → 会改变「空串不覆盖」的语义边界，且可读性下降。同属独立重构。

### D4：修正 spec 表述而非补齐环境变量

`openspec/specs/brand-config/spec.md:8` 称品牌四要素支持「默认值 + YAML 覆盖 + 环境变量」三层，但 `config.go:116-136` 的环境变量段只处理 server/database/jwt。选择**修正 spec 表述**为「默认值 + YAML 覆盖」两层。

**Why：** 品牌配置是项目级静态配置，`config.yaml`（gitignore，按部署一份）已是它的天然载体；为 logo 文件名之类的值再开一条环境变量通道收益极低，反而把「4 处」扩张成「5 处」，与 D3 的控制复杂度目标相悖。承认事实比扩张代码便宜。

**Alternatives considered：**
- 补齐 `APP_NAME` / `APP_LOGO` / ... 环境变量支持以符合 spec → 拒绝，理由如上；且容器化场景下挂载 config.yaml 同样可行。
- 把「三层配置范式」的措辞从 spec 整体删掉 → 部分采纳：保留范式描述，但明确其适用范围是 server / database / jwt。

## Risks / Trade-offs

- [G3 的 JS 正则对 `fetchSystemInfo` 写法变化敏感] → 断言扫描的是稳定的 `const { ... } = res.data` 解构行；若未来改为逐字段赋值，测试会失败并提示更新扫描规则，属于「护栏响了」，可接受。
- [G3 会让「后端加字段」变成必须连带两端改动的原子提交] → 这是设计意图（防止死配置复发）。代价是跨端 change 不能拆成两个独立 PR 分先后合入，需在 tasks 中显式配对。
- [G2 的 AST 匹配依赖 `if yf.App.X != ""` 的确切写法] → 若有人改用 `len(...) > 0` 之类写法，测试会红；红是可接受的（提示同步更新护栏），且该写法在 config.go 中已统一。

## Open Questions

- 是否要把 ①③ 合并的独立重构单独立项？（见 D3）本 change 只加锁，不改结构；若后续字段继续膨胀（如主题色、登录标语），重构的收益会显著上升。
