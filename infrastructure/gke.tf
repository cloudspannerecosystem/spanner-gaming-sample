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

variable "gke_config" {
  type = object({
    cluster_name = string
    location = string
    resource_labels = map(string)
  })

  description = "The configuration specifications for a GKE Autopilot cluster"
}

resource "google_compute_network" "vpc" {
  name                    = "cymbal-game-staging-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "subnet" {
  name          = "cymbal-game-subnet"
  ip_cidr_range = "10.1.0.0/16"
  region        = "us-central1"
  network       = google_compute_network.vpc.id
}

resource "google_container_cluster" "cymbal-games-gke" {
  name     = var.gke_config.cluster_name
  location = var.gke_config.location

  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name

  # See issue: https://github.com/hashicorp/terraform-provider-google/issues/10782
  ip_allocation_policy {}

# Enabling Autopilot for this cluster
  enable_autopilot = true
}
