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

# Basic compute permissions
resource "google_project_iam_member" "clouddeploy-iam" {
  project = var.gcp_project
  for_each = toset([
    "roles/container.admin",
    "roles/storage.admin"
  ])
  role   = each.key
  member = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"

  depends_on = [google_project_service.service_api]
}

# GKE Autopilot IAM
resource "google_service_account" "backend_sa" {
  for_each = toset(var.backend_service_accounts)
  account_id   = each.value
  display_name = "Service account for the '${each.value}' backend service"
  project    = var.gcp_project
}

# Adds each backend service to the roles/databaseUser binding for the spanner instance/database
resource "google_spanner_database_iam_binding" "backend_iam_spanner" {
  instance = google_spanner_instance.instance.name
  database = google_spanner_database.database.name
  role     = "roles/spanner.databaseUser"

  members = formatlist("serviceAccount:%s", values(google_service_account.backend_sa)[*].email)
}

# Allow each backend service to be impersonated by workload identity
resource "google_service_account_iam_binding" "spanner-workload-identity-binding" {
  for_each            = google_service_account.backend_sa
  service_account_id  = each.value.name
  role                = "roles/iam.workloadIdentityUser"

  members = [
     "serviceAccount:${var.gcp_project}.svc.id.goog[default/${each.key}]",
  ]

  depends_on = [google_project_service.service_api, google_service_account.backend_sa, google_container_cluster.sample-game-gke]
}

# Create a kubernetes service account for each backend service. Workloads use default service account
resource "kubernetes_service_account" "k8s-service-account" {
  for_each  = google_service_account.backend_sa

  metadata{
    name      = each.key
    namespace = "default"

    annotations = {
      "iam.gke.io/gcp-service-account" : each.value.email
    }
  }

  depends_on = [google_container_cluster.sample-game-gke]
}

##### Cloud Deploy IAM #####

resource "google_service_account" "cloudbuild-sa" {
  project      = var.gcp_project
  account_id   = "cloudbuild-cicd"
  display_name = "Cloud Build - CI/CD service account"
}

resource "google_project_iam_member" "cloudbuild-sa-cloudbuild-roles" {
  project = var.gcp_project
  for_each = toset([
    "roles/serviceusage.serviceUsageAdmin",
    "roles/clouddeploy.operator",
    "roles/cloudbuild.builds.builder",
    "roles/container.admin",
    "roles/storage.admin",
    "roles/iam.serviceAccountUser",
    "roles/spanner.viewer",
    "roles/spanner.databaseUser",
    "roles/gkehub.editor",
    "roles/logging.logWriter",
    "roles/clouddeploy.jobRunner"
  ])
  role   = each.key
  member = "serviceAccount:${google_service_account.cloudbuild-sa.email}"
}
