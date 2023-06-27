provider "google" {
  region = local.common_region
}

locals {
  common_region = "local_module_region"
}


output "common_region" {
  value = local.common_region
}