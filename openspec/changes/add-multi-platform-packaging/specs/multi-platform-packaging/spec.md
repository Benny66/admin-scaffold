# Spec: multi-platform-packaging

多平台交叉编译打包能力，产出可独立部署的 Linux / Windows / 本地部署包。

## ADDED Requirements

### Requirement: 多平台交叉编译打包脚本

项目 MUST 提供 `scripts/package.sh`，支持 `--linux`（amd64）、`--windows`（amd64）、默认本地平台三种目标，交叉编译后端（`CGO_ENABLED=0` + `-trimpath` + `-ldflags "-s -w"`）、构建前端与移动端、组装 `deploy/` 目录、生成压缩包。

#### Scenario: 打包 Linux 部署包

- **WHEN** 执行 `scripts/package.sh --linux`
- **THEN** 产出 Linux amd64 二进制 + `deploy/` 目录 + `base-backend-deploy-linux.tar.gz`

#### Scenario: 打包 Windows 部署包

- **WHEN** 执行 `scripts/package.sh --windows`
- **THEN** 产出 `base-backend.exe` + `deploy/` 目录 + `base-backend-deploy-windows.zip`

#### Scenario: 打包本地平台

- **WHEN** 执行 `scripts/package.sh`（无参数）
- **THEN** 按当前 `go env GOOS/GOARCH` 编译，产出本地平台的部署包

### Requirement: 部署包组装

打包脚本 MUST 组装 `deploy/` 目录，包含：后端二进制、`dist/`（前端）、`dist-mobile/`（移动端）、`config.yaml`（从 example 复制并强制 release）、启动脚本（start.sh / start.bat）。

#### Scenario: 部署目录结构完整

- **WHEN** 打包完成
- **THEN** `deploy/` 下存在后端二进制、dist、dist-mobile、config.yaml、启动脚本，缺一不可

#### Scenario: 配置为 release 模式

- **WHEN** 生成的 `deploy/config.yaml`
- **THEN** 其 `server.mode` 为 `release`，而非开发默认的 `debug`

### Requirement: 剥离业务专属内容

打包脚本 MUST NOT 包含 license 授权逻辑、品牌编译期替换、HTTPS 证书脚本、tutorial 指南、写死的默认密钥。

#### Scenario: 无 license 注入

- **WHEN** 交叉编译后端
- **THEN** ldflags 仅含 `-s -w`，无 `-X .../license.*` 注入，无 trial/paid 分支

#### Scenario: 不写死密钥

- **WHEN** 生成启动脚本
- **THEN** 启动脚本不包含写死的 JWT 密钥，密钥由 config.yaml 的 `jwt.secret` 控制

### Requirement: 单一入口命令

`Makefile` MUST 提供 `package` 目标，透传 `--linux`/`--windows` 参数，使打包无需记忆脚本路径。

#### Scenario: 通过 make 打包

- **WHEN** 执行 `make package --linux`（或等价参数透传）
- **THEN** 等价于执行 `scripts/package.sh --linux`

### Requirement: 产物不入版本库

`.gitignore` MUST 忽略 `deploy/` 目录与 `*.tar.gz`/`*.zip` 打包产物。

#### Scenario: 打包产物不被提交

- **WHEN** 打包完成后 `git status`
- **THEN** `deploy/` 与压缩包不出现在未跟踪文件列表
