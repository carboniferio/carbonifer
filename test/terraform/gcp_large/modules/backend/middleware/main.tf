module "users_ms" {
  source = "../../compute"
  project_id = var.project_id
  region = var.region
  subnet_name = var.subnet_name
  instance_role = "users"
  instance_type = "n1-standard-2"
}

module "api_ms" {
  source = "../../compute"
  project_id = var.project_id
  region = var.region
  subnet_name = var.subnet_name
  instance_role = "api"
  instance_type = "a2-highgpu-1g"
}


