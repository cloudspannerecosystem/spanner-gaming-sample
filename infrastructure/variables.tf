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

variable "gcp_project" {
  type = string
}

variable "resource_env_label" {
  type        = string
  description = "Label/Tag to apply to resources"
}


variable "spanner_config" {
  type = object({
    instance_name       = string
    database_name       = string
    configuration       = string
    display_name        = string
    processing_units    = number
    deletion_protection = bool
  })
  description = "The configuration specifications for the Spanner instance"

  validation {
    condition     = length(var.spanner_config.display_name) >= 4 && length(var.spanner_config.display_name) <= "30"
    error_message = "Display name must be between 4-30 characters long."
  }

  validation {
    condition     = (var.spanner_config.processing_units <= 1000) && (var.spanner_config.processing_units%100) == 0
    error_message = "Processing units must be 1000 or less, and be a multiple of 100."
  }
}

variable "gcp_project_services" {
  type        = list(any)
  description = "GCP Service APIs (<api>.googleapis.com) to enable for this project"
  default     = [
    "artifactregistry.googleapis.com",
    "clouddeploy.googleapis.com",
    "cloudbuild.googleapis.com",
    "container.googleapis.com",
    "spanner.googleapis.com"
  ]
}

variable "gke_config" {
  type = object({
    cluster_name = string
    location = string
    resource_labels = map(string)
  })

  description = "The configuration specifications for a GKE Autopilot cluster"
}

variable backend_sa_config {
  type = object({
    name            = string
    description     = string
  })
  description = "The configuration specifications for the backend service account"
}

variable "k8s_service_account_id" {
  description = "The kubernetes service account that will impersonate the IAM service account to access Cloud Spanner. This account will be created."
}

### Artifact Registry Variables ###

variable "artifact_registry_config" {
  type = object({
    id       = string
    location = string
  })
}

### Cloud Deploy Variables ###

variable "clouddeploy_config" {
  type = object({
    pipeline_name = string
    location      = string
  })
}

variable "skaffold_version" {
  type        = string
  description = "Version of skaffold to use for Cloud Build files"
  default     = "1.39"
}

variable "services_directory" {
  type        = string
  description = "Directory where backend services are found; used by Cloud Deploy"
  default     = "../backend_services" # Relative to Terraform directory
}

variable "workload_directory" {
  type        = string
  description = "Directory where workload generators are found; used by Cloud Deploy"
  default     = "../workloads" # Relative to Terraform directory
}
