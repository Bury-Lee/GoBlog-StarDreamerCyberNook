#!/usr/bin/env bash
# GoBlog (StarDreamerCyberNook) 多平台构建脚本
# 用法:
#   ./build.sh                 # 构建 windows/linux/macos (amd64) + 前端
#   BUILD_ARM64=1 ./build.sh   # 额外构建 linux/macos (arm64)
#   SKIP_FRONTEND=1 ./build.sh # 跳过前端构建
#   OUT_DIR=/path ./build.sh   # 自定义输出目录(默认 dist/)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/dist}"

if ! command -v go >/dev/null 2>&1; then
    echo "错误: 未找到 go 命令,请先安装 Go 并配置 PATH" >&2
    exit 1
fi

mkdir -p "$OUT_DIR"

LDFLAGS="-s -w"

build() {
    local goos="$1"
    local goarch="$2"
    local output="$3"
    echo "==> 构建 $goos/$goarch -> $output"
    (
        cd "$ROOT_DIR"
        CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
            go build -ldflags="$LDFLAGS" -trimpath -o "$OUT_DIR/$output" .
    )
}

build windows amd64 main_windows_amd64.exe
build linux   amd64 main_linux_amd64
build darwin  amd64 main_macos_amd64

if [ "${BUILD_ARM64:-0}" = "1" ]; then
    build linux  arm64 main_linux_arm64
    build darwin arm64 main_macos_arm64
fi

if [ "${SKIP_FRONTEND:-0}" != "1" ]; then
    echo "==> 构建前端"
    (
        cd "$ROOT_DIR/frontend"
        npm install --no-audit --no-fund
        npm run build
    )
    rm -rf "$OUT_DIR/static/assets" "$OUT_DIR/static/index.html" "$OUT_DIR/static/favicon.svg"
    mkdir -p "$OUT_DIR/static"
    cp -r "$ROOT_DIR/frontend/dist/." "$OUT_DIR/static/"
fi

mkdir -p "$OUT_DIR/init"
cp -f "$ROOT_DIR/init/ip2region.xdb" "$OUT_DIR/init/ip2region.xdb"

echo
echo "构建完成,输出目录: $OUT_DIR"
ls -lh "$OUT_DIR"
