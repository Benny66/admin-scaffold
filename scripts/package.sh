#!/usr/bin/env bash
# ============================================
#   Base 脚手架 - 多平台构建打包脚本
# ============================================
#   提炼自资产系统 build.sh 的纯工程能力，剥离了 license / 品牌替换 / 证书 / tutorial。
#   品牌由 config.yaml（brand-config 运行时机制）负责，本脚本不碰。
#
#   用法：
#     ./scripts/package.sh              编译当前平台（本地测试）
#     ./scripts/package.sh --linux       交叉编译 Linux amd64
#     ./scripts/package.sh --windows     交叉编译 Windows amd64
#
#   产物：deploy/ 目录 + base-backend-deploy-<platform>.tar.gz / .zip
#   目标机器无需 Go / Node.js / 任何运行时环境。
# ============================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DEPLOY_DIR="$ROOT_DIR/deploy"

# ---------- 参数解析 ----------
TARGET_OS=""
for arg in "$@"; do
    case "$arg" in
        --linux) TARGET_OS="linux" ;;
        --windows) TARGET_OS="windows" ;;
    esac
done

# ---------- 确定目标平台与产物名 ----------
if [ "$TARGET_OS" = "linux" ]; then
    GOOS="linux"; GOARCH="amd64"
    BINARY_NAME="base-backend"; ARCHIVE_EXT="tar.gz"
elif [ "$TARGET_OS" = "windows" ]; then
    GOOS="windows"; GOARCH="amd64"
    BINARY_NAME="base-backend.exe"; ARCHIVE_EXT="zip"
else
    GOOS="$(go env GOOS 2>/dev/null || uname -s | tr '[:upper:]' '[:lower:]')"
    GOARCH="$(go env GOARCH 2>/dev/null || uname -m)"
    BINARY_NAME="base-backend"; ARCHIVE_EXT="tar.gz"
    if [ "$GOOS" = "windows" ]; then
        BINARY_NAME="base-backend.exe"; ARCHIVE_EXT="zip"
    fi
fi
ARCHIVE_NAME="base-backend-deploy-${TARGET_OS:-local}"

# ---------- 环境检测 ----------
echo "[检测] 检查构建环境..."
command -v go >/dev/null 2>&1   || { echo "[错误] 未检测到 Go（>= 1.21）"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "[错误] 未检测到 Node.js（>= 18）"; exit 1; }
command -v npm >/dev/null 2>&1  || { echo "[错误] 未检测到 npm"; exit 1; }
if [ "$ARCHIVE_EXT" = "zip" ] && ! command -v zip >/dev/null 2>&1; then
    echo "[错误] 未检测到 zip 命令（Mac: brew install zip / Linux: apt install zip）"; exit 1
fi
echo "[通过] Go $(go version | awk '{print $3}')，Node $(node -v)"
echo ""

# ---------- 1. 编译后端 ----------
echo "[1/5] 编译后端（${GOOS}/${GOARCH}）..."
cd "$ROOT_DIR/backend"
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath \
    -ldflags="-s -w" \
    -o "$BINARY_NAME" .
echo "[完成] 后端编译成功: $BINARY_NAME"
cd "$ROOT_DIR"
echo ""

# ---------- 2. 编译前端 ----------
echo "[2/5] 构建前端..."
cd "$ROOT_DIR/frontend"
[ -d node_modules ] || { echo "  安装前端依赖..."; npm install; }
npm run build
echo "[完成] 前端构建完成"
cd "$ROOT_DIR"
echo ""

# ---------- 3. 编译移动端 ----------
echo "[3/5] 构建移动端..."
if [ -d "$ROOT_DIR/mobile" ]; then
    cd "$ROOT_DIR/mobile"
    [ -d node_modules ] || { echo "  安装移动端依赖..."; npm install; }
    npm run build
    echo "[完成] 移动端构建完成"
    cd "$ROOT_DIR"
else
    echo "  跳过（无 mobile/ 目录）"
fi
echo ""

# ---------- 4. 组装部署目录 ----------
echo "[4/5] 组装部署目录..."
rm -rf "$DEPLOY_DIR"
mkdir -p "$DEPLOY_DIR"

# 二进制
cp "$ROOT_DIR/backend/$BINARY_NAME" "$DEPLOY_DIR/"

# 前端产物 → dist；移动端产物 → dist-mobile（重命名，与后端托管路径对齐）
cp -r "$ROOT_DIR/frontend/dist" "$DEPLOY_DIR/dist"
if [ -d "$ROOT_DIR/mobile/dist" ]; then
    cp -r "$ROOT_DIR/mobile/dist" "$DEPLOY_DIR/dist-mobile"
fi

# 品牌静态资源 → static（后端 /static/ 托管：logo/favicon/登录背景图）
if [ -d "$ROOT_DIR/backend/static" ]; then
    cp -r "$ROOT_DIR/backend/static" "$DEPLOY_DIR/static"
fi

# config.yaml：从 example 复制，强制 release
cp "$ROOT_DIR/backend/config/config.example.yaml" "$DEPLOY_DIR/config.yaml"
sed -i.bak 's/mode: "debug"/mode: "release"/' "$DEPLOY_DIR/config.yaml" && rm -f "$DEPLOY_DIR/config.yaml.bak"

# 启动脚本
if [ "$GOOS" = "windows" ]; then
    cat > "$DEPLOY_DIR/start.bat" << 'BAT_EOF'
@echo off
chcp 65001 > nul
cd /d "%~dp0"
base-backend.exe
pause
BAT_EOF
else
    cat > "$DEPLOY_DIR/start.sh" << 'SH_EOF'
#!/usr/bin/env bash
APP_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$APP_DIR"
./base-backend
SH_EOF
    chmod +x "$DEPLOY_DIR/start.sh"
fi

echo "[完成] 部署目录组装完成"
echo ""

# ---------- 5. 打压缩包 ----------
echo "[5/5] 创建压缩包..."
ARCHIVE_FILE="$ROOT_DIR/${ARCHIVE_NAME}.${ARCHIVE_EXT}"
rm -f "$ARCHIVE_FILE"
if [ "$ARCHIVE_EXT" = "zip" ]; then
    (cd "$DEPLOY_DIR" && zip -r -q "$ARCHIVE_FILE" .)
else
    tar -czf "$ARCHIVE_FILE" -C "$DEPLOY_DIR" .
fi

echo ""
echo "============================================"
echo "  构建完成！"
echo "============================================"
echo "  目标平台: $GOOS/$GOARCH"
echo "  部署目录: $DEPLOY_DIR"
echo "  压缩包:   $ARCHIVE_FILE"
echo ""
echo "  部署目录内容："
echo "    config.yaml      - 配置文件（release 模式）"
echo "    $BINARY_NAME     - 后端服务（已编译，可独立运行）"
echo "    dist/            - 桌面端页面"
echo "    dist-mobile/     - 移动端 H5 页面"
echo "    static/          - 品牌静态资源（logo/favicon/登录背景图）"
echo "    启动脚本          - start.sh / start.bat"
echo ""
echo "  目标机器无需 Go / Node.js / 任何运行时环境。"
echo "  默认管理员: admin / admin123"
echo ""
