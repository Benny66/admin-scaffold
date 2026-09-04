# AGENTS.md — AI 开发规范（宪法）

本文件是 AI 在本仓库中生成、修改代码时必须遵守的**通用硬性约束**。违反以下规则的代码视为不合格。

> **域专属规则就近放置**：后端规则见 [`backend/CLAUDE.md`](backend/CLAUDE.md)，前端规则见 [`frontend/CLAUDE.md`](frontend/CLAUDE.md)。AI 在哪个目录干活，就优先读哪份域宪法，本文件只放跨端通用铁律。

***

## 1. 多端统一

- 状态管理目录 MUST 统一为 `stores/`（禁止 `store/` 单复数混用）。

- 路径别名 `@` 指向 `src/`（前端、移动端与小程序端一致）。

- 字段 JSON tag 统一 snake\_case（如 `real_name`、`created_at`）。

- 小程序端页面目录为 `pages/`（uniapp `pages.json` 硬约定，不强求与前端 `views/` 一致）。

## 2. 依赖登记制

新增依赖 MUST 登记到根 [`deps.yaml`](deps.yaml) 并附一句理由，而非仅 `go get` / `npm install`。登记制由 `internal/guard/` 的依赖护栏测试强制：清单里的直接依赖未登记、或登记项不存在于清单，`make test` 都会失败。

- 判据提示（非硬禁）：优先选择**可跨项目复用的通用依赖**（如 excelize 做导入导出、echarts 做图表、qrcode 做二维码），业务专属逻辑不进基座。

- 具体项目若要引入这些通用库，只需在 `deps.yaml` 登记即可，不存在「禁止引入」清单。

详细说明见 [`docs/依赖管理.md`](docs/依赖管理.md)。

## 3. 干净性

MUST NOT 提交：`*.db`、`*.exe`、`dist/`、`node_modules/`、`*.orig`、`*.rej`、`vite.config.js.timestamp-*.mjs`。这些已列入 `.gitignore`。

## 4. 唯一范例与代码生成

- 新增业务模块 MUST 参考 `backend/_example/`（黄金路径唯一范例），禁止参考 `views/system/` 五件套（它们是历史模块，已互相漂移）。

- 优先使用 `make gen name=<模块名> [group=<分组 path>]` 从范例生成骨架，再填充 `// TODO: 业务逻辑` 锚点。

  - 不传 `group`：自建分组（分组 path=<复数>，首叶子 `path: ''`，URL 为 `/<复数>`）

  - 传 `group=<已存在分组 path>`：把新模块注入该分组的 `children`（URL 为 `/<group>/<复数>`）

- 详细导航见 [`docs/map.md`](docs/map.md)。

> 与 README 的分工：README 讲「基座有哪些能力」，本文件讲「新增代码照着谁写」。
> 「系统管理五件套」是可运行的能力实现（开箱即用），但不是代码范例——范例唯一指向 `backend/_example/`。

## 4.1 前端菜单 MUST 归入分组（menu-grouping）

- `frontend/src/router/index.js` 的 `path: '/'` 路由 `children` MUST 采用两级结构：外层是分组容器（`path` + `meta.title` + `meta.icon`，无 `component`/`name`），内层是叶子页面。

- 现有「系统管理」五件套（用户/角色/权限/字典/日志）已收进 `path: 'system'` 分组；新增业务模块 MUST 归入分组（用 `make gen group=` 或按 `menu-grouping` 结构手工写 router），**禁止裸挂顶层**。

- 分组可达性：分组必须满足「首叶子 `path: ''` 或 分组自身声明 `redirect`」之一，否则直接访问分组 URL 会 404。

- 该约束由 ESLint 自定义规则 `eslint-rules/menu-group.js` 强制（`make lint` 报红），AST 解析不到根路由时也 MUST 报错而非静默放行。

- 详见 `openspec/changes/menu-grouping/`。

## 5. 验证入口

- 改完代码 MUST 跑 `make test`（含架构护栏 guard 测试）与 `make smoke`（冒烟），不得只 `go build`。

- guard 测试把分层铁律/响应协议/模型完整性编译成了会红的静态检查，违反即构建失败。

