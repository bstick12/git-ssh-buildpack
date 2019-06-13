#!/usr/bin/env bash
set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

TARGET_OS=${1:-linux}

for b in cmd/*; do
  [[ -e "$b" ]] || break
  b=$(basename "$b")
  echo -n "Building $b..."
  GOOS=$TARGET_OS go build -ldflags="-s -w" -o "bin/$b" "cmd/$b/main.go"
  echo "done"
done
