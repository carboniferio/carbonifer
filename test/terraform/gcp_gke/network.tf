module "network" {
  source = "../gcp_global_module/network"
  project_id = var.project_id
  region = var.region
}
