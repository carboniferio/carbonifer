resource "google_container_cluster" "my_cluster" {
  name = "${var.project_id}-gke"
  project = var.project_id
  remove_default_node_pool = true
  initial_node_count       = 1

  network    = module.network.vpc_name
  subnetwork = module.network.subnet_name
}

resource "google_container_node_pool" "gke_nodes" {
  name       = google_container_cluster.my_cluster.name
  cluster    = google_container_cluster.my_cluster.id
  node_count = var.num_nodes
  node_locations = [ "${var.region}-a", "${var.region}-b", "${var.region}-c" ]

  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    labels = {
      env = var.project_id
    }

    preemptible  = true
    machine_type = "n1-standard-2"
    disk_size_gb = 100
    disk_type = "pd-ssd"
    ephemeral_storage_local_ssd_config {
      local_ssd_count = 4
    }
    local_nvme_ssd_block_config {
      local_ssd_count = 2
    }
    local_ssd_count = 1
  }
}


resource "google_container_cluster" "my_cluster_no_pool" {
  name = "${var.project_id}-gke-2"
  project = var.project_id
  node_locations = [ "${var.region}-a", "${var.region}-b", "${var.region}-c" ]
  remove_default_node_pool = true
  initial_node_count       = 4
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

  network    = module.network.vpc_name
  subnetwork = module.network.subnet_name
}