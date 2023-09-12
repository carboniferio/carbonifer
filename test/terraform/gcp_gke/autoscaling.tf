resource "google_container_cluster" "my_cluster_autoscaled" {
  name = "${var.project_id}-gke"
  project = var.project_id
  remove_default_node_pool = true
  initial_node_count       = 1
  node_locations = [ "${var.region}-a", "${var.region}-b", "${var.region}-c" ]


  network    = module.network.vpc_name
  subnetwork = module.network.subnet_name
}

resource "google_container_node_pool" "gke_nodes_autoscaled" {
  name       = google_container_cluster.my_cluster_autoscaled.name
  cluster    = google_container_cluster.my_cluster_autoscaled.id

  autoscaling {
    min_node_count = 4
    max_node_count = 20
  }

  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    labels = {
      env = var.project_id
    }

    preemptible  = true
    machine_type = "n1-standard-2"
    disk_size_gb = 150
    disk_type = "pd-ssd"
  }
}

resource "google_container_cluster" "my_cluster_autoscaled_monozone" {
  name = "${var.project_id}-gke"
  project = var.project_id
  remove_default_node_pool = true
  initial_node_count       = 1
  location = "${var.region}-a"

  network    = module.network.vpc_name
  subnetwork = module.network.subnet_name
}

resource "google_container_node_pool" "gke_nodes_autoscaled_monozone" {
  name       = google_container_cluster.my_cluster_autoscaled_monozone.name
  cluster    = google_container_cluster.my_cluster_autoscaled_monozone.id

  autoscaling {
    min_node_count = 4
    max_node_count = 20
  }

  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    labels = {
      env = var.project_id
    }

    preemptible  = true
    machine_type = "n1-standard-2"
    disk_size_gb = 150
    disk_type = "pd-ssd"
  }
}

resource "google_container_cluster" "my_cluster_autoscaled_total" {
  name = "${var.project_id}-gke"
  project = var.project_id
  remove_default_node_pool = true
  initial_node_count       = 5
  node_locations = [ "${var.region}-a", "${var.region}-b", "${var.region}-c" ]


  network    = module.network.vpc_name
  subnetwork = module.network.subnet_name
}

resource "google_container_node_pool" "gke_nodes_autoscaled_total" {
  name       = google_container_cluster.my_cluster_autoscaled_total.name
  cluster    = google_container_cluster.my_cluster_autoscaled_total.id

  autoscaling {
    total_min_node_count = 40
    total_max_node_count = 100
  }

  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    labels = {
      env = var.project_id
    }

    preemptible  = true
    machine_type = "n1-standard-2"
    disk_size_gb = 150
    disk_type = "pd-ssd"
  }
}