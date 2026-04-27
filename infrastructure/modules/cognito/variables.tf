variable "name_prefix" {
  type = string
}

variable "domain_prefix" {
  type        = string
  description = "Cognito Hosted Domain prefix. Globally unique across AWS."
}

variable "callback_urls" {
  type = list(string)
}

variable "logout_urls" {
  type = list(string)
}

variable "allow_admin_password_auth" {
  type        = bool
  default     = false
  description = "Enable ALLOW_ADMIN_USER_PASSWORD_AUTH. Use only in development for CLI-driven token retrieval; never in production."
}
