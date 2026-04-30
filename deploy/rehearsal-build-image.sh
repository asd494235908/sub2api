#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="${ENV_FILE:-$SCRIPT_DIR/.env.localtest}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing env file: $ENV_FILE" >&2
  echo "Copy $SCRIPT_DIR/.env.localtest.example to $ENV_FILE first." >&2
  exit 1
fi

set -a
source "$ENV_FILE"
set +a

IMAGE_TAG="${SUB2API_IMAGE:-sub2api-local:v0.1.120-rehearsal}"
BUILD_DIR="$SCRIPT_DIR/localtest-build"
TMP_CONTAINER="sub2api-rehearsal-image-tmp"
mkdir -p "$BUILD_DIR"

pushd "$REPO_ROOT/frontend" >/dev/null
npm_config_registry="${NPM_CONFIG_REGISTRY:-https://registry.npmmirror.com}" pnpm install --frozen-lockfile
pnpm run build
popd >/dev/null

pushd "$REPO_ROOT/backend" >/dev/null
GOPROXY="${GOPROXY:-https://goproxy.cn,direct}" go mod download
VERSION_VALUE="$(tr -d '\r\n' < ./cmd/server/VERSION)"
DATE_VALUE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY="${GOPROXY:-https://goproxy.cn,direct}" \
  go build -tags embed \
  -ldflags="-s -w -X main.Version=${VERSION_VALUE} -X main.Commit=local-rehearsal -X main.Date=${DATE_VALUE} -X main.BuildType=release" \
  -trimpath \
  -o "$BUILD_DIR/sub2api" \
  ./cmd/server
popd >/dev/null

docker rm -f "$TMP_CONTAINER" >/dev/null 2>&1 || true
docker create --platform linux/amd64 --name "$TMP_CONTAINER" weishaw/sub2api:latest >/dev/null
docker cp "$BUILD_DIR/sub2api" "$TMP_CONTAINER:/app/sub2api"
docker cp "$REPO_ROOT/backend/resources/." "$TMP_CONTAINER:/app/resources"
docker commit "$TMP_CONTAINER" "$IMAGE_TAG" >/dev/null
docker rm -f "$TMP_CONTAINER" >/dev/null

echo "Built image: $IMAGE_TAG"
