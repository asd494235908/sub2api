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

mkdir -p "$SCRIPT_DIR/localtest-data/postgres" "$SCRIPT_DIR/localtest-data/redis" "$SCRIPT_DIR/localtest-backups"

docker compose --env-file "$ENV_FILE" -f "$SCRIPT_DIR/docker-compose.rehearsal-services.yml" up -d

for container in sub2api-rehearsal-postgres sub2api-rehearsal-redis; do
  echo "Waiting for $container to become healthy..."
  until [[ "$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container" 2>/dev/null)" == "healthy" ]]; do
    sleep 2
  done
done

docker compose --env-file "$ENV_FILE" -f "$SCRIPT_DIR/docker-compose.standalone.yml" up -d

echo "Rehearsal stack started."
echo "App:   http://127.0.0.1:${SERVER_PORT:-18080}"
echo "PG:    127.0.0.1:${LOCALTEST_POSTGRES_PORT:-15432}"
echo "Redis: 127.0.0.1:${LOCALTEST_REDIS_PORT:-16379}"
