# Design: frontend-guardrails

## Context

上一轮 change「ai-scaffold-maturity」完成了后端「约束即代码」：`internal/guard/` 用 5 个 guard 测试把分层铁律、响应协议、模型完整性编译成会红的静态检查。但前端（`frontend/` 与 `mobile/`）仍只有 `vite.config.js`，零 lint/test 工具。

`frontend/CLAUDE.md` 已写下四条硬约束：
1. 请求 MUST 走 `utils/request.js` 的 axios 实例，禁止绕过封装直接 `import axios`。
2. 状态管理目录 MUST 统一 `stores/`（禁止 `store/` 单复数混用）。
3. 组件 MUST `<script setup>` + 三段式结构。
4. API 定义集中 `api/index.js`。

现状：这些是散文，AI 违反时**没有任何东西会响**。本 design 把「约束即代码」对称补到前端。

**基线审计（已确认干净）**：
- `import axios` 仅出现在 `frontend/src/utils/request.js:1` 与 `mobile/src/utils/request.js:1`（各一处，且正是封装层本身）。
- 无 `@/store/` 单数残留（上一轮已统一为 `@/stores/`）。

## Goals / Non-Goals

**Goals:**

1. 引入 ESLint（flat config）+ `eslint-plugin-vue`，作为前端架构护栏。
2. 把「禁止直接 import axios」「统一 stores/」编译成会失败的 ESLint 规则，接入 `make lint` 与 CI。
3. 前端与移动端共享同一套规则（DRY，避免两处配置漂移）。
4. 首跑基线绿（当前代码已干净，不应产生新违规）。

**Non-Goals:**

- 不引入单元测试框架（vitest）——本 change 只做「禁止类」静态护栏，测试是后续独立 change。
- 不做「三段式结构」的强制（`script setup` 结构校验对 ESLint 而言是高复杂度低收益，且易误报），留作后续增强。
- 不改契约真源化（`contracts/openapi.yaml` 的消费是独立 change）。
- 不迁移到 TypeScript。

## Decisions

### D1：工具选型 —— ESLint flat config + eslint-plugin-vue

采用 ESLint 9 的 flat config（`eslint.config.js`）+ `eslint-plugin-vue`。

**Why：** 「禁止直接 import 某模块」「禁止某路径前缀」这类约束，ESLint 是社区标准表达，`no-restricted-imports` 规则原生支持，精度高、可与 IDE 集成。相比 node 脚本 + 正则，ESLint 不会误报注释里的字符串、能正确理解 `import` 语义。

**Alternatives considered：**
- node 脚本 + 正则扫 `import` 语句 → 拒绝，零依赖但易误报/漏报、无 IDE 集成、不可维护。
- Vite 插件 / unplugin transform → 拒绝，那不是"检查"而是"构建期拦截"，且需自定义插件逻辑，复杂度高于 ESLint。
- vitest + 单元测试 → 拒绝，那是"逻辑对不对"，不是"结构合不合规"，本 change 只做后者。

### D2：规则表达 —— 用 `no-restricted-imports` 两条

| 规则 | 实现 | 豁免 |
|---|---|---|
| 禁止直接 `import axios` | `no-restricted-imports` 指定 `axios` | `utils/request.js` 是唯一合法进口（它是封装层本身） |
| 禁止 `@/store/` 单数 | `no-restricted-imports` 的 `patterns` 匹配 `@/store`、`@/store/*` | 无 |

**Why：** 两条规则都精确落在"import 结构"上，`no-restricted-imports` 一条原生规则即可覆盖，无需写自定义 rule。

**关键细节：** 禁 axios 必须豁免 `utils/request.js`——否则把封装层自己也杀了。豁免通过 `overrides` 对 `utils/request.js` 关闭该规则实现。

### D3：配置共享 —— 根目录单一配置 + 两端引用

在根目录放一份 `eslint.config.js`，`frontend/` 与 `mobile/` 通过各自的 `package.json` 的 `lint` script 引用它（或各自放一个指向根配置的薄 wrapper）。

**Why：** 两端规则必须一致，避免"前端禁了、mobile 没禁"的漂移。单一配置是「唯一真相」。

**Alternatives considered：**
- 两端各写一份完整配置 → 拒绝，复制即漂移源，与脚手架哲学相悖。
- npm workspace + 共享依赖提升 → 拒绝，改动太大，超出本 change 范围。

### D4：接入点 —— 扩展 `make lint` + CI

`make lint` 从"仅 `go vet ./...`"扩展为"后端 go vet + 前端 ESLint + 移动端 ESLint"。CI 新增 frontend/mobile 的 lint job（或并入现有 build job）。

**Why：** `make lint` 是脚手架承诺的单一入口（上一轮 D6），新增能力应挂到它下面，而不是让 AI 记新命令。

## Risks / Trade-offs

- [ESLint 9 flat config 生态尚新，`eslint-plugin-vue` 对 flat config 的支持需要特定版本] → 锁定 `eslint-plugin-vue` 的 flat-config 兼容版本，pin 版本号避免漂移。
- [禁 axios 规则误伤 request.js 自身] → 用 `overrides` 明确豁免 `utils/request.js`（D2 已述）。
- [引入 devDependency 与「零依赖」洁癖的张力] → ESLint 是前端最通用的工具，符合 AGENTS.md「可跨项目复用」标准，已在 proposal 阶段与用户确认。
- [现有代码可能有隐藏违规导致首跑不绿] → 已审计：`import axios` 仅 request.js 一处、无 `@/store/` 残留，首跑应绿；若仍有违规，本 change 一并修复。
- [`no-restricted-imports` 对 `@/store` 别名的匹配依赖字面量] → ESLint 按 import 路径字符串字面量匹配，`@/store` 不会因别名解析失败，确认可行。

## Migration Plan

1. 加依赖：`frontend/package.json`、`mobile/package.json` 加 `eslint`、`eslint-plugin-vue`（devDependencies）。
2. 写根 `eslint.config.js`（flat config，含两条 `no-restricted-imports` + `eslint-plugin-vue`）。
3. 两端加 `lint` script，`make lint` 扩展，CI 加前端 lint job。
4. 首跑 `make lint`，确认基线绿（若有违规，一并修复）。

**Rollback：** 纯增量，移除依赖与配置文件即回滚，无业务代码改动。

## Open Questions

- **共享配置的落地形态**：根目录单一 `eslint.config.js`（两端引用）vs 根目录放 `eslint.base.js` 由两端各自 `eslint.config.js` import？倾向前者（更简单），落地时若发现 flat config 跨目录引用有坑，降级为后者。
- **是否纳入 `vue/` 推荐规则**：本 change 聚焦「禁止类」护栏，`eslint-plugin-vue` 的 `plugin:vue/vue3-recommended` 是否整包启用？倾向先只开「essential」级别，避免一次性引入大量风格噪音，后续再收紧。
