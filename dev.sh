#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

echo "Building twin-commander..."
go build -o twin-commander .

echo "Starting twin-commander..."
exec ./twin-commander "$@"
