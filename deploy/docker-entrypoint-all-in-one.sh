#!/bin/sh
set -eu

DATA_ROOT="${DATA_ROOT:-/app/data}"
SUB2API_DATA_DIR="${DATA_DIR:-$DATA_ROOT/sub2api}"
POSTGRES_DATA_DIR="${POSTGRES_DATA_DIR:-$DATA_ROOT/postgres}"
REDIS_DATA_DIR="${REDIS_DATA_DIR:-$DATA_ROOT/redis}"
POSTGRES_USER_VALUE="${POSTGRES_USER:-sub2api}"
POSTGRES_PASSWORD_VALUE="${POSTGRES_PASSWORD:-sub2api_local_pass}"
POSTGRES_DB_VALUE="${POSTGRES_DB:-sub2api}"
REDIS_PASSWORD_VALUE="${REDIS_PASSWORD:-}"

export DATA_DIR="$SUB2API_DATA_DIR"
export PGDATA="$POSTGRES_DATA_DIR"
export POSTGRES_USER="$POSTGRES_USER_VALUE"
export POSTGRES_PASSWORD="$POSTGRES_PASSWORD_VALUE"
export POSTGRES_DB="$POSTGRES_DB_VALUE"
export DATABASE_HOST="${DATABASE_HOST:-127.0.0.1}"
export DATABASE_PORT="${DATABASE_PORT:-5432}"
export DATABASE_USER="${DATABASE_USER:-$POSTGRES_USER_VALUE}"
export DATABASE_PASSWORD="${DATABASE_PASSWORD:-$POSTGRES_PASSWORD_VALUE}"
export DATABASE_DBNAME="${DATABASE_DBNAME:-$POSTGRES_DB_VALUE}"
export DATABASE_SSLMODE="${DATABASE_SSLMODE:-disable}"
export REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
export REDIS_PORT="${REDIS_PORT:-6379}"
export REDIS_PASSWORD="$REDIS_PASSWORD_VALUE"
export REDIS_DB="${REDIS_DB:-0}"
export AUTO_SETUP="${AUTO_SETUP:-true}"
export SERVER_HOST="${SERVER_HOST:-0.0.0.0}"
export SERVER_PORT="${SERVER_PORT:-8080}"

mkdir -p "$SUB2API_DATA_DIR" "$POSTGRES_DATA_DIR" "$REDIS_DATA_DIR"
chown -R sub2api:sub2api "$SUB2API_DATA_DIR" "$POSTGRES_DATA_DIR" "$REDIS_DATA_DIR" 2>/dev/null || true

cleanup() {
    if [ "${SUB2API_PID:-}" ]; then
        kill "$SUB2API_PID" 2>/dev/null || true
    fi
    if [ "${REDIS_PID:-}" ]; then
        kill "$REDIS_PID" 2>/dev/null || true
    fi
    if [ -s "$POSTGRES_DATA_DIR/postmaster.pid" ]; then
        su-exec sub2api pg_ctl -D "$POSTGRES_DATA_DIR" -m fast -w stop >/dev/null 2>&1 || true
    fi
}

trap cleanup INT TERM EXIT

if [ ! -s "$POSTGRES_DATA_DIR/PG_VERSION" ]; then
    PWFILE="$(mktemp)"
    printf '%s\n' "$POSTGRES_PASSWORD_VALUE" > "$PWFILE"
    chown sub2api:sub2api "$PWFILE"
    su-exec sub2api initdb -D "$POSTGRES_DATA_DIR" --username="$POSTGRES_USER_VALUE" --pwfile="$PWFILE"
    rm -f "$PWFILE"
fi

su-exec sub2api pg_ctl -D "$POSTGRES_DATA_DIR" -o "-c listen_addresses='127.0.0.1' -c unix_socket_directories='$POSTGRES_DATA_DIR' -p $DATABASE_PORT" -w start

export PGPASSWORD="$POSTGRES_PASSWORD_VALUE"
if ! su-exec sub2api psql -h 127.0.0.1 -p "$DATABASE_PORT" -U "$POSTGRES_USER_VALUE" -d postgres -lqt | cut -d '|' -f 1 | sed 's/^ *//;s/ *$//' | grep -Fxq "$POSTGRES_DB_VALUE"; then
    su-exec sub2api createdb -h 127.0.0.1 -p "$DATABASE_PORT" -U "$POSTGRES_USER_VALUE" "$POSTGRES_DB_VALUE"
fi

REDIS_CONFIG="$REDIS_DATA_DIR/redis.conf"
{
    printf 'dir %s\n' "$REDIS_DATA_DIR"
    printf 'save 60 1\n'
    printf 'appendonly yes\n'
    printf 'appendfsync everysec\n'
    printf 'bind 127.0.0.1\n'
    printf 'port %s\n' "$REDIS_PORT"
    if [ -n "$REDIS_PASSWORD_VALUE" ]; then
        printf 'requirepass %s\n' "$REDIS_PASSWORD_VALUE"
    fi
} > "$REDIS_CONFIG"
chown sub2api:sub2api "$REDIS_CONFIG" 2>/dev/null || true

if [ -n "$REDIS_PASSWORD_VALUE" ]; then
    export REDISCLI_AUTH="$REDIS_PASSWORD_VALUE"
fi
su-exec sub2api redis-server "$REDIS_CONFIG" &
REDIS_PID="$!"

for _ in $(seq 1 30); do
    if [ -n "$REDIS_PASSWORD_VALUE" ]; then
        if redis-cli -h 127.0.0.1 -p "$REDIS_PORT" ping 2>/dev/null | grep -q PONG; then
            break
        fi
    else
        if redis-cli -h 127.0.0.1 -p "$REDIS_PORT" ping 2>/dev/null | grep -q PONG; then
            break
        fi
    fi
    sleep 1
done

/app/sub2api &
SUB2API_PID="$!"
wait "$SUB2API_PID"
