# Proposal: add-multi-platform-packaging

## Why

脚手架目前只有开发态启动方式（`go run` + `vite dev`），无法产出可部署的多平台包。资产系统的 `build.sh` 已沉淀了一套多平台交叉编译打包能力，但其混入了 license 授权、品牌编译期替换、证书/tutorial 等业务专属逻辑，不能直接搬入业务无关的脚手架基座。本 change 提炼其中「多平台交叉编译 + 组装部署目录 + 压缩」的纯工程能力，配合已完成的 `add-static-serving`，让脚手架能一键产出「双击即跑」的 Linux / Windows / 本地三平台部署包。

## What Changes

- 新增 `scripts/package.sh`：多平台交叉编译打包脚本，提炼自资产系统 `build.sh` 的纯工程部分。
  - 参数：`--linux`（amd64）、`--windows`（amd64）、默认本地平台。
  - 流程：交叉编译后端（`CGO_ENABLED=0` + `-trimpath` + `-ldflags "-s -w"`）→ 构建前端/移动端 → 组装 `deploy/` → 打 `tar.gz`/`zip`。
  - 生成 `deploy/config.yaml`（从 `config.example.yaml` 复制，强制 `mode: release`）。
  - 生成 `deploy/start.sh`（Linux/Mac）与 `deploy/start.bat`（Windows）一键启动脚本。
- 明确**剥离**（与资产系统 `build.sh` 的差异）：license 授权（trial/paid）、品牌编译期替换（logo/favicon）、HTTPS 证书脚本、tutorial 指南、写死的默认密钥。
- `Makefile` 新增 `package` 目标（透传 `--linux`/`--windows`）。
- `.gitignore` 忽略 `deploy/` 与 `*.tar.gz`/`*.zip` 产物。
- 更新 `docs/配置体系.md`，说明打包与部署流程。

## Capabilities

### New Capabilities

- `multi-platform-packaging`: 多平台交叉编译打包能力，产出可独立部署的 Linux / Windows / 本地部署包。

### Modified Capabilities

（无。）

## Impact

- 新增文件：`scripts/package.sh`、修改 `Makefile`、`.gitignore`、`docs/配置体系.md`。
- 依赖前置 change：`add-static-serving`（打包产物需后端托管才能运行）。
- 无新第三方依赖（仅用 `go`、`node`、`npm`、`tar`、`zip` 系统工具）。
- 二进制默认名 `base-backend`（与 `backend/go.mod` 的 module 名一致）。
