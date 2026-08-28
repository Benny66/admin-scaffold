#!/usr/bin/env bash
# ============================================
#  Base 脚手架 - 一键初始化脚本
# ============================================
#   把「复制基座 → 改包名/标识 → 重置密钥 → 清历史/残留」固化为一条命令，
#   取代 README 里的手工 5 步流程。
#
#   用法：
#     ./scripts/init.sh <项目名> [选项]
#       （或 make init name=<项目名> [module=...] [db_name=...] [issuer=...] [app_name=...]）
#
#   选项：
#     --module <go模块名>    Go 模块名（默认 = 项目名）
#     --db-name <名>         数据库名（默认不改）
#     --issuer <名>          JWT Issuer（默认 = 项目名）
#     --app-name <系统名称>   中文品牌名，替换「企业管理系统」残留（默认不改）
#     --port <端口>          后端端口，生成 frontend/.env 与 mobile/.env（默认不改）
#     --yes                 跳过删除确认
#
#   前置：guard 已从 go.mod 动态读模块名（见 change guard-read-module-name），
#         本脚本可放心替换 import 而不拆护栏。
# ============================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
  echo "用法: $0 <项目名> [--module <go模块名>] [--db-name <名>] [--issuer <名>] [--app-name <系统名称>] [--port <端口>] [--yes]"
  echo ""
  echo "  项目名     必填，用于二进制名 / 包名 / 环境变量前缀"
  echo "  --module   Go 模块名（默认 = 项目名）"
  echo "  --db-name  数据库名（默认不改）"
  echo "  --issuer   JWT Issuer（默认 = 项目名）"
  echo "  --app-name 系统名称（中文品牌名，替换「企业管理系统」等残留，默认不改）"
  echo "  --port     后端端口（生成 frontend/.env 与 mobile/.env，默认不改）"
  echo "  --yes      跳过删除确认"
  exit 1
}

NAME=""
MODULE=""
DB_NAME=""
ISSUER=""
APP_NAME=""
PORT=""
YES=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --module)   MODULE="${2:-}"; shift 2 ;;
    --db-name)  DB_NAME="${2:-}"; shift 2 ;;
    --issuer)   ISSUER="${2:-}"; shift 2 ;;
    --app-name) APP_NAME="${2:-}"; shift 2 ;;
    --port)     PORT="${2:-}"; shift 2 ;;
    --yes|-y)   YES=1; shift ;;
    -*)         echo "未知参数: $1" >&2; usage ;;
    *)
      if [[ -z "$NAME" ]]; then NAME="$1"; else echo "多余参数: $1" >&2; usage; fi
      shift
      ;;
  esac
done

[[ -z "$NAME" ]] && usage
MODULE="${MODULE:-$NAME}"
ISSUER="${ISSUER:-$NAME}"

# 环境变量前缀：大写 + 非字母数字转下划线 + 尾部补一个下划线
ENV_PREFIX="$(printf '%s' "$NAME" | tr '[:lower:]' '[:upper:]' | tr -c '[:alnum:]' '_' | sed 's/_*$//')_"

# ---------- 幂等检查 ----------
cd "$ROOT_DIR"
CUR_MODULE="$(grep -m1 '^module ' backend/go.mod | awk '{print $2}')"
if [[ "$CUR_MODULE" != "base-backend" ]]; then
  echo "[错误] 当前 go.mod 模块名为 '$CUR_MODULE'，不是 'base-backend'。" >&2
  echo "       该项目可能已初始化过，拒绝重复执行。" >&2
  exit 1
fi

# ---------- 生成随机密钥 ----------
if command -v openssl >/dev/null 2>&1; then
  SECRET="$(openssl rand -hex 32)"
else
  SECRET="$(head -c 64 /dev/urandom | tr -dc 'a-f0-9' | head -c 64)"
fi

echo "==> 初始化项目「${NAME}」"
echo "    Go 模块名:   $MODULE"
echo "    env 前缀:    ${ENV_PREFIX}"
echo "    JWT Issuer:  $ISSUER"
[[ -n "$DB_NAME" ]] && echo "    数据库名:    $DB_NAME"
echo ""

# ---------- 执行文本替换 ----------
NAME="$NAME" MODULE="$MODULE" DB_NAME="$DB_NAME" ISSUER="$ISSUER" APP_NAME="$APP_NAME" \
ENV_PREFIX="$ENV_PREFIX" SECRET="$SECRET" SELF_PATH="$SCRIPT_DIR/init.sh" python3 <<'PY'
import os, re

name = os.environ["NAME"]
module = os.environ["MODULE"]
db_name = os.environ["DB_NAME"]
issuer = os.environ["ISSUER"]
app_name = os.environ["APP_NAME"]
env_prefix = os.environ["ENV_PREFIX"]
secret = os.environ["SECRET"]
self_path = os.path.realpath(os.environ["SELF_PATH"])

root = os.getcwd()
SKIP_DIRS = {".git", "node_modules", "dist", "dist-mobile", "deploy", ".claude", "openspec"}
TEXT_EXTS = {".go", ".mod", ".yaml", ".yml", ".json", ".sh", ".md", ".bat", ".js", ".vue", ".ts"}

# 扫描纳入：凡文本文件都纳入替换，靠 old==new 跳过无关文件。
# 不再白名单枚举，避免漏掉 backend/scripts/smoke.sh、docs/*.md 等含 base-backend 的文件。
# 但必须排除脚本自身，否则会把幂等哨兵与替换逻辑里的 base-backend 一并替换，自我破坏。
targets = []
for dirpath, dirnames, filenames in os.walk(root):
    dirnames[:] = [d for d in dirnames if d not in SKIP_DIRS]
    for fn in filenames:
        p = os.path.join(dirpath, fn)
        if os.path.realpath(p) == self_path:
            continue
        if os.path.splitext(fn)[1].lower() in TEXT_EXTS:
            targets.append(p)

def transform(s):
    # 1) 密钥（最长的特定串，最先处理）
    s = s.replace("base-backend-secret-key-change-me", secret)
    # 2) 环境变量前缀
    s = s.replace("BASE_BACKEND_", env_prefix)
    # 3) module 声明（go.mod）
    s = re.sub(r'^module base-backend\b', f'module {module}', s, flags=re.M)
    # 4) import 路径（含 _example/ 模板）：base-backend/ → module/
    s = s.replace("base-backend/", module + "/")
    # 5) 复合名（先于裸名）：压缩包 / 前后端包名 / 根包名
    s = s.replace("base-backend-deploy", name + "-deploy")
    s = s.replace("base-backend-frontend", name + "-frontend")
    s = s.replace("base-backend-mobile", name + "-mobile")
    s = s.replace("base-scaffold", name)
    # 6) JWT Issuer（精确，保留原对齐）
    s = re.sub(r'(Issuer:\s*)"base-backend"', rf'\1"{issuer}"', s)
    # 7) 数据库名（可选）
    if db_name:
        s = s.replace("base_backend", db_name)
    # 8) 剩余裸 base-backend（二进制名 / start 脚本）
    s = s.replace("base-backend", name)
    # 9) 中文品牌名（可选）：只有显式传 --app-name 才替换，避免臆断。
    #    只替换「企业管理系统」这一明确的中文品牌残留，不碰 Base Admin 中性占位
    #    （中性占位由运行时 brand-config 覆盖，不在此改动）。
    if app_name:
        s = s.replace("企业管理系统", app_name)
    return s

changed = []
for p in targets:
    with open(p, encoding="utf-8") as f:
        old = f.read()
    new = transform(old)
    if new != old:
        with open(p, "w", encoding="utf-8") as f:
            f.write(new)
        changed.append(os.path.relpath(p, root))

print(f"    替换 {len(changed)} 个文件:")
for c in changed:
    print(f"      - {c}")
PY

# ---------- 清 OpenSpec 历史（保留 config.yaml 工作流引擎） ----------
echo ""
echo "==> 将清理基座历史与运行时残留:"
if [[ -d openspec/specs ]]; then
  echo "    openspec/specs/   （$(find openspec/specs -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ') 个基线 spec）"
fi
if [[ -d openspec/changes ]]; then
  echo "    openspec/changes/ （$(find openspec/changes -mindepth 1 -maxdepth 1 | wc -l | tr -d ' ') 项）"
fi
ls backend/*.db backend/*.db-shm backend/*.db-wal backend/config.yaml 2>/dev/null | sed 's/^/    /' || true

if [[ "$YES" != "1" ]]; then
  read -r -p "确认删除？[y/N] " ans
  [[ "$ans" == "y" || "$ans" == "Y" ]] || { echo "已取消"; exit 1; }
fi

rm -rf openspec/specs openspec/changes
rm -f backend/*.db backend/*.db-shm backend/*.db-wal backend/config.yaml

# ---------- 生成前端/移动端 .env（可选，传 --port 才生成） ----------
if [[ -n "$PORT" ]]; then
  for dir in frontend mobile; do
    cat > "$dir/.env" <<EOF
VITE_API_BASE=http://localhost:${PORT}
EOF
  done
  echo ""
  echo "==> 已生成 frontend/.env 与 mobile/.env（VITE_API_BASE=http://localhost:${PORT}）"
fi

# 重写 openspec/config.yaml：保留 schema，context 填成新项目
cat > openspec/config.yaml <<EOF
schema: spec-driven

context: |
  $NAME 项目
  技术栈: Go 1.21+ / Gin / GORM / SQLite | Vue3 / Element Plus / Pinia / Vite
  认证: JWT (HS256)
  响应格式: {code, message, data}
  编码语言: 中文

rules:
  proposal:
    - 保持与现有 spec 一致的术语和命名
EOF

echo ""
echo "============================================"
echo "  初始化完成！"
echo "============================================"
echo "  项目名:     $NAME"
echo "  Go 模块名:  $MODULE"
echo "  env 前缀:   ${ENV_PREFIX}"
echo ""
echo "  下一步："
echo "    1. cd backend && go mod tidy && go test ./..."
echo "    2. cd frontend && npm install"
echo "    3. 按需把 backend/config.example.yaml 复制为 config.yaml 并设置密钥/端口"
echo ""
