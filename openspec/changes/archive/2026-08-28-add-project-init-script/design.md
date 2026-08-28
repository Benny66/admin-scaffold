# Design: add-project-init-script

## Context

基座 README 的手工改名流程（5 步）易错且漏改运行时标识。`base-backend` 标识散布在四类位置：① 模块名（go.mod + ~33 处 import，含 `_example/`）、② 运行时标识（env var 前缀 `BASE_BACKEND_*` 7 处、package.sh 二进制名/压缩包名、JWT Issuer）、③ 默认值（数据库名 `base_backend.db` / `base_backend`）、④ 密钥（`base-backend-secret-key-change-me`）。前置 change `guard-read-module-name` 已使 guard 从 go.mod 动态读模块名，改包名不再拆护栏。

## Goals / Non-Goals

**Goals:**
1. 一条命令完成「改包名 + 改标识 + 重置密钥 + 清历史/残留」。
2. 覆盖 `_example/` 模板 import，杜绝「生成器产出坏代码」的延迟故障。
3. 幂等安全：目标文件已存在/被改过则提示，不静默覆盖用户代码。
4. 零新依赖，仅用系统工具。

**Non-Goals:**
- 不做 `go mod tidy` / `npm install`（依赖安装交给使用方，脚本只改名字）。
- 不改 `gen-module.sh` 自身逻辑。
- 不做「检测基座是否已被初始化过」的复杂状态机（靠幂等检查 + 明确提示即可）。

## Decisions

### D1：脚本参数与命名

`scripts/init.sh <项目名> [--module <go模块名>] [--db-name <名>] [--issuer <名>]`

- `<项目名>`：必填，用于前端/mobile package name、env var 前缀、压缩包名。
- `--module`：Go 模块名，默认从项目名推导（如 `my-system`）；若需域名前缀可显式传 `github.com/foo/my-system`。
- `--db-name` / `--issuer`：可选，默认**不动**（保留基座默认值）。

**Why：** 项目名与 Go module 名不是一回事（module 可带域名），拆开参数更准确。db-name/issuer 是运行时语义，不该机械塞项目名，默认不动、需要时显式覆盖。

### D2：包名替换的精确范围（四类分离处理）

| 类 | 内容 | 处理 |
|---|---|---|
| ① 模块名 | `go.mod` module 行、`import "base-backend/..."`（含 `_example/`） | 全局替换为 `<module>` |
| ② 运行时标识 | env var `BASE_BACKEND_*`（7 处）、package.sh 二进制名/压缩包名、JWT Issuer | 替换为大写项目名（env）/ 项目名（二进制/issuer） |
| ③ 默认值 | 数据库名 `base_backend.db` / `base_backend` | 默认不动，`--db-name` 时替换 |
| ④ 密钥 | `base-backend-secret-key-change-me` | 随机生成，必做 |

**Why：** 四类语义不同，混在一个「全局替换」里会误伤（比如把 db 默认名也换成项目名）。分离处理，每类有自己的默认策略。

### D3：`_example/` 的 import 一并替换

init.sh 的包名替换范围**必须包含** `backend/_example/` 下三个模板文件的 `import "base-backend/..."`。

**Why：** `gen-module.sh` 从 `_example/` 复制生成新模块，若模板 import 仍是旧名，生成器产出坏代码（import 不存在的包）。这是手工流程最容易漏、且延迟暴露的坑。guard 的 `Test_GoldenPathExampleIntact` 只校验文件存在 + package 声明，不校验 import，故替换 import 不会触发 guard 误报。

### D4：幂等与安全护栏

- 替换前先断言「当前 module 名仍是 `base-backend`」，否则提示「可能已初始化过，拒绝重复执行」。
- 目标文件若存在（如 `backend/config.yaml`）则跳过，不覆盖。
- 删除 `openspec/specs`/`changes` 前打印将删内容，要求 `--yes` 或交互确认。

**Why：** init 是破坏性操作（删历史 + 改全仓），幂等检查防「重复初始化把 `my-system` 又替换成 `another-system`」的误伤。

### D5：随机密钥生成

用 `openssl rand -hex 32` 生成 64 字符十六进制密钥；`openssl` 不存在则回退 `cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 64`。替换 `config.go` 默认值 + `config.example.yaml` 两处。

**Why：** 密钥是安全刚需，复制上线前必须换。生成放 init 阶段，杜绝「所有项目共用 `change-me` 密钥」。

### D6：OpenSpec 历史的处理

删除 `openspec/specs/`、`openspec/changes/`；保留 `openspec/config.yaml`，并把其空 `context:` 填成新项目技术栈/领域。

**Why：** specs/changes 是基座自己的身份，跟业务项目无关；config.yaml + `.claude/` 是工作流引擎，必须保留。context 从空模板填成项目名，让后续 openspec 操作有正确上下文。

## Risks / Trade-offs

- [替换误伤] → 四类分离 + 精确锚点（`import "base-backend/`、`BASE_BACKEND_` 前缀、`base-backend-secret-key-change-me` 整串）降低误伤；D4 幂等检查兜底。
- [脚本破坏性] → 删除操作需确认，且只删明确的 `openspec/specs`/`changes` 与 `*.db`，不碰代码。
- [sed 跨平台差异] → 沿用 `gen-module.sh` 既有做法（用 `sed` 输出到临时文件再 mv，或用 python3 兜底），避免 `sed -i` 的 BSD/GNU 差异。
- [env var 前缀命名] → 项目名含连字符（如 `my-system`）时，env var 前缀需转大写并去连字符（`MY_SYSTEM_`），脚本内做规范化。

## Open Questions

- **init.sh 是否也初始化 git**（`git init` + 首次 commit）：当前不做，交给使用方；若需要可加 `--git` 参数。暂不纳入。
- **是否生成 config.yaml 而非只留 example**：当前删掉 config.yaml 让首次启动用默认值，与 README 现状一致；若希望 init 直接生成可运行的 config.yaml，需另定。暂不纳入。
