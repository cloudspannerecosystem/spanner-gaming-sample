// Copyright 2023 Google LLC
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


resource "google_clouddeploy_target" "services_deploy_target" {
  location    = var.gke_config.location
  name        = "backend-services-target"
  description = "Backend Services Deploy Target"

  gke {
    cluster = data.google_container_cluster.gke-provider.id
  }

  require_approval = false

  labels = {
    "environment" = var.resource_env_label
  }

  depends_on = [google_project_service.service_api, google_container_cluster.sample-game-gke]
}

resource "google_clouddeploy_delivery_pipeline" "services_pipeline" {
  location = var.gke_config.location
  name     = var.services_clouddeploy_config.pipeline_name

  description = var.services_clouddeploy_config.pipeline_description

  labels = {
    "environment" = var.resource_env_label
  }

  serial_pipeline {
    stages {
      target_id = google_clouddeploy_target.services_deploy_target.target_id
    }
  }

  depends_on = [google_clouddeploy_target.services_deploy_target]
}

# Workloads deploy to same GKE cluster, but separate pipeline
resource "google_clouddeploy_target" "workloads_deploy_target" {
  location    = var.gke_config.location
  name        = "workloads-target"
  description = "Workloads Deploy Target"

  gke {
    cluster = data.google_container_cluster.gke-provider.id
  }

  require_approval = false

  labels = {
    "environment" = var.resource_env_label
  }

  depends_on = [google_project_service.service_api, google_container_cluster.sample-game-gke]
}

resource "google_clouddeploy_delivery_pipeline" "workloads_pipeline" {
  location = var.gke_config.location
  name     = var.workloads_clouddeploy_config.pipeline_name

  description = var.workloads_clouddeploy_config.pipeline_description

  labels = {
    "environment" = var.resource_env_label
  }

  serial_pipeline {
    stages {
      target_id = google_clouddeploy_target.workloads_deploy_target.target_id
    }
  }

  depends_on = [google_clouddeploy_target.workloads_deploy_target]
}
