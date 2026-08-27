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
#     frontend/src/views/<name>/index.vue
#     frontend/src/api/index.js 追加 API 定义
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
ROUTER_FILE="$BACKEND_DIR/router/router.go"
if ! grep -q '【gen:routes】' "$ROUTER_FILE"; then
  echo "错误：router.go 缺少 【gen:routes】 锚点" >&2
  exit 1
fi
ROUTE_BLOCK=$(cat <<EOF

	// ==================== ${PASCAL} 管理 ====================
	${NAME}Group := protected.Group("/${NAME}")
	{
		${NAME}Group.GET("", middleware.PermissionRequired("${NAME}:view"), controllers.Get${PASCAL}List)
		${NAME}Group.GET("/:id", middleware.PermissionRequired("${NAME}:view"), controllers.Get${PASCAL})
		${NAME}Group.POST("", middleware.PermissionRequired("${NAME}:create"), controllers.Create${PASCAL})
		${NAME}Group.PUT("/:id", middleware.PermissionRequired("${NAME}:edit"), controllers.Update${PASCAL})
		${NAME}Group.DELETE("/:id", middleware.PermissionRequired("${NAME}:delete"), controllers.Delete${PASCAL})
	}
EOF
)
# 在锚点行之前插入路由块（按行插入，跨平台可靠）
ROUTE_BLOCK="$ROUTE_BLOCK" python3 -c '
import os, sys
path = sys.argv[1]
block = os.environ["ROUTE_BLOCK"]
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
DB_FILE="$BACKEND_DIR/database/database.go"
if ! grep -q '【gen:migrate】' "$DB_FILE"; then
  echo "错误：database.go 缺少 【gen:migrate】 锚点" >&2
  exit 1
fi
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

# ---------- 生成前端 ----------
FRONTEND_VIEWS="$REPO_ROOT/frontend/src/views/${NAME}"
require_new "$FRONTEND_VIEWS/index.vue"
mkdir -p "$FRONTEND_VIEWS"
replace "$EXAMPLE_DIR/frontend/index.vue" "$FRONTEND_VIEWS/index.vue"
echo "  ✓ frontend/src/views/${NAME}/index.vue"

API_FILE="$REPO_ROOT/frontend/src/api/index.js"
API_BLOCK=$(cat <<EOF

// ==================== ${PASCAL} 管理 ====================
export const get${PASCAL}List = (params) => request.get('/${NAME}', { params })
export const get${PASCAL} = (id) => request.get(\`/${NAME}/\${id}\`)
export const create${PASCAL} = (data) => request.post('/${NAME}', data)
export const update${PASCAL} = (id, data) => request.put(\`/${NAME}/\${id}\`, data)
export const delete${PASCAL} = (id) => request.delete(\`/${NAME}/\${id}\`)
EOF
)
printf '%s\n' "$API_BLOCK" >> "$API_FILE"
echo "  ✓ frontend/src/api/index.js 已追加 API 定义"

# ---------- 完成提示 ----------
cat <<EOF

✅ 模块「${NAME}」骨架生成完毕。

后续手工步骤（生成器不自动处理，需 AI/开发者完成）：
1. 填充业务逻辑：搜索「// TODO: 业务逻辑」锚点，实现唯一性校验、级联、关联等。
2. 权限码落库：在 database.go 的 initBaseData 中为「${NAME}:view/create/edit/delete」新增 Permission 记录。
3. 前端路由/菜单：在 frontend/src/router/index.js 与 Layout.vue 的 menus 中新增该模块入口。
4. 路由路径：默认用单数 /${NAME}，如需复数请手动调整 router.go 与 api/index.js。
5. 验证：cd backend && go build ./... && make test && make smoke
EOF
