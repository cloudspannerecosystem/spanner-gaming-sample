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
  name     = var.gke_config.cluster_name
  location = var.gke_config.location

  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name

  # See issue: https://github.com/hashicorp/terraform-provider-google/issues/10782
  ip_allocation_policy {}

  # Enabling Autopilot for this cluster
  enable_autopilot = true
}

resource "google_service_account" "backend_sa" {
  account_id   = var.backend_sa_config.name
  display_name = var.backend_sa_config.description
  project    = var.gcp_project
}

resource "kubernetes_service_account" "k8s-service-account" {
  metadata {
    name      = var.k8s_service_account_id
    namespace = "default"
    annotations = {
      "iam.gke.io/gcp-service-account" : "${google_service_account.backend_sa.email}"
    }
  }
}

data "google_iam_policy" "spanner-policy" {
  binding {
    role = "roles/iam.workloadIdentityUser"
    members = [
      "serviceAccount:${var.gcp_project}.svc.id.goog[default/${kubernetes_service_account.k8s-service-account.metadata[0].name}]"
    ]
  }
}

resource "google_service_account_iam_policy" "backend-service-account-iam" {
  service_account_id = google_service_account.backend_sa.name
  policy_data        = data.google_iam_policy.spanner-policy.policy_data
}
