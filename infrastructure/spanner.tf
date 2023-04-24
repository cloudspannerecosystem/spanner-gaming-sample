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
  labels           = {
    "env" = var.resource_env_label
  }

  depends_on = [google_project_service.service_api]
}

resource "google_spanner_database" "database" {
  instance            = google_spanner_instance.instance.name
  name                = var.spanner_config.database_name
  deletion_protection = var.spanner_config.deletion_protection
}

# Make Config file for deploy with Cloud Deploy
resource "local_file" "spanner-config" {
  content = templatefile(
    "${path.module}/files/backend_services/spanner-config.yaml.tpl", {
      project_id    = var.gcp_project
      instance_id   = google_spanner_instance.instance.name
      database_id   = google_spanner_database.database.name
  })
  filename = "${path.module}/${var.services_directory}/spanner_config.yaml"
  depends_on = [ google_spanner_instance.instance, google_spanner_database.database]
}

