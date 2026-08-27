# Design: AI 脚手架的成熟形态

## Context

当前 `admin-scaffold` 是「可运行的三端 monorepo 骨架」（Go + Gin + GORM + Vue3 + Element Plus + Vant），自称「给 AI 使用」。现状审计结论：

- **约束层（已部分完成）**：根 `AGENTS.md` + 6 份 `docs/` 作为「宪法」，质量尚可（分层铁律、统一响应协议、命名规范、禁依赖清单）。
- **生成层（缺失）**：五件套是五个手写的、已互相漂移的隐性范例（dict 有子表、log 只读、user 有角色绑定），AI 无法判断哪个是标准答案。
- **验证层（缺失）**：4788 行代码 0 个测试文件；CI 三个 job 全部 `go build` / `npm run build`，从不跑测试；「业务可测试」是一句未落地的承诺。
- **安全假象（关键缺陷）**：`PermissionRequired` / `AdminRequired` 中间件已定义，但 `router.go` 从未调用，所有 `protected` 路由只过 JWT + 操作日志，权限码形同虚设。

### 根本判断

AI 脚手架 ≠ 人类脚手架的加速版。人类靠时间积累心智模型，AI 靠**当场读文件**、无跨会话记忆；人类强在判断力、AI 弱在**会漂移/会幻觉/会谎报完成**；人类的关键资源是时间，AI 的关键资源是**上下文窗口（token）**。

推论：人类脚手架优化「起步」，AI 脚手架必须优化「**猜**」和「**验证**」——让 AI 不需要猜，且猜错/做错时**一定会响**。

> 一句话：**AI 脚手架不是「起步代码库」，而是「上下文最优的知识」+「会大声报错的护栏」。**

### AI 循环的四步与对策

```
        AI 循环              这一步怎么翻车            脚手架的对应构件
   ┌──────────────┐      ┌────────────────────┐      ┌──────────────────────────┐
   │ observe      │─────▶│ 读太多→上下文爆掉     │─────▶│ 分层 CLAUDE.md + 导航地图  │
   │  读文件       │      │ 读太少→幻觉/臆造API  │      │ （就近取规则，不全局加载）   │
   ├──────────────┤      ├────────────────────┤      ├──────────────────────────┤
   │ decide       │─────▶│ 选非惯例做法           │─────▶│ 唯一「黄金路径」模块        │
   │  决定改法     │      │ （5个例子互相漂移）     │      │ + codegen/模板            │
   ├──────────────┤      ├────────────────────┤      ├──────────────────────────┤
   │ act          │─────▶│ 生成一致性差的样板      │─────▶│ 约束写成可执行 lint/测试    │
   │  写代码       │      │ （约束只是散文，不绑定）  │      │ （违反就编译失败/测试红）    │
   ├──────────────┤      ├────────────────────┤      ├──────────────────────────┤
   │ verify       │─────▶│ 谎报完成               │─────▶│ 快速冒烟循环 + 契约检查    │
   │  验证对错     │      │ （没跑就 claim done）   │      │ （一条命令证明"它能跑"）    │
   └──────────────┘      └────────────────────┘      └──────────────────────────┘
```

## Goals / Non-Goals

**Goals:**

1. 把「约束即代码」落地——规则从散文变成会红的测试/lint，AI 违反分层铁律时 CI 直接失败而非等 code review。
2. 建立「黄金路径 + 代码生成」——新增模块从「模仿五件套抄 300 行」变成「跑一条命令，填业务逻辑」。
3. 建立「冒烟验证闭环」——「完成了」从一句声称变成可观察的事实（一条命令证明它能跑）。
4. 建立「机器可读契约」——字段名/响应形状有唯一真相，三端不漂移。
5. 收口一致性——规则就近可读（域宪法）、单一入口（Makefile）、导航地图（map.md）。
6. 修安全假象——RBAC 权限码校验真正接线。

**Non-Goals:**

- 不重写五件套业务逻辑（它们继续可用，只是从「隐性范例」降级为「历史模块」）。
- 不引入新的业务依赖（遵守 AGENTS.md 第 5 条禁依赖清单）。
- 不做性能优化、不做多租户/数据权限（超出「AI 脚手架」本 change 的边界）。
- 不强制接入外部 OpenSpec 治理工作流（那是独立于本 change 的组织决策）。

## Decisions

### D1：规则分三层，就近可读（治 observe 的「读太多/读太少」）

| 层 | 文件 | 放什么 |
|---|---|---|
| 根宪法 | `AGENTS.md` | 通用铁律 + 指向域宪法的指针（约 20% 规则，必须短） |
| 域宪法 | `backend/CLAUDE.md`、`frontend/CLAUDE.md` | 三层分层/响应协议/事务约定（后端）；stores/request.js/组件三段式（前端） |
| 导航地图 | `docs/map.md` | 「哪类代码在哪个文件」——AI 靠它定位，不靠 `find` |

**Why：** AI 在哪个目录干活，就自动就近读到哪份规则。根 AGENTS.md 不再塞全部规则，因为 AI 改后端时不需要前端规范。

**Alternatives considered：**
- 单一根 AGENTS.md 塞满所有规则 → 拒绝，AI 每次任务都要读大量无关内容，浪费 token。
- 全靠 docs/ 散文 → 拒绝，散文无执行力（见 D2）。

### D2：约束即代码——护栏优先（治 act 的「约束不绑定」）

把 AGENTS.md 里「禁止」类规则编译成 **guard 测试**（Go 侧用 AST/源码扫描，落在 `backend/internal/guard/` 或 `backend/*_test.go`）：

- 分层铁律：`services/` 不得 `import "github.com/gin-gonic/gin"`；`controllers/` 不得 `import "gorm.io/gorm"`。
- 响应协议：`controllers/` 不得手写 `c.JSON(...)`，必须走 `utils`。
- 模型完整性：`models/` 里每个模型结构体必须出现在 `database.go` 的 `AutoMigrate` 列表。
- 命名规范：`stores/` 目录存在、`store/` 不存在（前端由 lint 规则兜底）。

**Why：** 这是最高杠杆的一层。散文规则 AI 读了依然可能违反；测试红了 AI 会看到、CI 会拦截，等于「宪法自我执行」。

**Alternatives considered：**
- 只靠 lint（golangci-lint / eslint）→ 部分可行但不充分，架构规则（如"service 不碰 gin"）lint 难以表达，guard 测试更精确。
- 靠 code review 兜底 → 拒绝，无法规模化，且 AI 不会自我 review。

### D3：唯一黄金路径 + 代码生成（治 decide 的「选非惯例」）

- 保留**一个** `backend/_example/` 作为唯一范例模块，注释明确标注「这是模式，其余模块由生成器产出」。
- `backend/scripts/gen-module.sh <name>` 从 `_example` 生成 `models/` + `services/` + `controllers/` + `router` 注册 + 前端 `views/` + `api/`。
- 五件套降级为「历史模块」，不再作为 AI 模仿对象。

**Why：** 五个互相漂移的例子是「选非惯例」的温床；唯一范例 + 生成器让判断负担归零，AI 只填业务逻辑。

**Alternatives considered：**
- 保留五件套作范例 → 拒绝，已漂移（dict 有子表、log 只读），不是标准答案。
- 引入第三方脚手架（如 go-zero / scaffolding 库）→ 拒绝，违反「零业务依赖 + 可控」原则，且引入学习成本。

### D4：冒烟验证闭环（治 verify 的「谎报完成」）

`backend/scripts/smoke.sh`：启动 → 登录拿 token → 带 token 命中一个受保护路由 → 断言 200 → 清理。

**Why：** `go build` 只证明编译过；冒烟证明「能跑、能鉴权、能响应」。AI 每改一次跑一遍，「完成」变成可观察事实。

**Alternatives considered：**
- 只跑单测 → 不充分，单测不覆盖启动链路/中间件接线（本 change 的 BUG 正是接线层）。
- 集成测试框架（testcontainers 等）→ 拒绝，重、引入依赖，冒烟脚本对脚手架场景更轻更快。

### D5：机器可读契约（治 act 的「臆造字段名」）

`contracts/openapi.yaml` 作为字段名/响应形状唯一真相；`gen-module.sh` 与前端 `api/index.js` 从它派生，三端一致。

**Why：** `API协议.md` 是散文，AI 大概率写出 `pageSize` 而非 `page_size`。机器契约让漂移无处发生。

**Alternatives considered：**
- 手写 OpenAPI + 人工同步 → 拒绝，契约价值在"机器可读 + 可校验"，手工同步就失去意义。
- gRPC/protobuf 定义契约 → 拒绝，本项目是 HTTP/JSON，引入 gRPC 过重。

### D6：单一入口 Makefile（收敛）

`make test / lint / smoke / gen / dev`，AI 记一条命令即可。

**Why：** 每多一个需要 AI 记住的零散命令，就多一个它做错的概率。

**Alternatives considered：**
- justfile / taskfile → 均可，选 Makefile 因其零额外依赖、CI 与开发环境通用。

### D7：修安全假象——RBAC 接线

在 `router.go` 为受保护路由按权限码挂 `PermissionRequired("users:view")` 等；`AdminRequired` 用于仅超管接口。

**Why：** 当前「权限已保护」是假象；这是安全缺陷而非风格问题，必须修。

**Alternatives considered：**
- 只挂 JWT 不加权限码 → 拒绝，等于承认 RBAC 模型空转。
- 引入 Casbin → 拒绝，超出本 change 范围（现有 RBAC 已够用，见 Open Questions）。

## Risks / Trade-offs

- [guard 测试扫源码可能误报（如注释里出现 gin 字样）] → 用 AST 解析而非文本 grep，精确匹配 import 语句。
- [代码生成器生成出"万能但不贴业务"的样板，业务逻辑仍要手填] → 生成器只产出骨架 + 明确的 `// TODO: 业务逻辑` 锚点，不承诺零手写。
- [唯一范例 `_example` 自身会腐化，变成新的漂移源] → 把 `_example` 纳入 guard 测试的"黄金路径一致性"校验（其他模块与 `_example` 结构对齐）。
- [迁移期间五件套与新范例并存，AI 仍可能去抄五件套] → `docs/map.md` 明确标注「五件套=历史模块，范例=`_example`」，并在 AGENTS.md 根宪法里显式写死。
- [冒烟脚本依赖环境（端口占用、db 残留）] → 脚本内用随机端口 + 临时 db 路径，退出时清理。

## Migration Plan

1. 先落「验证层」（guard 测试 + smoke + Makefile）——不改业务代码，纯增量，风险最低，先让「响铃」工作。
2. 再修「安全假象」（router 挂权限中间件）——独立、可快速验证、有回归冒烟兜底。
3. 再落「生成层」（`_example` + gen-module + 契约）——依赖 1 的护栏来保证范例一致性。
4. 最后收口「一致性」（域宪法拆分、map.md、改名零残留、store/stores 统一）。

**Rollback：** 每步均纯增量或独立可回滚；步骤 2 若需回滚，去掉 router 的中间件调用即可，无需改数据。

## Open Questions

- **权限码粒度**：D7 是按 `resource:action` 粗粒度挂路由，还是需要更细的数据权限（如"只能看自己部门"）？当前脚手架无部门/数据权限模型，若未来需要，是否引入 Casbin 或扩展 RBAC 表？
- **移动端是否纳入黄金路径**：`mobile/` 是否也建 `_example` + 生成器，还是移动端场景太薄（仅 Login/Home/Mine），暂不纳入？当前设计暂定**不纳入**，先聚焦 backend + frontend。
- **契约优先还是 codegen 优先**：D5 的 openapi 契约是"先有契约后有代码"（contract-first）还是"代码导出契约"？对脚手架，倾向 contract-first（AI 先读契约再写代码），但需确认是否引入 openapi 代码生成工具（如 oapi-codegen）——这与「零依赖」原则有张力。
- **guard 测试的放置位置**：独立 `internal/guard/` 包 vs 散落在各 `*_test.go`。倾向集中一处，便于 AI 理解「这是一套护栏」，待落地时确认。
