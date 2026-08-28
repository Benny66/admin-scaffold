# Proposal: add-project-init-script

## Why

README 第 112-120 行把「基于基座新建项目」写成 5 步**手工**流程：复制目录 → 改 `go.mod` module 名 + 全局替换 import → 改前端 package name → 删运行时残留 → 改配置。

这套手工流程有四个痛点：

1. **易错**：全局替换 `base-backend` 若漏掉 `_example/*.go` 的 import，之后 `make gen` 生成的模块会 import 一个不存在的包，编译失败——问题在"生成器产出坏代码"时才暴露，与改名的时点相距甚远。
2. **漏改运行时标识**：`base-backend` 不只出现在 import，还散布在 env var 前缀 `BASE_BACKEND_*`（config.go 7 处）、package.sh 的二进制名/压缩包名、JWT Issuer、数据库默认名里。手工很难改全。
3. **密钥残留**：`base-backend-secret-key-change-me` 是基座默认密钥，复制上线是安全隐患。
4. **OpenSpec 历史残留**：基座的 `openspec/specs/`（5 个脚手架能力）+ `changes/`（基座开发史）会跟着复制进新项目，与项目自己的 spec 混在一起。

本 change 提供一个 `scripts/init.sh`，把「复制基座 → 改名 → 清理」固化为一键命令，消除手工流程的隐患。

## What Changes

- 新增 `scripts/init.sh <项目名> [--module <go模块名>] [--db-name <名>] [--issuer <名>]`：
  - **改包名**：`go.mod` module 行 + 全量 `import "base-backend/..."`（含 `_example/` 模板）+ 前端/mobile 的 package name。
  - **改运行时标识**：env var 前缀 `BASE_BACKEND_*` → 大写项目名、package.sh 的二进制名/压缩包名、JWT Issuer（默认项目名）。
  - **重置密钥**：`base-backend-secret-key-change-me` → 随机串，写入 `config.go` 默认值 + `config.example.yaml`。
  - **清 OpenSpec 历史**：删除 `openspec/specs/`、`openspec/changes/`，保留 `openspec/config.yaml` 并把 context 填成新项目名。
  - **清运行时残留**：删除 `backend/*.db*`、`backend/config.yaml`。
- `Makefile` 新增 `init` 目标透传参数。
- 更新 README「如何基于基座新建项目」段，指向 `make init`。
- 前置依赖 change `guard-read-module-name`（已完成）——guard 已从 go.mod 动态读模块名，本脚本可无脑替换 import 而不拆护栏。

## Capabilities

### New Capabilities

- `project-init`: 一键初始化脚本，把「复制基座 → 改包名/标识 → 重置密钥 → 清历史/残留」固化为可重复命令。

### Modified Capabilities

（无。）

## Impact

- **新增文件**：`scripts/init.sh`。
- **修改文件**：`Makefile`（加 `init` 目标）、`README.md`（新建项目段改写）。
- **无新第三方依赖**：init.sh 是 bash 脚本，仅用 `sed`/`grep`/`openssl`/`uuidgen` 等系统工具（随机密钥用 `openssl rand`，不存在则回退 `/dev/urandom`）。
- **前置依赖**：`guard-read-module-name`（已归档/待归档）——若未先做，改包名会拆掉 controller 直连 database 的护栏。
- **范围外**：不改 `gen-module.sh` 自身（它从 `_example/` 复制，import 名已由 init.sh 一并替换）、不做依赖安装（`npm install`/`go mod tidy` 交给使用方）、不改基座现有默认值语义（db-name/issuer 默认不动）。

## 待确认的决策（已在 design.md 记录）

- 密钥重置：必做（安全刚需）。
- 数据库默认名 / JWT Issuer：可选参数，默认不动（避免机械塞项目名污染语义）。
