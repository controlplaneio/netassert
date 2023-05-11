terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.57.0"
    }
  }
}

provider "google" {
  zone    = var.zone
}

resource "google_container_cluster" "e2etest" {
  name               = var.cluster_name
  initial_node_count = 4
  addons_config {
    network_policy_config {
      disabled = false
    }
  }
  network_policy {
    enabled = true
  }
  ip_allocation_policy {}
  node_config {
    machine_type = "e2-standard-2"
  }

  release_channel {
    channel = var.cluster_version
  }

  provisioner "local-exec" {
    environment = {
      KUBECONFIG = var.kubeconfig_file
    }
    command = "gcloud container clusters get-credentials ${self.name} --region ${self.location}"
  }
}



