# Design: add-multi-platform-packaging

## Context

资产系统 `build.sh`（20884 字节）已沉淀了成熟的多平台打包能力，但混入了四类内容：A 多平台交叉编译打包（通用）、B 品牌编译期替换（logo/favicon 临时替换）、C license 授权（trial/paid + ldflags 注入）、D 业务专属部署附件（HTTPS 证书、tutorial、写死密钥）。

本 change 只提炼 **A**，剥离 B/C/D。依赖前置 change `add-static-serving`（打包产物需后端托管才能运行）。

## Goals / Non-Goals

**Goals:**

1. 一键产出 Linux amd64 / Windows amd64 / 本地平台的可部署包。
2. 打包流程：交叉编译后端 → 构建前端/移动端 → 组装 deploy/ → 打 tar.gz/zip。
3. 生成的包「双击即跑」（含 start.sh / start.bat）。
4. 模块名、二进制名、配置均泛化到基座（`base-backend`），不残留资产系统痕迹。

**Non-Goals:**

- 剥离 license 授权（trial/paid、BUILD_KEY、ldflags 注入 license.Mode 等）。
- 剥离品牌编译期替换（logo/favicon 由 brand-config 运行时机制负责，打包脚本不碰）。
- 剥离 HTTPS 证书脚本、mkcert、tutorial 指南。
- 不做付费版/试用版切换。
- 不做跨编译 ARM 架构（当前只 amd64，可后续扩展）。

## Decisions

### D1：脚本独立为 `scripts/package.sh`，Makefile 提供入口

核心逻辑放 `scripts/package.sh`（与资产系统 `build.sh` 习惯一致），`Makefile` 加 `package` 目标透传参数。

**Why：** 脚本可独立运行（`./scripts/package.sh --linux`），Makefile 提供 `make package --linux` 快捷入口，两者不重复实现。

### D2：二进制名与 module 名统一为 `base-backend`

二进制名 `base-backend`，Windows 为 `base-backend.exe`，压缩包 `base-backend-deploy-<platform>.tar.gz`/`.zip`。

**Why：** 与 `backend/go.mod` 的 `module base-backend` 一致，泛化到基座，无资产系统痕迹。

**Alternatives considered：**
- 沿用 `asset-management` → 拒绝，残留业务名。

### D3：剥离 license 注入 —— ldflags 只保留 `-s -w`

交叉编译的 ldflags 只保留 `-s -w -trimpath`，去掉 `-X asset-management/license.*` 与 `BUILD_KEY`。

**Why：** 脚手架无 license 模块，注入会编译失败；`-s -w` 去掉符号表与调试信息，减小二进制体积。

### D4：config.yaml 从 config.example.yaml 复制 + 强制 release

打包脚本将 `backend/config/config.example.yaml` 复制为 `deploy/config.yaml`，并强制 `mode: "release"`。

**Why：** 基座无 brand.config 拼接逻辑（品牌走运行时 config），从 example 复制最干净；release 模式是部署默认。

**Alternatives considered：**
- 复制用户自己的 `config.yaml`（若存在）→ 拒绝，config.yaml 被 gitignore，打包脚本不应依赖运行时生成的文件；example 才是稳定基线。

### D5：启动脚本泛化，去除写死密钥

start.sh / start.bat 只设置 `GIN_MODE=release`（对应基座的 `BASE_BACKEND_GIN_MODE`），不再写死 JWT 密钥（用户自行在 config.yaml 配置）。

**Why：** 资产系统的 `asset-management-secret-key-2024` 是安全隐患，基座不写死密钥；JWT 密钥由 config.yaml 的 `jwt.secret` 控制。

### D6：环境变量名对齐基座

启动脚本用 `BASE_BACKEND_GIN_MODE=release`（而非资产系统的 `ASSET_ADMIN_GIN_MODE`），与 `config.go` 的环境变量命名一致。

## Risks / Trade-offs

- [交叉编译 Windows 时 `zip` 命令在 Linux CI 可能缺失] → 脚本检测 zip 缺失时给出明确提示（brew install zip / apt install zip），不静默失败。
- [前端/移动端 build 需 node_modules，首次打包耗时] → 脚本检测 node_modules 不存在才 npm install（与资产系统一致），复用本地缓存。
- [`-s -w` 去除符号表后 panic 堆栈可读性下降] → 部署包可接受（体积优先）；开发态仍用 `go run` 保留符号。
- [打包脚本引用的 dist/dist-mobile 依赖前置 change 的托管路径] → 产物目录命名与 `add-static-serving` 严格对齐（dist、dist-mobile），写进 tasks 交叉验证。

## Open Questions

- **是否支持 ARM（linux/arm64）**：当前只 amd64，与资产系统一致；若未来要部署到树莓派/ARM 服务器再扩展。暂不支持。
- **前端 build 的 base 路径**：当前前端 `vite.config.js` 的 build 无 `base` 配置（默认 `/`），与后端根路径托管一致；若未来需子路径部署需联动调整，暂不处理。
