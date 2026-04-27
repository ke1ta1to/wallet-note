#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../infrastructure/environments/development"

EMAIL=$(terraform output -raw test_user_email)
PASSWORD=$(terraform output -raw test_user_password)
USER_POOL_ID=$(terraform output -raw cognito_user_pool_id)
CLIENT_ID=$(terraform output -raw cognito_user_pool_client_id)

aws cognito-idp admin-initiate-auth \
  --user-pool-id "${USER_POOL_ID}" \
  --client-id "${CLIENT_ID}" \
  --auth-flow ADMIN_USER_PASSWORD_AUTH \
  --auth-parameters "USERNAME=${EMAIL},PASSWORD=${PASSWORD}" \
  --query 'AuthenticationResult.AccessToken' \
  --output text
