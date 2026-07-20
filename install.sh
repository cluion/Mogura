#!/usr/bin/env sh
# Mogura one-line installer: fetch the latest static binary into ~/.local/bin
# Usage: curl -fsSL https://raw.githubusercontent.com/cluion/Mogura/main/install.sh | sh
set -eu

REPO="${MOGURA_REPO:-cluion/Mogura}"
INSTALL_DIR="${MOGURA_INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -m)" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

if [ "$(uname -s)" != "Linux" ]; then
  echo "Mogura only supports Linux" >&2
  exit 1
fi

echo "⛏️ Resolving latest release..."
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | head -1 | cut -d '"' -f 4)
[ -n "$TAG" ] || { echo "Could not resolve the latest release of ${REPO}" >&2; exit 1; }

VERSION="${TAG#v}"
URL="https://github.com/${REPO}/releases/download/${TAG}/mogura_${VERSION}_linux_${ARCH}.tar.gz"

echo "⛏️ Downloading ${TAG} (linux/${ARCH})..."
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
curl -fsSL "$URL" -o "$TMP/mogura.tar.gz"
tar -xzf "$TMP/mogura.tar.gz" -C "$TMP" mogura

mkdir -p "$INSTALL_DIR"
install -m 0755 "$TMP/mogura" "$INSTALL_DIR/mogura"

echo "✨ Installed to $INSTALL_DIR/mogura"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Note: $INSTALL_DIR is not in PATH; add it to your shell profile" ;;
esac
"$INSTALL_DIR/mogura" version
