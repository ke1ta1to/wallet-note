variable "name_prefix" {
  type = string
}

variable "account_id" {
  type = string
}

variable "api_gw_domain" {
  type        = string
  description = "API Gateway HTTP API domain name (no scheme). Used as the second CloudFront origin for /api/* behavior."
}
