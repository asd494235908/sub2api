#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${ENV_FILE:-$SCRIPT_DIR/.env.localtest}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing env file: $ENV_FILE" >&2
  echo "Copy $SCRIPT_DIR/.env.localtest.example to $ENV_FILE first." >&2
  exit 1
fi

set -a
source "$ENV_FILE"
set +a

require_var() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    echo "Missing required environment variable: $name" >&2
    exit 1
  fi
}

require_var REMOTE_SSH_HOST
require_var REMOTE_SSH_USER
require_var REMOTE_SSH_PASSWORD
require_var REMOTE_DB_PASSWORD
require_var POSTGRES_USER
require_var POSTGRES_PASSWORD
require_var POSTGRES_DB

REMOTE_DB_CONTAINER="${REMOTE_DB_CONTAINER:-sub2api-postgres}"
REMOTE_DB_NAME="${REMOTE_DB_NAME:-sub2api}"
REMOTE_DB_USER="${REMOTE_DB_USER:-sub2api}"
LOCAL_DB_CONTAINER="${LOCAL_DB_CONTAINER:-sub2api-postgres-dev}"

shell_quote() {
  local value="$1"
  printf "'%s'" "${value//\'/\'\"\'\"\'}"
}

tmp_askpass="$(mktemp "${TMPDIR:-/tmp}/sub2api-askpass.XXXXXX")"
cleanup() {
  rm -f "$tmp_askpass"
}
trap cleanup EXIT
cat >"$tmp_askpass" <<'EOF'
#!/bin/sh
printf '%s\n' "$REMOTE_SSH_PASSWORD"
EOF
chmod 700 "$tmp_askpass"

remote_env_cmd="REMOTE_DB_CONTAINER=$(shell_quote "$REMOTE_DB_CONTAINER") REMOTE_DB_NAME=$(shell_quote "$REMOTE_DB_NAME") REMOTE_DB_USER=$(shell_quote "$REMOTE_DB_USER") REMOTE_DB_PASSWORD=$(shell_quote "$REMOTE_DB_PASSWORD") bash -s"

ssh_remote_pg_dump() {
  SSH_ASKPASS="$tmp_askpass" SSH_ASKPASS_REQUIRE=force DISPLAY=none ssh \
    -T \
    -o PreferredAuthentications=password \
    -o PubkeyAuthentication=no \
    -o PasswordAuthentication=yes \
    -o NumberOfPasswordPrompts=1 \
    -o StrictHostKeyChecking=no \
    -o UserKnownHostsFile=/dev/null \
    "${REMOTE_SSH_USER}@${REMOTE_SSH_HOST}" \
    "$remote_env_cmd"
}

echo "Checking remote dump access ..."
ssh_remote_pg_dump <<'REMOTE_SCRIPT'
set -euo pipefail
docker exec -e PGPASSWORD="$REMOTE_DB_PASSWORD" "$REMOTE_DB_CONTAINER" \
  pg_dump -Fc -s -U "$REMOTE_DB_USER" -d "$REMOTE_DB_NAME" >/dev/null
REMOTE_SCRIPT

echo "Recreating local database in ${LOCAL_DB_CONTAINER} ..."
docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" "$LOCAL_DB_CONTAINER" psql -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${POSTGRES_DB}' AND pid <> pg_backend_pid();"
docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" "$LOCAL_DB_CONTAINER" psql -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${POSTGRES_DB} WITH (FORCE);"
docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" "$LOCAL_DB_CONTAINER" psql -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE ${POSTGRES_DB};"

echo "Streaming remote dump into local database ..."
ssh_remote_pg_dump <<'REMOTE_SCRIPT' \
  | python3 -c 'import base64, sys; sys.stdout.buffer.write(base64.b64decode(sys.stdin.buffer.read()))' \
  | docker exec -i -e PGPASSWORD="${POSTGRES_PASSWORD}" "$LOCAL_DB_CONTAINER" pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists --no-owner --no-privileges
set -euo pipefail
docker exec -e PGPASSWORD="$REMOTE_DB_PASSWORD" "$REMOTE_DB_CONTAINER" \
  pg_dump -Fc -U "$REMOTE_DB_USER" -d "$REMOTE_DB_NAME" \
  --exclude-table-data='usage_logs' \
  --exclude-table-data='billing_usage_entries' \
  --exclude-table-data='usage_cleanup_tasks' \
  --exclude-table-data='scheduler_outbox' \
  --exclude-table-data='usage_billing_dedup' \
  --exclude-table-data='usage_billing_dedup_archive' \
  --exclude-table-data='usage_dashboard_daily' \
  --exclude-table-data='usage_dashboard_hourly' \
  --exclude-table-data='ops_*' \
  --exclude-table-data='*_logs' \
  --exclude-table-data='*_audit*' \
  --exclude-table-data='*_metrics_*' \
  --exclude-table-data='*_watermark' \
  --exclude-table-data='channel_monitor_histories' \
  --exclude-table-data='channel_monitor_daily_rollups' \
  | base64 | tr -d '\n'
REMOTE_SCRIPT

echo "Local schema_migrations summary:"
docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" "$LOCAL_DB_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "SELECT COUNT(*), max(filename) FROM schema_migrations;"

echo "Database sync completed"
