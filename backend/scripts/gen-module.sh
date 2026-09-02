#!/usr/bin/env bash
# 模块代码生成器：从「黄金路径」唯一范例 _example/ 生成新模块的完整骨架。
#
# 用法：backend/scripts/gen-module.sh <模块名>    （或 make gen name=<模块名>）
#   模块名用小写（如 asset、asset_category），生成：
#     backend/models/<name>.go
#     backend/services/<name>_service.go
#     backend/controllers/<name>.go
#     router.go 注入路由块（【gen:routes】锚点前）
#     database.go 注入 AutoMigrate（【gen:migrate】锚点后）
#     database.go 注入权限码（【gen:permissions】锚点后）
#     frontend/src/views/<name>/index.vue
#     frontend/src/api/index.js 追加 API 定义
#     frontend/src/router/index.js 注入路由条目（【gen:route】锚点前）
#
# 路由路径与权限码前缀统一使用资源复数（/assets、assets:view），由 scripts/pluralize.sh
# 产出——它是复数规则的单一真相，本脚本与 guard 测试共同调用。
#
# 生成后每个文件带 `// TODO: 业务逻辑` 锚点，AI 只需填充业务逻辑。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$BACKEND_DIR/.." && pwd)"

# ---------- 参数解析 ----------
NAME="${1:-}"
if [[ -z "$NAME" ]]; then
  echo "错误：缺少模块名" >&2
  echo "用法：$0 <模块名>（如 $0 asset）" >&2
  exit 1
fi

# 校验模块名：小写字母/数字/下划线/连字符
if [[ ! "$NAME" =~ ^[a-z][a-z0-9_-]*$ ]]; then
  echo "错误：模块名必须是小写字母/数字/下划线/连字符，且以字母开头（当前：$NAME）" >&2
  exit 1
fi

# PascalCase 转换：asset_category → AssetCategory
PASCAL="$(printf '%s' "$NAME" | awk -F'[-_]' '{ for(i=1;i<=NF;i++) printf "%s%s", toupper(substr($i,1,1)), substr($i,2) }')"
EXAMPLE_DIR="$BACKEND_DIR/_example"

# 资源复数名：asset → assets，用于路由路径与权限码前缀（与存量 /users、roles:view 一致）。
# 复数规则的单一真相是 scripts/pluralize.sh——本脚本与 guard 测试共同调用它，
# 避免规则在两处各写一份而漂移。
PLURALIZE_SCRIPT="$SCRIPT_DIR/pluralize.sh"
if [[ ! -f "$PLURALIZE_SCRIPT" ]]; then
  echo "错误：复数化脚本缺失：$PLURALIZE_SCRIPT" >&2
  exit 1
fi
PLURAL="$(bash "$PLURALIZE_SCRIPT" "$NAME")"
if [[ -z "$PLURAL" ]]; then
  echo "错误：pluralize.sh 对「$NAME」返回空结果" >&2
  exit 1
fi

# ---------- 前置校验（fail fast）----------
# 锚点缺失与资源占用都必须在动任何文件之前检查完。否则中途失败会留下
# 「后端已生成、前端未注入」的不一致状态，而重跑又会被 require_new 拒绝，
# 只能手工清理。
ROUTER_FILE="$BACKEND_DIR/router/router.go"
DB_FILE="$BACKEND_DIR/database/database.go"
FRONTEND_ROUTER="$REPO_ROOT/frontend/src/router/index.js"

# grep -F：锚点含中文方括号，按固定字符串匹配更稳妥
check_anchor() {
  local file="$1" anchor="$2" label="$3"
  if ! grep -qF "$anchor" "$file"; then
    echo "错误：$label 缺少 $anchor 锚点——生成器无法定位注入位置" >&2
    exit 1
  fi
}
check_anchor "$ROUTER_FILE" '【gen:routes】' 'router.go'
check_anchor "$DB_FILE" '【gen:migrate】' 'database.go'
check_anchor "$DB_FILE" '【gen:permissions】' 'database.go'
check_anchor "$FRONTEND_ROUTER" '【gen:route】' 'frontend/src/router/index.js'

# 资源占用检查：复数化后新模块可能与存量五件套撞车
# （如 make gen name=user 会产出 /users 与 users:view，而两者都已存在）。
# 撞车的后果比过去更严重——Gin 重复注册路由会直接 panic。
if grep -qF "Group(\"/${PLURAL}\")" "$ROUTER_FILE"; then
  echo "错误：路由 /${PLURAL} 已被占用（router.go 中已存在该路由组），请换一个模块名" >&2
  exit 1
fi
if grep -qF "\"${PLURAL}:view\"" "$DB_FILE"; then
  echo "错误：权限码 ${PLURAL}:view 已存在，请换一个模块名" >&2
  exit 1
fi
if grep -qF "path: '${PLURAL}'" "$FRONTEND_ROUTER"; then
  echo "错误：前端路由 path '${PLURAL}' 已被占用，请换一个模块名" >&2
  exit 1
fi

# ---------- 工具函数 ----------
# replace <src> <dst>：替换占位符（先 Pascal 后小写），输出到新文件（可移植，避开 sed -i 差异）
replace() {
  local src="$1" dst="$2"
  sed "s/Example/${PASCAL}/g; s/example/${NAME}/g" "$src" > "$dst"
}

# 幂等检查：目标文件已存在则报错，避免覆盖用户业务代码
require_new() {
  local path="$1"
  if [[ -e "$path" ]]; then
    echo "错误：目标已存在，拒绝覆盖：$path（请先删除或换模块名）" >&2
    exit 1
  fi
}

# ---------- 生成后端 ----------
echo "==> 生成后端模块「${NAME}」（PascalCase: ${PASCAL}）"

mkdir -p "$BACKEND_DIR/models" "$BACKEND_DIR/services" "$BACKEND_DIR/controllers"

require_new "$BACKEND_DIR/models/${NAME}.go"
replace "$EXAMPLE_DIR/models/example.go" "$BACKEND_DIR/models/${NAME}.go"
echo "  ✓ models/${NAME}.go"

require_new "$BACKEND_DIR/services/${NAME}_service.go"
replace "$EXAMPLE_DIR/services/example_service.go" "$BACKEND_DIR/services/${NAME}_service.go"
echo "  ✓ services/${NAME}_service.go"

require_new "$BACKEND_DIR/controllers/${NAME}.go"
replace "$EXAMPLE_DIR/controllers/example.go" "$BACKEND_DIR/controllers/${NAME}.go"
echo "  ✓ controllers/${NAME}.go"

# ---------- 注入 router.go ----------
# ROUTER_FILE 与 【gen:routes】 锚点已在前置校验中确认
ROUTE_BLOCK=$(cat <<EOF

	// ==================== ${PASCAL} 管理 ====================
	${NAME}Group := protected.Group("/${PLURAL}")
	{
		${NAME}Group.GET("", middleware.PermissionRequired("${PLURAL}:view"), controllers.Get${PASCAL}List)
		${NAME}Group.GET("/:id", middleware.PermissionRequired("${PLURAL}:view"), controllers.Get${PASCAL})
		${NAME}Group.POST("", middleware.PermissionRequired("${PLURAL}:create"), controllers.Create${PASCAL})
		${NAME}Group.PUT("/:id", middleware.PermissionRequired("${PLURAL}:edit"), controllers.Update${PASCAL})
		${NAME}Group.DELETE("/:id", middleware.PermissionRequired("${PLURAL}:delete"), controllers.Delete${PASCAL})
	}
EOF
)
# 在锚点行之前插入路由块（按行插入，跨平台可靠）
# 注意：命令替换 $(cat <<EOF ...) 会剥离尾部换行，必须补回，否则注入块末尾会与
# 下一行粘连（既有的路由注入就曾因此产出 `}\t// 【gen:routes】`）。
ROUTE_BLOCK="$ROUTE_BLOCK" python3 -c '
import os, sys
path = sys.argv[1]
block = os.environ["ROUTE_BLOCK"]
if not block.endswith("\n"):
    block += "\n"
lines = open(path, encoding="utf-8").read().splitlines(keepends=True)
out = []
for ln in lines:
    if "【gen:routes】" in ln:
        out.append(block)
    out.append(ln)
open(path, "w", encoding="utf-8").write("".join(out))
' "$ROUTER_FILE"
echo "  ✓ router.go 已注入路由块"

# ---------- 注入 database.go AutoMigrate ----------
# DB_FILE 与 【gen:migrate】 锚点已在前置校验中确认
MIGRATE_LINE=$'\t\t&models.'${PASCAL}'{},'
# 在锚点行之后插入模型行（按行插入，跨平台可靠）
MIGRATE_LINE="$MIGRATE_LINE" python3 -c '
import os, sys
line = os.environ["MIGRATE_LINE"]
path = sys.argv[1]
lines = open(path, encoding="utf-8").read().splitlines(keepends=True)
out = []
for ln in lines:
    out.append(ln)
    if "【gen:migrate】" in ln:
        out.append(line + "\n")
open(path, "w", encoding="utf-8").write("".join(out))
' "$DB_FILE"
echo "  ✓ database.go 已注入 AutoMigrate"

# ---------- 注入 database.go initBaseData 权限码 ----------
# 不写 Sort：它在 database.go 的 syncPermissions 中按当前表内最大值递增计算，
# 写死会与存量权限的排序值冲突（见 design D3）。
PERM_ANCHOR='【gen:permissions】'
PERM_BLOCK=$(cat <<EOF
		// ${PASCAL} 模块权限（TODO: 把 Name 改成中文业务名称）
		{Name: "查看 ${PASCAL}", Code: "${PLURAL}:view", Type: "api", Status: 1},
		{Name: "创建 ${PASCAL}", Code: "${PLURAL}:create", Type: "api", Status: 1},
		{Name: "编辑 ${PASCAL}", Code: "${PLURAL}:edit", Type: "api", Status: 1},
		{Name: "删除 ${PASCAL}", Code: "${PLURAL}:delete", Type: "api", Status: 1},
EOF
)
# 在锚点行之后插入（按行插入，跨平台可靠）
# 同样补回被命令替换剥离的尾部换行，否则多个模块的权限块会首尾粘连。
PERM_ANCHOR="$PERM_ANCHOR" PERM_BLOCK="$PERM_BLOCK" python3 -c '
import os, sys
anchor = os.environ["PERM_ANCHOR"]
block = os.environ["PERM_BLOCK"]
if not block.endswith("\n"):
    block += "\n"
path = sys.argv[1]
lines = open(path, encoding="utf-8").read().splitlines(keepends=True)
out = []
for ln in lines:
    out.append(ln)
    if anchor in ln:
        out.append(block)
open(path, "w", encoding="utf-8").write("".join(out))
' "$DB_FILE"
echo "  ✓ database.go 已注入 initBaseData 权限码（${PLURAL}:view/create/edit/delete）"

# ---------- 生成前端 ----------
FRONTEND_VIEWS="$REPO_ROOT/frontend/src/views/${NAME}"
require_new "$FRONTEND_VIEWS/index.vue"
mkdir -p "$FRONTEND_VIEWS"
replace "$EXAMPLE_DIR/frontend/index.vue" "$FRONTEND_VIEWS/index.vue"
echo "  ✓ frontend/src/views/${NAME}/index.vue"

API_FILE="$REPO_ROOT/frontend/src/api/index.js"
API_BLOCK=$(cat <<EOF

// ==================== ${PASCAL} 管理 ====================
export const get${PASCAL}List = (params) => request.get('/${PLURAL}', { params })
export const get${PASCAL} = (id) => request.get(\`/${PLURAL}/\${id}\`)
export const create${PASCAL} = (data) => request.post('/${PLURAL}', data)
export const update${PASCAL} = (id, data) => request.put(\`/${PLURAL}/\${id}\`, data)
export const delete${PASCAL} = (id) => request.delete(\`/${PLURAL}/\${id}\`)
EOF
)
printf '%s\n' "$API_BLOCK" >> "$API_FILE"
echo "  ✓ frontend/src/api/index.js 已追加 API 定义"

# ---------- 注入 frontend/src/router/index.js 路由条目 ----------
# 菜单由 Layout.vue 从路由派生并按 meta.permission 过滤，故无需改动 Layout.vue。
ROUTE_ANCHOR='【gen:route】'
FE_ROUTE_BLOCK=$(cat <<EOF
      {
        path: '${PLURAL}',
        name: '${PASCAL}',
        component: () => import('@/views/${NAME}/index.vue'),
        // TODO: 把 title 换成中文菜单标题，并为该模块挑一个合适的图标
        meta: { title: '${PASCAL}', icon: 'Document', permission: '${PLURAL}:view' },
      },
EOF
)
# 在锚点行之前插入（按行插入，跨平台可靠）
# 同样补回被命令替换剥离的尾部换行。
ROUTE_ANCHOR="$ROUTE_ANCHOR" FE_ROUTE_BLOCK="$FE_ROUTE_BLOCK" python3 -c '
import os, sys
anchor = os.environ["ROUTE_ANCHOR"]
block = os.environ["FE_ROUTE_BLOCK"]
if not block.endswith("\n"):
    block += "\n"
path = sys.argv[1]
lines = open(path, encoding="utf-8").read().splitlines(keepends=True)
out = []
for ln in lines:
    if anchor in ln:
        out.append(block)
    out.append(ln)
open(path, "w", encoding="utf-8").write("".join(out))
' "$FRONTEND_ROUTER"
echo "  ✓ frontend/src/router/index.js 已注入路由条目（菜单自动出现，无需改动 Layout.vue）"

# ---------- 完成提示 ----------
cat <<EOF

✅ 模块「${NAME}」骨架生成完毕（路由 /${PLURAL}，权限码 ${PLURAL}:view/create/edit/delete）。

已自动生成，无需手工处理：
  · 后端三层 models / services / controllers
  · router.go 路由注册（挂载在 /${PLURAL}）
  · database.go 的 AutoMigrate 与 initBaseData 权限码
  · 前端页面 views/${NAME}/index.vue 与 API 定义
  · 前端路由条目 router/index.js
    （菜单由 Layout.vue 从路由派生并按 meta.permission 过滤，无需也禁止改动 Layout.vue）

后续手工步骤（生成器不自动处理，需 AI/开发者完成）：
1. 填充业务逻辑：搜索「// TODO: 业务逻辑」锚点，实现唯一性校验、级联、关联等。
2. 替换占位文案：前端路由条目的 title / icon，以及 initBaseData 中的权限中文名称。
3. 验证：在仓库根目录执行 make test && make smoke
   （Makefile 只在根目录提供，backend/ 下没有 Makefile，故 make 目标不能在那里跑；
    若只需验证后端，可在 backend/ 下执行 go build ./... && go test ./...）
EOF
