#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
GOROOT_DIR=$(go env GOROOT)
OUT_DIR="$ROOT_DIR/packages/web/public"
WASM_EXEC_JS="$GOROOT_DIR/lib/wasm/wasm_exec.js"

if [ ! -f "$WASM_EXEC_JS" ]; then
  WASM_EXEC_JS="$GOROOT_DIR/misc/wasm/wasm_exec.js"
fi

if [ ! -f "$WASM_EXEC_JS" ]; then
  printf 'wasm_exec.js was not found under GOROOT: %s\n' "$GOROOT_DIR" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"
cp "$WASM_EXEC_JS" "$OUT_DIR/wasm_exec.js"
GOOS=js GOARCH=wasm go build -o "$OUT_DIR/uisketch.wasm" ./cmd/uisketch-wasm

printf 'built packages/web/public/uisketch.wasm and packages/web/public/wasm_exec.js\n'
