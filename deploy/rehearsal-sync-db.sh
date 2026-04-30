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
LOCAL_DB_CONTAINER="${LOCAL_DB_CONTAINER:-sub2api-rehearsal-postgres}"

mkdir -p "$SCRIPT_DIR/localtest-backups"
STAMP="$(date +%Y%m%d-%H%M%S)"
REMOTE_DUMP_PATH="/tmp/sub2api-localtest-${STAMP}.dump"
LOCAL_DUMP_PATH="$SCRIPT_DIR/localtest-backups/sub2api-${STAMP}.dump"

echo "Creating remote dump at $REMOTE_DUMP_PATH ..."
expect <<EOF
set timeout -1
spawn ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ${REMOTE_SSH_USER}@${REMOTE_SSH_HOST} "docker exec -e PGPASSWORD=${REMOTE_DB_PASSWORD} ${REMOTE_DB_CONTAINER} pg_dump -Fc -U ${REMOTE_DB_USER} -d ${REMOTE_DB_NAME} > ${REMOTE_DUMP_PATH}"
expect {
  "*yes/no*" { send "yes\r"; exp_continue }
  "*password:*" { send "${REMOTE_SSH_PASSWORD}\r" }
}
expect eof
EOF

echo "Downloading dump to $LOCAL_DUMP_PATH ..."
expect <<EOF
set timeout -1
spawn scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ${REMOTE_SSH_USER}@${REMOTE_SSH_HOST}:${REMOTE_DUMP_PATH} ${LOCAL_DUMP_PATH}
expect {
  "*yes/no*" { send "yes\r"; exp_continue }
  "*password:*" { send "${REMOTE_SSH_PASSWORD}\r" }
}
expect eof
EOF

if [[ ! -s "$LOCAL_DUMP_PATH" ]]; then
  echo "Downloaded dump is missing or empty: $LOCAL_DUMP_PATH" >&2
  exit 1
fi

echo "Removing remote dump ..."
expect <<EOF
set timeout -1
spawn ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ${REMOTE_SSH_USER}@${REMOTE_SSH_HOST} "rm -f ${REMOTE_DUMP_PATH}"
expect {
  "*yes/no*" { send "yes\r"; exp_continue }
  "*password:*" { send "${REMOTE_SSH_PASSWORD}\r" }
}
expect eof
EOF

echo "Recreating local database in ${LOCAL_DB_CONTAINER} ..."
docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" "$LOCAL_DB_CONTAINER" psql -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${POSTGRES_DB}' AND pid <> pg_backend_pid();"
docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" "$LOCAL_DB_CONTAINER" psql -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${POSTGRES_DB};"
docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" "$LOCAL_DB_CONTAINER" psql -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE ${POSTGRES_DB};"

echo "Restoring dump into local database ..."
docker exec -i -e PGPASSWORD="${POSTGRES_PASSWORD}" "$LOCAL_DB_CONTAINER" pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists --no-owner --no-privileges < "$LOCAL_DUMP_PATH"

echo "Local schema_migrations summary:"
docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" "$LOCAL_DB_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "SELECT COUNT(*), max(filename) FROM schema_migrations;"

echo "Database sync completed: $LOCAL_DUMP_PATH"
