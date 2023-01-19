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

resource "google_spanner_instance" "instance" {
  name             = var.spanner_config.instance_name  # << be careful changing this in production
  config           = var.spanner_config.configuration
  display_name     = var.spanner_config.display_name
  processing_units = var.spanner_config.processing_units
  labels           = { "env" = var.spanner_config.environment }
}

resource "google_spanner_database" "database" {
  instance = google_spanner_instance.instance.name
  name     = var.spanner_config.database_name
  deletion_protection = true
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
