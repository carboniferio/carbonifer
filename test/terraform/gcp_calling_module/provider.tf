module "globals" {
  source = "../gcp_global_module"
}

provider "google" {
  region = module.globals.common_region
}
