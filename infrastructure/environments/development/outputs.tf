output "cognito_user_pool_id" {
  value = module.cognito.user_pool_id
}

output "cognito_user_pool_client_id" {
  value = module.cognito.user_pool_client_id
}

output "cognito_domain" {
  value = module.cognito.domain
}

output "cognito_hosted_login_url" {
  value = "https://${module.cognito.domain}.auth.${var.region}.amazoncognito.com"
}
