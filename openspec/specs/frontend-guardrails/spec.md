# frontend-guardrails Specification

## Purpose
TBD - created by archiving change frontend-guardrails. Update Purpose after archive.
## Requirements
### Requirement: 前端 lint 架构护栏存在

项目 MUST 为 `frontend/` 与 `mobile/` 提供 ESLint 配置（flat config + `eslint-plugin-vue`），并通过 `make lint` 一键触发，使前端规则违反时 CI 失败而非靠人工发现。

#### Scenario: AI 违反前端域宪法

- **WHEN** AI 在 `frontend/src/` 或 `mobile/src/` 下写出的代码触犯一条被 ESLint 规则覆盖的约束
- **THEN** `make lint` 返回非零退出码，CI 变红，AI 得到明确的失败信号

#### Scenario: 两端规则一致

- **WHEN** 前端与移动端共享同一份 ESLint 配置（单一真相）
- **THEN** 不存在「前端禁了、移动端没禁」的规则漂移

### Requirement: 禁止绕过请求封装直接导入 axios

前端与移动端 MUST 通过 ESLint 的 `no-restricted-imports` 禁止直接 `import axios`，强制所有请求走 `utils/request.js` 封装，`utils/request.js` 自身是唯一豁免。

#### Scenario: AI 在页面组件直接 import axios

- **WHEN** 任一 `views/`、`api/`、`stores/` 文件出现 `import axios from 'axios'`
- **THEN** ESLint 报错，提示必须使用 `@/utils/request` 的封装实例

#### Scenario: 封装层自身合法导入

- **WHEN** `utils/request.js` 内部 `import axios`
- **THEN** 该导入被豁免，不触发规则（它是封装层本身）

### Requirement: 禁止 store 单复数回潮

前端与移动端 MUST 通过 ESLint 的 `no-restricted-imports` patterns 禁止从 `@/store/`（单数）导入，强制统一使用 `@/stores/`（复数）。

#### Scenario: AI 新建了 store/ 单数目录并导入

- **WHEN** 任一文件出现 `import ... from '@/store/...'`（单数）
- **THEN** ESLint 报错，提示状态管理目录必须统一为 `stores/`

#### Scenario: 合法 stores/ 导入不被误伤

- **WHEN** 文件使用 `import ... from '@/stores/app'`（复数）
- **THEN** 该导入不触发规则

### Requirement: lint 单一入口与 CI 接线

`make lint` MUST 扩展为「后端 go vet + 前端 ESLint + 移动端 ESLint」的聚合命令，CI MUST 包含前端与移动端的 lint 步骤。

#### Scenario: AI 想跑全部静态检查

- **WHEN** 执行 `make lint`
- **THEN** 后端 vet 与前端、移动端 ESLint 依次执行，任一失败即非零退出码

#### Scenario: 合并前自动拦截

- **WHEN** 一个 PR 引入了前端违规
- **THEN** CI 的前端 lint job 失败，阻断合并

