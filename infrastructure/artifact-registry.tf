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

resource "google_artifact_registry_repository" "container_registry" {
  repository_id = var.artifact_registry_config.id
  location      = var.artifact_registry_config.location
  description   = "Repository for container images for the global game"
  format        = "Docker"

  labels = {
    "environment" = var.resource_env_label
  }

  depends_on = [google_project_service.project]
}

resource "local_file" "backend-service-build" {
  content = templatefile(
    "${path.module}/files/backend_services/cloudbuild.yaml.tpl", {
      project_id  = var.gcp_project
      artifact_registry_id = var.artifact_registry_config.id
      artifact_registry_location = var.artifact_registry_config.location
      skaffold_version = var.skaffold_version
      delivery_pipeline = google_clouddeploy_delivery_pipeline.services_pipeline.name
  })
  filename = "${path.module}/${var.services_directory}/cloudbuild.yaml"

  depends_on = [ google_clouddeploy_delivery_pipeline.services_pipeline ]
}

resource "local_file" "workloads-build" {
  content = templatefile(
    "${path.module}/files/workloads/cloudbuild.yaml.tpl", {
      project_id  = var.gcp_project
      artifact_registry_id = var.artifact_registry_config.id
      artifact_registry_location = var.artifact_registry_config.location
      skaffold_version = var.skaffold_version
      delivery_pipeline = google_clouddeploy_delivery_pipeline.services_pipeline.name
  })
  filename = "${path.module}/${var.workload_directory}/cloudbuild.yaml"

  depends_on = [ google_clouddeploy_delivery_pipeline.services_pipeline ]
}
