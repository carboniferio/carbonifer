
resource "google_container_cluster" "my_cluster_sub_pool" {
  name = "${var.project_id}-gke-2"
  project = var.project_id
  node_locations = [ "${var.region}-a", "${var.region}-b", "${var.region}-c" ]

  node_pool {
    name = "my-sub-pool"
    initial_node_count = 4
    node_config {
      oauth_scopes = [
        "https://www.googleapis.com/auth/cloud-platform"
      ]

      labels = {
        env = var.project_id
      }

      preemptible  = true
      machine_type = "n1-standard-2"
      disk_size_gb = 200
      disk_type = "pd-ssd"
      ephemeral_storage_local_ssd_config {
        local_ssd_count = 2
      }
    }
  }

  network    = module.network.vpc_name
  subnetwork = module.network.subnet_name
}