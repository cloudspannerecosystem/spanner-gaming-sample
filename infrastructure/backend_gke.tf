// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

resource "google_container_cluster" "sample-game-gke" {
  name              = var.gke_config.cluster_name
  location          = var.gke_config.location
  network           = google_compute_network.vpc.name
  subnetwork        = google_compute_subnetwork.subnet.name

  # Use locked down service account
  cluster_autoscaling {
    auto_provisioning_defaults {
      service_account = google_service_account.gke-sa.email
    }
  }

  # Enabling Autopilot for this cluster
  enable_autopilot  = true

  resource_labels = {
    "environment" = var.resource_env_label
  }

  # See issue: https://github.com/hashicorp/terraform-provider-google/issues/10782
  ip_allocation_policy {}

  depends_on = [google_service_account.gke-sa]
}

data "google_container_cluster" "gke-provider" {
  name        = var.gke_config.cluster_name
  location    = var.gke_config.location

  depends_on  = [ google_container_cluster.sample-game-gke ]
}

provider "kubernetes" {
  host  = "https://${data.google_container_cluster.gke-provider.endpoint}"
  token = data.google_client_config.provider.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.gke-provider.master_auth[0].cluster_ca_certificate,
  )
}

