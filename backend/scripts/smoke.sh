#!/usr/bin/env bash
# 冒烟测试：一条命令证明后端「能启动、能登录、能鉴权、能响应」。
# 流程：构建 → 随机端口 + 临时 db 启动 → 登录拿 token → 命中受保护路由 → 断言 → 清理。
# 用于 AI 每改一次后端后快速验证，把「完成了」变成可观察的事实。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

TMP_DIR="$(mktemp -d /tmp/base_smoke_XXXXXX)"
DB_FILE="$TMP_DIR/base.db"
BIN_FILE="$TMP_DIR/base"
SERVER_LOG="$TMP_DIR/server.log"

PORT="$((20000 + RANDOM % 20000))"
BASE_URL="http://127.0.0.1:${PORT}"
SERVER_PID=""

cleanup() {
  if [[ -n "$SERVER_PID" ]]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

echo "==> 构建后端..."
(cd "$BACKEND_DIR" && go build -o "$BIN_FILE" .)

echo "==> 启动服务（端口 ${PORT}，临时 db ${DB_FILE}）..."
BASE_BACKEND_SERVER_PORT="$PORT" \
BASE_BACKEND_DB_DSN="$DB_FILE" \
BASE_BACKEND_GIN_MODE=release \
"$BIN_FILE" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

# 等待就绪（最多 ~15s）
ready=0
for _ in $(seq 1 30); do
  if curl -sf "$BASE_URL/api/system/info" >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 0.5
done
if [[ "$ready" != "1" ]]; then
  echo "❌ 服务启动超时" >&2
  cat "$SERVER_LOG" >&2
  exit 1
fi

echo "==> 登录获取 token..."
LOGIN_RESP="$(curl -sf -X POST "$BASE_URL/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}')"
TOKEN="$(printf '%s' "$LOGIN_RESP" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')"

if [[ -z "$TOKEN" ]]; then
  echo "❌ 登录失败，未获取到 token" >&2
  echo "$LOGIN_RESP" >&2
  exit 1
fi
echo "==> 登录成功"

echo "==> 命中受保护路由 /api/auth/info ..."
INFO_RESP="$(curl -sf "$BASE_URL/api/auth/info" -H "Authorization: Bearer $TOKEN")"
CODE="$(printf '%s' "$INFO_RESP" | sed -n 's/^{"code":\([0-9]*\).*/\1/p')"
if [[ "$CODE" != "200" ]]; then
  echo "❌ 受保护路由返回 code=${CODE}，期望 200" >&2
  echo "$INFO_RESP" >&2
  exit 1
fi

echo "==> 命中 mp-login 接口可达性（未配置 wechat 段，期望 code=500 + 明确指引）..."
MP_RESP="$(curl -sf -X POST "$BASE_URL/api/auth/mp-login" \
  -H 'Content-Type: application/json' \
  -d '{"code":"smoke-test"}')"
MP_CODE="$(printf '%s' "$MP_RESP" | sed -n 's/^{"code":\([0-9]*\).*/\1/p')"
if [[ "$MP_CODE" != "500" ]]; then
  echo "❌ mp-login 返回 code=${MP_CODE}，期望 500（未配置 wechat 段）" >&2
  echo "$MP_RESP" >&2
  exit 1
fi
if ! printf '%s' "$MP_RESP" | grep -q "未配置"; then
  echo "❌ mp-login 响应未含「未配置」指引：" >&2
  echo "$MP_RESP" >&2
  exit 1
fi

echo "✅ 冒烟通过：后端可启动、可登录、可鉴权、可响应、mp-login 接口可达"
