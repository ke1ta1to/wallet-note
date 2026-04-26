variable "project" {
  type    = string
  default = "wallet-note"
}

variable "environment" {
  type    = string
  default = "development"
}

variable "region" {
  type    = string
  default = "ap-northeast-1"
}

variable "cognito_callback_urls" {
  type    = list(string)
  default = ["http://localhost:5173/auth/callback"]
}

variable "cognito_logout_urls" {
  type    = list(string)
  default = ["http://localhost:5173/"]
}
