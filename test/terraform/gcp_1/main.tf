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

resource "google_compute_instance" "first" {
  name         = "cbf-test-vm"
  machine_type = "a2-highgpu-1g"
  zone         = "europe-west9-a"
  tags         = ["ssh"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
      size  = 567
      type  = "pd-balanced"
    }
  }

  scratch_disk {
    interface = "NVME"
  }
  scratch_disk {
    interface = "NVME"
  }

  # Install Flask
  metadata_startup_script = "sudo apt-get update; sudo apt-get install -yq build-essential python3-pip rsync; pip install flask"

  network_interface {
    subnetwork = google_compute_subnetwork.first.id

    access_config {
      # Include this section to give the VM an external IP address
    }
  }

  guest_accelerator {
    type = "nvidia-tesla-a100"
    count = 2
  }
}

resource "google_compute_disk" "first" {
  name = "cbf-disk-first"
  type = "pd-standard"
  zone = "europe-west9-a"
  size = 1024
}

resource "google_compute_region_disk" "regional-first" {
  name          = "cbf-disk-regional-first"
  type          = "pd-standard"
  region        = "europe-west9"
  replica_zones = ["europe-west9-a", "europe-west9-b"]
  size          = 1024
}


resource "google_compute_instance" "second" {
  name             = "cbf-test-other"
  machine_type     = "custom-2-4098"
  min_cpu_platform = "Intel Cascade Lake"
  zone             = "europe-west9-a"
  tags             = ["ssh"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
    }
  }

  attached_disk {
    source = google_compute_disk.first.self_link
  }

  # Install Flask
  metadata_startup_script = "sudo apt-get update; sudo apt-get install -yq build-essential python3-pip rsync; pip install flask"

  network_interface {
    subnetwork = google_compute_subnetwork.first.id

    access_config {
      # Include this section to give the VM an external IP address
    }
  }
}

resource "google_sql_database_instance" "instance" {
  name             = "my-database-instance"
  region           = "europe-west9"
  database_version = "POSTGRES_14"
  settings {
    tier = "db-n1-standard-4"
    availability_type = "REGIONAL"
  }
}

