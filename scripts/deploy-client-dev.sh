#!/usr/bin/env bash
set -euo pipefail

BUCKET="wallet-note-development-web-client-075472845547"
DIST_ID="E1KXTPW691CU0X"

REPO_ROOT=$(cd "$(dirname "$0")/.." && pwd)
DIST_DIR="${REPO_ROOT}/web-client/dist"

echo "==> Building web-client..."
(cd "${REPO_ROOT}/web-client" && pnpm build)

echo "==> Syncing to s3://${BUCKET}/ ..."
aws s3 sync "${DIST_DIR}/" "s3://${BUCKET}/" --delete --no-cli-pager > /dev/null

echo "==> Creating CloudFront invalidation on ${DIST_ID}..."
aws cloudfront create-invalidation \
  --distribution-id "${DIST_ID}" \
  --paths "/*" \
  --no-cli-pager > /dev/null

echo "==> Done."
