resource "google_container_cluster" "auto_provisioned" {
  name = "${var.project_id}-gke"
  project = var.project_id
  remove_default_node_pool = true
  initial_node_count       = 1
  location = var.region

  network    = module.network.vpc_name
  subnetwork = module.network.subnet_name

  cluster_autoscaling {
    resource_limits {
      resource_type = "cpu"
      minimum = 1
      maximum = 10
    }
    resource_limits {
      resource_type = "memory"
      minimum = 4
      maximum = 16
    }
    resource_limits {
      resource_type = "nvidia-tesla-k80"
      minimum = 2
      maximum = 4
    }
    auto_provisioning_defaults {
      disk_size = 300
      disk_type = "pd-ssd"
    }
  }
}
