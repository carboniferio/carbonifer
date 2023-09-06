module "middleware" {
  source = "./middleware"
  project_id = var.project_id
  region = var.region
  subnet_name = var.subnet_name
}

module "db" {
  source = "./database"
  project_id = var.project_id
  region = var.region
}



