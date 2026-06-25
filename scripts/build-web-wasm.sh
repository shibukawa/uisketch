#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
GOROOT_DIR=$(go env GOROOT)

cp "$GOROOT_DIR/lib/wasm/wasm_exec.js" "$ROOT_DIR/web/wasm_exec.js"
GOOS=js GOARCH=wasm go build -o "$ROOT_DIR/web/uisketch.wasm" ./cmd/uisketch-wasm

printf 'built web/uisketch.wasm and web/wasm_exec.js\n'
