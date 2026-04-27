variable "name_prefix" {
  type = string
}

variable "region" {
  type = string
}

variable "dynamodb_table_name" {
  type = string
}

variable "dynamodb_table_arn" {
  type = string
}

variable "cognito_user_pool_id" {
  type = string
}

variable "cognito_user_pool_client_id" {
  type = string
}
