resource "google_sql_database_instance" "instance" {
  name             = "my-database-instance"
  region           = var.region
  project          = var.project_id
  database_version = "POSTGRES_14"
  settings {
    tier              = "db-g1-small"
    availability_type = "REGIONAL"
  }
}
  
