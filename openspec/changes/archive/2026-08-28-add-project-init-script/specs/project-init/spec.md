# Spec: project-init

一键初始化脚本，把「复制基座 → 改包名/标识 → 重置密钥 → 清历史/残留」固化为可重复命令。

## ADDED Requirements

### Requirement: 一键初始化命令

项目 MUST 提供 `scripts/init.sh <项目名> [--module <go模块名>] [--db-name <名>] [--issuer <名>]`，由 `make init` 透传，一条命令完成基座到新项目的初始化。

#### Scenario: 复制基座后一键改名

- **WHEN** 开发者复制基座为新项目目录，执行 `make init name=my-system`
- **THEN** 脚本将 `go.mod` module 名与所有 `import "base-backend/..."`（含 `_example/` 模板）替换为 `my-system`

#### Scenario: 生成器产出代码仍可编译

- **WHEN** 初始化完成后执行 `make gen name=asset`
- **THEN** 生成的模块 import `<新模块名>/...`，而非已不存在的 `base-backend/...`，编译通过

### Requirement: 运行时标识与密钥重置

初始化 MUST 替换运行时标识（env var 前缀 `BASE_BACKEND_*`、package.sh 二进制名/压缩包名、JWT Issuer），并 MUST 将默认密钥替换为随机串。

#### Scenario: 环境变量前缀随项目名

- **WHEN** 项目名含连字符（如 `my-system`）
- **THEN** env var 前缀规范化为大写去连字符（`MY_SYSTEM_`），如 `MY_SYSTEM_SERVER_PORT`

#### Scenario: 默认密钥被替换

- **WHEN** 初始化执行完成
- **THEN** `config.go` 默认值与 `config.example.yaml` 中的 `base-backend-secret-key-change-me` 被替换为随机生成的密钥

#### Scenario: 数据库名默认不动

- **WHEN** 未传 `--db-name`
- **THEN** 数据库默认名保持 `base_backend.db` / `base_backend` 不变；仅当显式传 `--db-name` 时替换

### Requirement: 清理 OpenSpec 历史与运行时残留

初始化 MUST 删除基座自带的 OpenSpec 历史（`openspec/specs/`、`openspec/changes/`）与运行时残留（`backend/*.db*`、`backend/config.yaml`），并保留工作流引擎（`openspec/config.yaml` + `.claude/`）。

#### Scenario: 新项目获得干净的 OpenSpec 工作区

- **WHEN** 初始化完成
- **THEN** `openspec/specs/` 与 `openspec/changes/` 被清空，`openspec/config.yaml` 保留且 context 填成新项目名

#### Scenario: 运行时残留被清除

- **WHEN** 初始化完成
- **THEN** `backend/*.db`、`backend/*.db-shm`、`backend/*.db-wal`、`backend/config.yaml` 被删除

### Requirement: 幂等安全

初始化 MUST 幂等安全：若当前 module 名已不是 `base-backend`，则拒绝重复执行并提示。

#### Scenario: 重复初始化被拦截

- **WHEN** 一个已初始化过的项目再次执行 `make init`
- **THEN** 脚本检测到 module 名已非 `base-backend`，拒绝执行并提示「可能已初始化过」
