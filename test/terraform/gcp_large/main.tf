provider "google" {
  #project     = "carbonifer-sandbox"
  region      = var.region
  credentials = file("/Users/olivier/carbonifer/carbonifer-study/accounts/carbonifer-sandbox-c59717d85f65.json")
  //access_token = "foo"
}

module "network" {
  source = "./modules/network"
  project_id = var.project_id
  region = var.region
}

module "backend" {
  source = "./modules/backend"
  project_id = var.project_id
  region = var.region
  subnet_name = module.network.subnet_name
}

