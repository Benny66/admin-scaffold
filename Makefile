# Base 脚手架统一入口：AI 只需记住 make <target>，无需记忆零散脚本路径。
# 目标：test / lint / smoke / gen / build / dev

.PHONY: test lint smoke gen build dev dev-backend dev-frontend

# 后端全部测试（含 internal/guard 架构护栏）
test:
	cd backend && go test ./...

# 静态检查：后端 vet + 前端 ESLint + 移动端 ESLint
lint:
	cd backend && go vet ./...
	cd frontend && npm run lint
	cd mobile && npm run lint

# 冒烟：构建 → 启动 → 登录 → 命中受保护路由 → 断言 → 清理
smoke:
	backend/scripts/smoke.sh

# 生成新模块骨架（用法：make gen name=<模块名>）
gen:
	@if [ -z "$(name)" ]; then \
		echo "用法: make gen name=<模块名>（如 make gen name=asset）"; \
		exit 1; \
	fi
	backend/scripts/gen-module.sh "$(name)"

# 多平台打包（用法：make package           = 本地平台
#              make package TARGET=--linux    = Linux amd64
#              make package TARGET=--windows  = Windows amd64）
package:
	scripts/package.sh $(TARGET)

# 构建后端
build:
	cd backend && go build ./...

# 开发：并行启动后端 + 前端（Ctrl-C 同时退出）
dev:
	$(MAKE) -j2 dev-backend dev-frontend

dev-backend:
	cd backend && go run main.go

dev-frontend:
	cd frontend && npm run dev
