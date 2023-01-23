data "google_compute_image" "debian" {
  family  = "debian-11"
  project = "debian-cloud"
}

resource "google_compute_disk" "diskImage" {
  name   = "cbf-disk-image"
  zone = "europe-west9-a"
  image = data.google_compute_image.debian.self_link
}
