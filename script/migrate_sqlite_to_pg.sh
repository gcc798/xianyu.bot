#!/usr/bin/env bash
set -euo pipefail

sqlite_db="${1:-server/dbdata/xianyu_assistant.db}"
pg_url="${2:-postgresql://xianyu:xianyu-local-only@127.0.0.1:15432/xianyu}"

if ! command -v pgloader >/dev/null 2>&1; then
  echo "pgloader is required; install it before migrating" >&2
  exit 1
fi
if [[ ! -f "$sqlite_db" ]]; then
  echo "SQLite database not found: $sqlite_db" >&2
  exit 1
fi

mkdir -p tmp
snapshot="tmp/xianyu-assistant-$(date +%Y%m%d-%H%M%S).db"
sqlite3 "$sqlite_db" ".backup '$snapshot'"
pgloader "sqlite://$(cd "$(dirname "$snapshot")" && pwd)/$(basename "$snapshot")" "$pg_url"
echo "Migrated snapshot: $snapshot"
