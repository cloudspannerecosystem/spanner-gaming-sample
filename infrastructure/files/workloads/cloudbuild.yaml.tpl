# Copyright 2023 Google LLC All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

serviceAccount: projects/$${PROJECT_ID}/serviceAccounts/cloudbuild-cicd@$${PROJECT_ID}.iam.gserviceaccount.com
steps:

#
# Building of images
#
  - name: gcr.io/cloud-builders/docker
    id: profile-workload
    args: ["build", ".", "-t", "$${_PROFILE_IMAGE}"]
    dir: profile
    waitFor: ['-']
  - name: gcr.io/cloud-builders/docker
    id: matchmaking-workload
    args: ["build", ".", "-t", "$${_MATCHMAKING_IMAGE}"]
    dir: matchmaking
    waitFor: ['-']
  - name: gcr.io/cloud-builders/docker
    id: game-workload
    args: ["build", ".", "-t", "$${_GAME_IMAGE}"]
    dir: game
    waitFor: ['-']
  - name: gcr.io/cloud-builders/docker
    id: tradepost-workload
    args: ["build", ".", "-t", "$${_TRADEPOST_IMAGE}"]
    dir: tradepost
    waitFor: ['-']
  - name: gcr.io/cloud-builders/docker
    id: item-generator
    args: ["build", ".", "-t", "$${_ITEMGEN_IMAGE}"]
    dir: item_generator
    waitFor: ['-']


#
# Deployment
#
  - name: gcr.io/google.com/cloudsdktool/cloud-sdk
    id: cloud-deploy-release
    entrypoint: gcloud
    args:
      [
        "deploy", "releases", "create", "$${_REL_NAME}",
        "--delivery-pipeline", "${delivery_pipeline}",
        "--skaffold-file", "skaffold.yaml",
        "--skaffold-version", "${skaffold_version}",
        "--images", "profile-workload=$${_PROFILE_IMAGE},matchmaking-workload=$${_MATCHMAKING_IMAGE},game-workload=$${_GAME_IMAGE},tradepost-workload=$${_TRADEPOST_IMAGE},item-generator=$${_ITEMGEN_IMAGE}",
        "--region", "us-central1"
      ]

artifacts:
  images:
    - $${_REGISTRY}/profile-workload
    - $${_REGISTRY}/matchmaking-workload
    - $${_REGISTRY}/game-workload
    - $${_REGISTRY}/tradepost-workload
    - $${_REGISTRY}/item-generator

substitutions:
  _PROFILE_IMAGE: $${_REGISTRY}/profile-workload:$${BUILD_ID}
  _MATCHMAKING_IMAGE: $${_REGISTRY}/matchmaking-workload:$${BUILD_ID}
  _GAME_IMAGE: $${_REGISTRY}/game-workload:$${BUILD_ID}
  _TRADEPOST_IMAGE: $${_REGISTRY}/tradepost-workload:$${BUILD_ID}
  _ITEMGEN_IMAGE: $${_REGISTRY}/item-generator:$${BUILD_ID}
  _REGISTRY: ${artifact_registry_location}-docker.pkg.dev/$${PROJECT_ID}/${artifact_registry_id}
  _REL_NAME: rel-$${BUILD_ID:0:8}
options:
  dynamic_substitutions: true
  machineType: E2_HIGHCPU_8
  logging: CLOUD_LOGGING_ONLY
