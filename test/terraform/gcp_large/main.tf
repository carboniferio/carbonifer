provider "google" {
  region      = var.region
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

