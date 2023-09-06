resource "google_compute_instance" "cbf-test-vm" {
  name         = "cbf-compute-${var.instance_role}"
  machine_type = var.instance_type
  zone         = "${var.region}-a"
  project      = var.project_id
  tags         = ["ssh"]

  boot_disk {
    initialize_params {
      image = data.google_compute_image.debian.self_link
    }
  }


  # Install Flask
  metadata_startup_script = "sudo apt-get update; sudo apt-get install -yq build-essential python3-pip rsync; pip install flask"

  network_interface {
    subnetwork = var.subnet_name
  }

}


data "google_compute_image" "debian" {
  family  = "debian-11"
  project = "debian-cloud"
}
