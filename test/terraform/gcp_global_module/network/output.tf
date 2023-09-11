output "subnet_name" {
  value = google_compute_subnetwork.default.name
}

output "vpc_name" {
  value = google_compute_network.vpc_network.name
}

