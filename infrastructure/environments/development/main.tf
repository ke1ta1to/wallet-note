data "aws_caller_identity" "current" {}

module "cognito" {
  source = "../../modules/cognito"

  name_prefix   = local.name_prefix
  domain_prefix = "${local.name_prefix}-${data.aws_caller_identity.current.account_id}"
  callback_urls = [
    "https://${module.web_client.distribution_domain}/auth/callback",
    "http://localhost:5173/auth/callback",
  ]
  logout_urls = [
    "https://${module.web_client.distribution_domain}/",
    "http://localhost:5173/",
  ]
}

module "dynamodb" {
  source = "../../modules/dynamodb"

  name_prefix = local.name_prefix
}

module "web_client" {
  source = "../../modules/web-client"

  name_prefix   = local.name_prefix
  account_id    = data.aws_caller_identity.current.account_id
  api_gw_domain = trimprefix(module.web_server.api_endpoint, "https://")
}

module "web_server" {
  source = "../../modules/web-server"

  name_prefix                 = local.name_prefix
  region                      = var.region
  dynamodb_table_name         = module.dynamodb.table_name
  dynamodb_table_arn          = module.dynamodb.table_arn
  cognito_user_pool_id        = module.cognito.user_pool_id
  cognito_user_pool_client_id = module.cognito.user_pool_client_id
}
