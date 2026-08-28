# Tasks: add-project-init-script

## 1. init.sh 主体

- [x] 1.1 新增 `scripts/init.sh`：参数解析（项目名必填、--module/--db-name/--issuer 可选），模块名默认从项目名推导
- [x] 1.2 幂等检查：断言当前 `go.mod` module 名仍为 `base-backend`，否则拒绝并提示
- [x] 1.3 替换模块名：`go.mod` module 行 + 全量 `import "base-backend/..."`（含 `_example/` 三个模板）
- [x] 1.4 替换运行时标识：env var 前缀 `BASE_BACKEND_*` → 大写去连字符项目名；package.sh 二进制名/压缩包名 → 项目名；JWT Issuer → 项目名（或 --issuer）
- [x] 1.5 重置密钥：`base-backend-secret-key-change-me` → 随机串（openssl rand -hex 32，回退 /dev/urandom），替换 config.go 默认值 + config.example.yaml
- [x] 1.6 数据库名：默认不动，传 `--db-name` 时替换 config.go + config.example.yaml 的 `base_backend`
- [x] 1.7 前端/mobile package name：`base-backend-frontend`/`base-backend-mobile` → `<项目名>-frontend`/`<项目名>-mobile`

## 2. 清理历史与残留

- [x] 2.1 清 OpenSpec 历史：删除 `openspec/specs/`、`openspec/changes/`，保留 `openspec/config.yaml` 并把 context 填成新项目名
- [x] 2.2 清运行时残留：删除 `backend/*.db`、`*.db-shm`、`*.db-wal`、`backend/config.yaml`
- [x] 2.3 删除操作前打印将删内容并要求确认（`--yes` 跳过确认）

## 3. 接线与文档

- [x] 3.1 `Makefile` 新增 `init` 目标，透传 name/module/db-name/issuer 到 scripts/init.sh
- [x] 3.2 更新 README「如何基于基座新建项目」段，改为 `make init name=<项目名>` 一键流程

## 4. 验证

- [x] 4.1 在临时副本目录执行 `make init name=my-system`，确认：go.mod/import（含 _example）改名、env var 前缀 MY_SYSTEM_、密钥已换、前端 package 改名、specs/changes 已清空、config.yaml 已删
- [x] 4.2 初始化后 `cd backend && go test ./...`（含 guard）通过，`go build ./...` 通过
- [x] 4.3 初始化后 `make gen name=asset` 生成模块，确认 import 为新模块名且编译通过
- [x] 4.4 反向验证：重复执行 init，确认幂等检查拒绝
- [x] 4.5 反向验证：不传 --db-name，确认数据库默认名保持 `base_backend` 不变
