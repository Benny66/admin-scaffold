# Tasks: frontend-guardrails

## 1. 依赖与配置骨架

- [x] 1.1 在 `frontend/package.json`、`mobile/package.json` 加 devDependencies：`eslint`、`eslint-plugin-vue`（锁定 flat-config 兼容版本）
- [x] 1.2 编写根目录 `eslint.config.js`（flat config），含两条 `no-restricted-imports`（禁 axios、禁 `@/store/`）+ `eslint-plugin-vue` essential 规则
- [x] 1.3 配置 `utils/request.js` 的 `overrides` 豁免，使封装层自身可合法 `import axios`

## 2. 接线

- [x] 2.1 在 `frontend/package.json`、`mobile/package.json` 加 `lint` script，指向根配置
- [x] 2.2 扩展根 `Makefile` 的 `lint` 目标：`go vet` + 前端 ESLint + 移动端 ESLint
- [x] 2.3 在 `.github/workflows/ci.yml` 为 frontend 与 mobile 各加 lint 步骤（或并入现有 build job）

## 3. 验证与收口

- [x] 3.1 安装依赖并首跑 `make lint`，确认基线绿（当前代码已审计干净，不应有违规）
- [x] 3.2 若首跑暴露违规，一并修复（保持 `import axios` 仅 request.js、无 `@/store/` 残留）
- [x] 3.3 确认前端与移动端 `npm run build` 仍通过（lint 引入未破坏构建）
- [x] 3.4 更新 `frontend/CLAUDE.md`，把「禁止直接 import axios」「统一 stores/」从散文标注为「已由 ESLint 强制」
