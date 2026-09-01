# Spec: unified-port-config

端口单一真相，后端端口改动自动联动前端/移动端 dev proxy，消除三处硬编码漂移。

## ADDED Requirements

### Requirement: vite proxy 端口可配置

前端与移动端 vite.config.js 的 dev proxy target MUST 从 `.env` 的 `VITE_API_BASE` 读取，缺省回退 `http://localhost:8080`。

#### Scenario: 配置了 VITE_API_BASE

- **WHEN** `frontend/.env` 或 `mobile/.env` 设置 `VITE_API_BASE=http://localhost:9090`
- **THEN** dev 模式下 vite proxy 将 `/api` 与 `/static` 转发到 9090 而非 8080

#### Scenario: 未配置 VITE_API_BASE

- **WHEN** 不存在 `.env` 文件
- **THEN** vite proxy 回退到 `http://localhost:8080`，行为与现状一致

### Requirement: .env 模板提供

前端与移动端 MUST 提供 `.env.example`，说明 `VITE_API_BASE` 的用途与默认值。

#### Scenario: 查看模板

- **WHEN** 查看 `frontend/.env.example` 与 `mobile/.env.example`
- **THEN** 内容包含 `VITE_API_BASE=http://localhost:8080` 及用途注释

### Requirement: init.sh 支持端口参数

`scripts/init.sh` MUST 支持 `--port <端口>` 参数，初始化时生成 `frontend/.env` 与 `mobile/.env`，内容指向该端口。

#### Scenario: 带端口初始化

- **WHEN** 执行 `scripts/init.sh my-system --port 9090`
- **THEN** 生成 `frontend/.env` 与 `mobile/.env`，`VITE_API_BASE=http://localhost:9090`

#### Scenario: 不带端口初始化

- **WHEN** 执行 `scripts/init.sh my-system`（未传 --port）
- **THEN** 不生成 `.env`（沿用缺省 8080）

### Requirement: .env 不入版本库

`.gitignore` MUST 忽略 `.env`（但保留 `.env.example` 提交）。

#### Scenario: 检查 git 状态

- **WHEN** 项目生成 `.env` 后执行 `git status`
- **THEN** `.env` 不出现在未跟踪列表，`.env.example` 正常提交
