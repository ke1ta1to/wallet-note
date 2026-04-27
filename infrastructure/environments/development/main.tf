data "aws_caller_identity" "current" {}

module "cognito" {
  source = "../../modules/cognito"

  name_prefix   = local.name_prefix
  domain_prefix = "${local.name_prefix}-${data.aws_caller_identity.current.account_id}"
  callback_urls = var.cognito_callback_urls
  logout_urls   = var.cognito_logout_urls
}

module "dynamodb" {
  source = "../../modules/dynamodb"

  name_prefix = local.name_prefix
}

module "web_client" {
  source = "../../modules/web-client"

  name_prefix = local.name_prefix
  account_id  = data.aws_caller_identity.current.account_id
}
