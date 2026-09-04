# Base 脚手架统一入口：AI 只需记住 make <target>，无需记忆零散脚本路径。
# 目标：test / lint / smoke / gen / build / dev

.PHONY: test lint smoke gen init build dev dev-backend dev-frontend dev-mp build-mp

# 后端全部测试（含 internal/guard 架构护栏）
test:
	cd backend && go test ./...

# 静态检查：后端 vet + 前端 ESLint + 移动端 ESLint + 小程序 ESLint
lint:
	cd backend && go vet ./...
	cd frontend && npm run lint
	cd mobile && npm run lint
	cd miniapp && npm run lint

# 冒烟：构建 → 启动 → 登录 → 命中受保护路由 → 断言 → 清理
smoke:
	backend/scripts/smoke.sh

# 生成新模块骨架（用法：make gen name=<模块名> [group=<分组 path>]）
#   - 不传 group：自建分组（path=<复数>，首叶子 path:''，URL 为 /<复数>）
#   - 传 group=<已存在分组 path>：把新模块注入该分组的 children（URL 为 /<group>/<复数>）
gen:
	@if [ -z "$(name)" ]; then \
		echo "用法: make gen name=<模块名>（如 make gen name=asset，可选 group=<分组>）"; \
		exit 1; \
	fi
	backend/scripts/gen-module.sh "$(name)" \
		$$([ -n "$(group)" ] && echo "group=$(group)")

# 一键初始化新项目（用法：make init name=<项目名> [module=...] [db_name=...] [issuer=...] [app_name=...]）
init:
	@if [ -z "$(name)" ]; then \
		echo "用法: make init name=<项目名>（如 make init name=my-system）"; \
		exit 1; \
	fi
	scripts/init.sh "$(name)" \
		$$([ -n "$(module)" ] && echo --module "$(module)") \
		$$([ -n "$(db_name)" ] && echo --db-name "$(db_name)") \
		$$([ -n "$(issuer)" ] && echo --issuer "$(issuer)") \
		$$([ -n "$(app_name)" ] && echo --app-name "$(app_name)") \
		$$([ -n "$(port)" ] && echo --port "$(port)")

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

# 小程序 dev（需配合微信开发者工具打开 miniapp/dist/dev/mp-weixin/）
dev-mp:
	cd miniapp && npm run dev:mp-weixin

# 构建小程序 mp-weixin 产物
build-mp:
	cd miniapp && npm run build:mp-weixin
