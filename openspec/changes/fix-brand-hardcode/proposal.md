# Proposal: fix-brand-hardcode

## Why

脚手架一边在 `frontend/CLAUDE.md` 立下「系统名称 MUST 来自 appStore.systemName，禁止硬编码『企业管理系统』」的规矩，一边自己的 `frontend/index.html:7` 与 `mobile/index.html:6` 的 `<title>` 就是硬编码的「企业管理系统」——自相矛盾。更关键的是 `scripts/init.sh` 号称「一键初始化」，却只做 ASCII 替换（`base-backend` → 项目名），完全不碰中文品牌，导致外部使用者跑完 `make init name=my-system` 后，浏览器标签页仍显示「企业管理系统」，需手动全局搜中文去改。这戳破了「clone 后改名即可」的承诺。

## What Changes

- `frontend/index.html`、`mobile/index.html` 的 `<title>` 改为中性占位（如 `Base Admin`），不再硬编码「企业管理系统」。
- `scripts/init.sh` 新增 `--app-name` 参数，把「企业管理系统」与中性占位一并替换为新项目的系统名称。
- 前端/移动端 `fetchSystemInfo` 成功后动态更新 `document.title`（已有 `router.beforeEach` 设 title 的前端逻辑，需确保品牌名来自后端而非 index.html 残留）。
- 消解 `CLAUDE.md` 与 index.html 的自相矛盾：index.html 的 title 作为「编译期占位」，运行时由 brand-config 覆盖。

## Capabilities

### New Capabilities

- `brand-hardcode-cleanup`: 消除硬编码品牌残留，使 `init.sh` 的品牌改名闭环覆盖浏览器标题，占位与运行时品牌解耦。

### Modified Capabilities

（无。）

## Impact

- 修改文件：`frontend/index.html`、`mobile/index.html`、`scripts/init.sh`、`frontend/src/stores/app.js`、`mobile/src/stores/app.js`（可能）、`frontend/src/router/index.js`（可能）、`frontend/CLAUDE.md`。
- 无新第三方依赖。
- 无破坏：`--app-name` 是可选参数，缺省时行为与现状一致。
