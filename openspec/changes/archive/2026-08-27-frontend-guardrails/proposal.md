# Proposal: frontend-guardrails

## Why

上一轮 change「ai-scaffold-maturity」把「约束即代码」做成了**后端会响、前端裸奔**：后端有 5 个 guard 测试让分层铁律/响应协议违反即失败，但前端（`frontend/` 与 `mobile/`）里只有 `vite.config.js`，没有任何 lint/test 工具。`frontend/CLAUDE.md` 里白纸黑字写的「MUST 走 request.js」「MUST 统一 stores/」「禁止直接 import axios」，AI 违反时**不会有任何东西响**——这些仍是散文，不是会红的护栏。本 change 把后端的「约束即代码」哲学对称地补到前端，让前端规则也能在 CI 里失败而非靠 code review 发现。

## What Changes

- 为 `frontend/` 与 `mobile/` 引入 ESLint（含 `eslint-plugin-vue`）作为前端架构护栏。
- 用 ESLint 规则把前端域宪法编译成会失败的检查：
  - 禁止直接 `import axios`（必须走 `utils/request.js`）。
  - 强制状态管理统一从 `@/stores/` 导入，禁止 `@/store/` 单复数回潮。
  - 强制组件使用 `<script setup>` + 三段式结构（可选，作为后续增强）。
- 新增 `make lint` 覆盖前端 ESLint，接入 `.github/workflows/ci.yml`。
- 修复本 change 可能暴露的现有违规（若有）。

## Capabilities

### New Capabilities

- `frontend-guardrails`: 前端/移动端的 ESLint 架构护栏，把「禁止绕过 request.js」「统一 stores/」等域宪法规则编译成会失败的静态检查，并与后端 guard 测试对称。

### Modified Capabilities

（无。本 change 是纯增量基建，不改变现有 spec 级行为。）

## Impact

- 新增依赖（devDependencies）：`eslint`、`eslint-plugin-vue`、`@eslint/js`（或等价 flat config 依赖）。
- 新增配置：`frontend/eslint.config.js`、`mobile/eslint.config.js`（或共享根配置）。
- 修改文件：`frontend/package.json`、`mobile/package.json`（加 lint script）、根 `Makefile`（`lint` 目标覆盖前端）、`.github/workflows/ci.yml`（加前端 lint job）。
- 现有业务代码零破坏：若 lint 首次运行暴露违规，本 change 一并修复。
