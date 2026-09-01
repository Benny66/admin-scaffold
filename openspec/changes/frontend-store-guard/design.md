# Design: frontend-store-guard

## Context

`login-brand-visual` 实现过程中引入了一个被四道关卡放行的缺陷：`frontend/src/stores/app.js`
的 `fetchSystemInfo` 调用了两个未定义的 action。因为该函数外层有静默 `try/catch`，
`TypeError` 被吞掉，表现为「背景图不加载 + favicon 不注入」，且无任何错误提示。

```
[mobile]    定义(47,51) + 调用(68,69)   ✓
[frontend]  仅调用(62,63)，无定义        ✗ ← 同一份逻辑，一端漏了
```

值得注意的是：**两端 store 是同构副本**，移动端写对了、桌面端漏了。这类「两端复制时的漏改」
与 `brand-config-guard` 针对的「四处副本漏改」是同一种病，只是发生在 JS 层，而 JS 层目前没有任何护栏。

`internal/guard/` 已有 G1-G3（品牌配置四处副本 + 后端字段被前端消费）。本 change 补 G4，
方向相反：**不是检查后端字段有没有被消费，而是检查前端引用有没有东西可消费。**

## Goals / Non-Goals

**Goals:**

1. 引用了 store 中不存在的成员（`appStore.<X>`）→ `make test` 变红。
2. store 内部调用了未定义的 action（`this.<X>()`）→ `make test` 变红。
3. 失败信息可直接定位（文件 + 成员 + 可用成员列表）。
4. 前端与移动端同等覆盖（AGENTS.md 第 1 条三端统一）。

**Non-Goals:**

- 不做类型检查（不校验参数类型、返回值）。
- 不做完整的 JS 语义分析（那是 TypeScript / ESLint 的领域）。
- 不覆盖非 `appStore` 命名的 store（当前两端只有 `useAppStore`）。
- 不检查模板中的其它表达式错误，只针对 store 成员引用。

## Decisions

### D1：用 regexp 解析 JS 源码，不引入 JS 解析器

成员集合通过三组正则从 `stores/app.js` 提取：

- **actions**：`^ {4}(?:async )?(\w+)\(` —— 覆盖 `setName(x) {` 与 `async fetchSystemInfo() {`
- **getters**：`^ {4}(\w+): \(state\)` —— getters 区特有的 `(state)` 形参
- **state**：在 `state: () => ({ ... })` 块内取 `^ {4}(\w+):`

引用点通过 `appStore\.(\w+)` 与 `this\.(\w+)\(` 两组正则扫描 `src/` 下所有 `.vue` / `.js`。

**Why：** 引入 JS 解析器（如 goja）意味着一个运行时依赖，仅为一条断言不划算，且违反
deps.yaml 的「通用依赖才进基座」判据。而这三组正则匹配的都是本项目高度稳定的固定写法
（4 空格缩进的 Pinia options store），脆弱性可控——若写法变更，护栏会报红并提示同步更新，
属于「护栏响了」而非「护栏漏了」。

**Alternatives considered：**
- 引入 Go 的 JS 解析器做真正的 AST 分析 → 拒绝，依赖成本远高于收益。
- 改用 Node 脚本 + `make lint` 集成 → 拒绝，会分裂验证入口；仓库既定方针是 guard 测试
  统一由 `make test` 触发（AGENTS.md 第 5 条）。
- 只扫 `.vue` 不扫 `.js` → 拒绝，`this.<X>()` 类调用发生在 store 内部，是本次缺陷的实际形态。

### D2：区分「模板引用」与「store 内部调用」两类引用点

- G4a 扫 `appStore.<X>`（覆盖 `.vue` 与 `.js`）
- G4b 扫 `this.<X>()`（仅 store 自身）

**Why：** 两者失败形态不同。G4a 是拼写错误或成员被删；G4b 是 action 漏定义，
且 G4b 正是本次缺陷的形态（`this.setLoginBg(...)`）。分开断言可让失败信息更精确。

**Alternatives considered：**
- 合并为一条断言 → 拒绝，失败信息会变模糊，无法区分「哪里引用了不存在的成员」。

### D3：豁免 Pinia 内置成员（`$` 前缀）

`appStore.$reset`、`appStore.$patch`、`appStore.$subscribe` 等是 Pinia 注入的内置方法，
不在 store 定义中，匹配到应跳过。

**Why：** 否则使用者一旦合法调用内置 API 就会被护栏误报，护栏会被加白名单绕过，失去意义。

### D4：解析不到任何成员时直接 Fatal 而非静默通过

若从 store 文件解析出的成员集合为空（说明 store 写法已变、正则失效），必须 `t.Fatalf`。

**Why：** 这是护栏自身的失效模式。若静默通过，护栏会变成永远绿的假护栏——
正如 `brand-config-guard` 第一版因 `strings.Trim` bug 导致所有 yamlTag 为空却全绿。
**护栏必须能感知自己瞎了。**

## Risks / Trade-offs

- [正则对 store 写法变化敏感] → 由 D4 兜底：写法一变，成员集合为空或骤减，护栏报红提示更新，
  而不是静默放过。这比「护栏失效但显示绿」安全得多。
- [可能误报动态成员访问，如 `appStore[someVar]`] → 当前两端无此写法；正则只匹配
  `appStore.<标识符>` 字面形式，动态访问不会被扫到，故不会误报，但也无法覆盖（属已知边界）。
- [新增 store 文件不会被自动纳入] → 当前两端各只有一个 `stores/app.js`，由护栏硬编码路径；
  若未来新增 store，需同步扩展扫描列表。属可接受的显式性。

## Open Questions

- 是否要把「store 成员引用」检查扩展到 `api/index.js` 的导出函数（如调用了不存在的 API 方法）？
  那需要另一组解析规则，本次不做，留待评估。
