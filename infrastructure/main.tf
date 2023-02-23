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

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.44.1"
    }
  }

  required_version = ">= 0.14"
}

provider "google" {
  project = var.gcp_project
}

data "google_project" "project" {
}

resource "google_project_service" "project" {
  for_each = toset(var.gcp_project_services)
  service  = each.value

  disable_on_destroy = false
}

data "google_client_config" "provider" {}

data "google_container_cluster" "gke-provider" {
  name     = var.gke_config.cluster_name
  location = var.gke_config.location
}

provider "kubernetes" {
  host  = "https://${data.google_container_cluster.gke-provider.endpoint}"
  token = data.google_client_config.provider.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.gke-provider.master_auth[0].cluster_ca_certificate,
  )
}
