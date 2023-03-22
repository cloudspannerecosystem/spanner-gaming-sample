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
    "roles/storage.admin",
    "roles/logging.logWriter",
    "roles/clouddeploy.jobRunner"
  ])
  role   = each.key
  member = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"

  depends_on = [google_project_service.project]
}

# GKE Autopilot IAM
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

# Create IAM service account for locked down cloud run container to access Spanner service
resource "google_spanner_database_iam_binding" "backend_iam_spanner" {
  instance = google_spanner_instance.instance.name
  database = google_spanner_database.database.name
  role     = "roles/spanner.databaseUser"

  members = [
    "serviceAccount:${google_service_account.backend_sa.email}",
  ]
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
    "roles/spanner.databaseUser",
    "roles/gkehub.editor"
  ])
  role   = each.key
  member = "serviceAccount:${google_service_account.cloudbuild-sa.email}"
}
