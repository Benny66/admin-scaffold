# AGENTS.md — AI 开发规范（宪法）

本文件是 AI 在本仓库中生成、修改代码时必须遵守的**通用硬性约束**。违反以下规则的代码视为不合格。

> **域专属规则就近放置**：后端规则见 [`backend/CLAUDE.md`](backend/CLAUDE.md)，前端规则见 [`frontend/CLAUDE.md`](frontend/CLAUDE.md)。AI 在哪个目录干活，就优先读哪份域宪法，本文件只放跨端通用铁律。

---

## 1. 三端统一

- 状态管理目录 MUST 统一为 `stores/`（禁止 `store/` 单复数混用）。
- 路径别名 `@` 指向 `src/`（前端与移动端一致）。
- 字段 JSON tag 统一 snake_case（如 `real_name`、`created_at`）。

## 2. 依赖登记制

新增依赖 MUST 登记到根 [`deps.yaml`](deps.yaml) 并附一句理由，而非仅 `go get` / `npm install`。登记制由 `internal/guard/` 的依赖护栏测试强制：清单里的直接依赖未登记、或登记项不存在于清单，`make test` 都会失败。

- 判据提示（非硬禁）：优先选择**可跨项目复用的通用依赖**（如 excelize 做导入导出、echarts 做图表、qrcode 做二维码），业务专属逻辑不进基座。
- 具体项目若要引入这些通用库，只需在 `deps.yaml` 登记即可，不存在「禁止引入」清单。

详细说明见 [`docs/依赖管理.md`](docs/依赖管理.md)。

## 3. 干净性

MUST NOT 提交：`*.db`、`*.exe`、`dist/`、`node_modules/`、`*.orig`、`*.rej`、`vite.config.js.timestamp-*.mjs`。这些已列入 `.gitignore`。

## 4. 唯一范例与代码生成

- 新增业务模块 MUST 参考 `backend/_example/`（黄金路径唯一范例），禁止参考 `views/system/` 五件套（它们是历史模块，已互相漂移）。
- 优先使用 `make gen name=<模块名>` 从范例生成骨架，再填充 `// TODO: 业务逻辑` 锚点。
- 详细导航见 [`docs/map.md`](docs/map.md)。

> 与 README 的分工：README 讲「基座有哪些能力」，本文件讲「新增代码照着谁写」。
> 「系统管理五件套」是可运行的能力实现（开箱即用），但不是代码范例——范例唯一指向 `backend/_example/`。

## 5. 验证入口

- 改完代码 MUST 跑 `make test`（含架构护栏 guard 测试）与 `make smoke`（冒烟），不得只 `go build`。
- guard 测试把分层铁律/响应协议/模型完整性编译成了会红的静态检查，违反即构建失败。
