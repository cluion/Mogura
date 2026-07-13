#!/usr/bin/env sh
# Mogura 一鍵安裝:抓取最新 release 的靜態 binary 放進 ~/.local/bin
# 用法: curl -fsSL https://raw.githubusercontent.com/cluion/Mogura/main/install.sh | sh
set -eu

REPO="${MOGURA_REPO:-cluion/Mogura}"
INSTALL_DIR="${MOGURA_INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -m)" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) echo "不支援的架構: $(uname -m)" >&2; exit 1 ;;
esac

if [ "$(uname -s)" != "Linux" ]; then
  echo "Mogura 只支援 Linux" >&2
  exit 1
fi

echo "🦡 查詢最新版本..."
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | head -1 | cut -d '"' -f 4)
[ -n "$TAG" ] || { echo "查不到最新版本,請確認 ${REPO} 已有 release" >&2; exit 1; }

VERSION="${TAG#v}"
URL="https://github.com/${REPO}/releases/download/${TAG}/mogura_${VERSION}_linux_${ARCH}.tar.gz"

echo "🦡 下載 ${TAG}(linux/${ARCH})..."
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
curl -fsSL "$URL" -o "$TMP/mogura.tar.gz"
tar -xzf "$TMP/mogura.tar.gz" -C "$TMP" mogura

mkdir -p "$INSTALL_DIR"
install -m 0755 "$TMP/mogura" "$INSTALL_DIR/mogura"

echo "✨ 已安裝到 $INSTALL_DIR/mogura"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "提醒:$INSTALL_DIR 不在 PATH,請加入 shell 設定檔" ;;
esac
"$INSTALL_DIR/mogura" version
