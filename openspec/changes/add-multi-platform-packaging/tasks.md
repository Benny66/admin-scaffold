# Tasks: add-multi-platform-packaging

## 1. 打包脚本

- [x] 1.1 新增 `scripts/package.sh`：参数解析（--linux/--windows/默认本地），目标平台与二进制名确定
- [x] 1.2 环境检测：go/node/npm 存在性检查，Windows 目标检查 zip 命令
- [x] 1.3 交叉编译后端：`CGO_ENABLED=0` + `GOOS/GOARCH` + `-trimpath` + `-ldflags "-s -w"`，二进制名 `base-backend`（Windows 加 .exe）
- [x] 1.4 构建前端与移动端：检测 node_modules 缺失才 npm install，然后 npm run build
- [x] 1.5 组装 `deploy/`：复制二进制 + dist + dist-mobile + 生成 config.yaml（从 example 复制并强制 release）+ 生成 start.sh / start.bat
- [x] 1.6 打压缩包：Linux/Mac 用 tar.gz，Windows 用 zip，产物命名 `base-backend-deploy-<platform>`

## 2. 接线

- [x] 2.1 `Makefile` 加 `package` 目标，透传参数到 `scripts/package.sh`
- [x] 2.2 `.gitignore` 忽略 `deploy/` 与 `*.tar.gz`/`*.zip`

## 3. 文档与验证

- [x] 3.1 更新 `docs/配置体系.md`：打包与部署流程说明
- [x] 3.2 交叉验证：确认 `scripts/package.sh` 产出的 dist/dist-mobile 目录名与 `add-static-serving` 的托管路径严格一致
- [x] 3.3 本地平台打包 smoke：跑 `scripts/package.sh`（无参数），确认 deploy/ 目录结构与压缩包生成
- [x] 3.4 验证 deploy/config.yaml 为 release 模式、启动脚本无写死密钥
