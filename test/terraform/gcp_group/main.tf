resource "google_compute_network" "vpc_network" {
  name                    = "cbf-network"
  auto_create_subnetworks = false
  mtu                     = 1460
}

resource "google_compute_subnetwork" "first" {
  name          = "cbf-subnet"
  ip_cidr_range = "10.0.1.0/24"
  region        = "europe-west9"
  network       = google_compute_network.vpc_network.id
}

resource "google_compute_instance_template" "my-instance-template" {
  name             = "my-instance-template"
  machine_type     = "e2-standard-2"

 disk {
    boot              = true
    disk_size_gb = 20
  }

}

resource "google_compute_instance_group_manager" "my-group-manager" {
  name = "my-group-manager"

  base_instance_name = "managed"
  zone               = "europe-west9-a"

  version {
    instance_template  = google_compute_instance_template.my-instance-template.id
  }

  target_size  = 3
}