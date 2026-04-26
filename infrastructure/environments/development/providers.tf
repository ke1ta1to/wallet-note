provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project     = "wallet-note"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}
