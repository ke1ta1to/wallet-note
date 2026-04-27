#!/usr/bin/env bash
set -euo pipefail

FUNCTION_NAME="wallet-note-development-web-server"

REPO_ROOT=$(cd "$(dirname "$0")/.." && pwd)
BUILD_DIR="${REPO_ROOT}/web-server/build"
BINARY="${BUILD_DIR}/bootstrap"
ZIP="${BUILD_DIR}/wallet-note.zip"

mkdir -p "${BUILD_DIR}"

echo "==> Building Go binary (linux/arm64)..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
  go -C "${REPO_ROOT}/web-server" build -o "${BINARY}" ./cmd/wallet-note

echo "==> Packaging zip..."
rm -f "${ZIP}"
(cd "${BUILD_DIR}" && zip -q "${ZIP}" bootstrap)

echo "==> Updating Lambda function code: ${FUNCTION_NAME}..."
aws lambda update-function-code \
  --function-name "${FUNCTION_NAME}" \
  --zip-file "fileb://${ZIP}" \
  --no-cli-pager > /dev/null

echo "==> Waiting for Lambda update to complete..."
aws lambda wait function-updated-v2 --function-name "${FUNCTION_NAME}"

echo "==> Done."
