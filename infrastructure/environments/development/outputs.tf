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

output "dynamodb_table_name" {
  value = module.dynamodb.table_name
}

output "dynamodb_table_arn" {
  value = module.dynamodb.table_arn
}

output "web_client_bucket_name" {
  value = module.web_client.bucket_name
}

output "web_client_distribution_id" {
  value = module.web_client.distribution_id
}

output "web_client_distribution_domain" {
  value = module.web_client.distribution_domain
}

output "web_client_url" {
  value = "https://${module.web_client.distribution_domain}"
}
