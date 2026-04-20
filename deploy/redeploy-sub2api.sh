#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

DEFAULT_REMOTE_HOST="101.96.208.132"
DEFAULT_REMOTE_USER="root"
DEFAULT_REMOTE_DEPLOY_DIR="/opt/sub2api-current"
DEFAULT_IMAGE_TAG="weishaw/sub2api:latest"
DEFAULT_REMOTE_ENV_FILE="${DEFAULT_REMOTE_DEPLOY_DIR}/deploy/.env"
DEFAULT_REMOTE_DATA_DIR="${DEFAULT_REMOTE_DEPLOY_DIR}/deploy/data"
DEFAULT_ALPINE_IMAGE="alpine:3.21"
DEFAULT_NODE_IMAGE="node:24-alpine"
DEFAULT_GOLANG_IMAGE="golang:1.26.1-alpine"
DEFAULT_POSTGRES_IMAGE="postgres:18-alpine"

REMOTE_HOST="${DEFAULT_REMOTE_HOST}"
REMOTE_USER="${DEFAULT_REMOTE_USER}"
REMOTE_DEPLOY_DIR="${DEFAULT_REMOTE_DEPLOY_DIR}"
IMAGE_TAG="${DEFAULT_IMAGE_TAG}"
DRY_RUN="false"
ALPINE_IMAGE="${ALPINE_IMAGE:-${DEFAULT_ALPINE_IMAGE}}"
NODE_IMAGE="${NODE_IMAGE:-${DEFAULT_NODE_IMAGE}}"
GOLANG_IMAGE="${GOLANG_IMAGE:-${DEFAULT_GOLANG_IMAGE}}"
POSTGRES_IMAGE="${POSTGRES_IMAGE:-${DEFAULT_POSTGRES_IMAGE}}"

TMP_DIR="$(mktemp -d /tmp/sub2api-redeploy.XXXXXX)"
CONTROL_SOCKET="${TMP_DIR}/ssh-control.sock"
IMAGE_ARCHIVE="${TMP_DIR}/sub2api-image.tar.gz"

cleanup() {
  if [[ -S "${CONTROL_SOCKET}" ]]; then
    ssh -S "${CONTROL_SOCKET}" -O exit "${REMOTE_USER}@${REMOTE_HOST}" >/dev/null 2>&1 || true
  fi
  rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

usage() {
  cat <<'EOF'
Usage: deploy/redeploy-sub2api.sh [options]

Rebuild local code and redeploy only the remote sub2api container.
sub2api-postgres and sub2api-redis are left untouched.

Options:
  --host <host>         Remote host (default: 101.96.208.132)
  --user <user>         Remote user (default: root)
  --deploy-dir <path>   Remote deploy root (default: /opt/sub2api-current)
  --image <tag>         Local/remote image tag (default: weishaw/sub2api:latest)
  --dry-run             Print commands without executing them
  -h, --help            Show this help

Optional environment overrides for the build stage:
  ALPINE_IMAGE          Runtime base image (default: alpine:3.21)
  NODE_IMAGE            Frontend builder image (default: node:24-alpine)
  GOLANG_IMAGE          Backend builder image (default: golang:1.26.1-alpine)
  POSTGRES_IMAGE        pg client image (default: postgres:18-alpine)
EOF
}

run() {
  if [[ "${DRY_RUN}" == "true" ]]; then
    printf '[dry-run] %s\n' "$*"
  else
    "$@"
  fi
}

run_remote() {
  local remote_cmd="$1"
  if [[ "${DRY_RUN}" == "true" ]]; then
    printf '[dry-run] ssh remote: %s\n' "${remote_cmd}"
  else
    ssh \
      -S "${CONTROL_SOCKET}" \
      -o ControlMaster=auto \
      -o ControlPath="${CONTROL_SOCKET}" \
      -o ControlPersist=10m \
      -o StrictHostKeyChecking=no \
      "${REMOTE_USER}@${REMOTE_HOST}" \
      "${remote_cmd}"
  fi
}

require_cmd() {
  local cmd="$1"
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    printf 'Missing required command: %s\n' "${cmd}" >&2
    exit 1
  fi
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host)
      REMOTE_HOST="$2"
      shift 2
      ;;
    --user)
      REMOTE_USER="$2"
      shift 2
      ;;
    --deploy-dir)
      REMOTE_DEPLOY_DIR="$2"
      shift 2
      ;;
    --image)
      IMAGE_TAG="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      printf 'Unknown option: %s\n\n' "$1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

REMOTE_ENV_FILE="${REMOTE_DEPLOY_DIR}/deploy/.env"
REMOTE_DATA_DIR="${REMOTE_DEPLOY_DIR}/deploy/data"

require_cmd docker
require_cmd ssh
require_cmd scp
require_cmd gzip

printf 'Target: %s@%s\n' "${REMOTE_USER}" "${REMOTE_HOST}"
printf 'Deploy dir: %s\n' "${REMOTE_DEPLOY_DIR}"
printf 'Image tag: %s\n' "${IMAGE_TAG}"
printf 'Build images: alpine=%s node=%s golang=%s postgres=%s\n' \
  "${ALPINE_IMAGE}" "${NODE_IMAGE}" "${GOLANG_IMAGE}" "${POSTGRES_IMAGE}"

run docker buildx build \
  --platform linux/amd64 \
  --build-arg "ALPINE_IMAGE=${ALPINE_IMAGE}" \
  --build-arg "NODE_IMAGE=${NODE_IMAGE}" \
  --build-arg "GOLANG_IMAGE=${GOLANG_IMAGE}" \
  --build-arg "POSTGRES_IMAGE=${POSTGRES_IMAGE}" \
  -t "${IMAGE_TAG}" \
  --load \
  "${REPO_ROOT}"
if [[ "${DRY_RUN}" == "true" ]]; then
  printf '[dry-run] docker save --platform linux/amd64 %s | gzip > %s\n' "${IMAGE_TAG}" "${IMAGE_ARCHIVE}"
else
  docker save --platform linux/amd64 "${IMAGE_TAG}" | gzip > "${IMAGE_ARCHIVE}"
fi

if [[ "${DRY_RUN}" == "true" ]]; then
  printf '[dry-run] ssh master connect %s@%s\n' "${REMOTE_USER}" "${REMOTE_HOST}"
else
  ssh \
    -M \
    -S "${CONTROL_SOCKET}" \
    -o ControlMaster=yes \
    -o ControlPath="${CONTROL_SOCKET}" \
    -o ControlPersist=10m \
    -o StrictHostKeyChecking=no \
    -fnNT \
    "${REMOTE_USER}@${REMOTE_HOST}"
fi

run_remote "mkdir -p '${REMOTE_DEPLOY_DIR}'"

if [[ "${DRY_RUN}" == "true" ]]; then
  printf '[dry-run] scp %s -> %s@%s:%s/\n' "${IMAGE_ARCHIVE}" "${REMOTE_USER}" "${REMOTE_HOST}" "${REMOTE_DEPLOY_DIR}"
else
  scp \
    -S "$(command -v ssh)" \
    -o ControlMaster=auto \
    -o ControlPath="${CONTROL_SOCKET}" \
    -o ControlPersist=10m \
    -o StrictHostKeyChecking=no \
    "${IMAGE_ARCHIVE}" \
    "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DEPLOY_DIR}/"
fi

read -r -d '' REMOTE_SCRIPT <<EOF || true
set -euo pipefail
cd '${REMOTE_DEPLOY_DIR}'
gunzip -c sub2api-image.tar.gz | docker load
docker rm -f sub2api >/dev/null 2>&1 || true
sleep 2
docker run -d \
  --name sub2api \
  --restart unless-stopped \
  --network host \
  --ulimit nofile=100000:100000 \
  --add-host sub2api-postgres:127.0.0.1 \
  --add-host sub2api-redis:127.0.0.1 \
  --env-file '${REMOTE_ENV_FILE}' \
  -v '${REMOTE_DATA_DIR}:/app/data' \
  '${IMAGE_TAG}'

for _ in \$(seq 1 40); do
  if [ "\$(docker inspect --format '{{.State.Health.Status}}' sub2api 2>/dev/null || true)" = "healthy" ]; then
    break
  fi
  sleep 3
done

STATUS="\$(docker inspect --format '{{.State.Status}} {{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}} {{.RestartCount}}' sub2api)"
if [ "\$(docker inspect --format '{{.State.Health.Status}}' sub2api 2>/dev/null || true)" != "healthy" ]; then
  echo "sub2api failed to become healthy: \${STATUS}" >&2
  docker logs --tail 80 sub2api >&2 || true
  exit 1
fi

docker image inspect '${IMAGE_TAG}' --format '{{.Id}} {{.Architecture}}/{{.Os}}'
docker inspect sub2api --format 'status={{.State.Status}} health={{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}} restart={{.RestartCount}}'
curl -sS http://127.0.0.1:8080/health
EOF

run_remote "${REMOTE_SCRIPT}"

printf 'Redeploy complete.\n'
