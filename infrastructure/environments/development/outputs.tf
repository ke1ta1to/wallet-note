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

output "web_server_lambda_function_name" {
  value = module.web_server.lambda_function_name
}

output "web_server_api_endpoint" {
  value = module.web_server.api_endpoint
}

output "test_user_email" {
  value = aws_cognito_user.test.username
}

output "test_user_password" {
  value     = random_password.test_user.result
  sensitive = true
}
