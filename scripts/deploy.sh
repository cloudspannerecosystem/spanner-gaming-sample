#!/bin/bash
#
# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# A convenience script to help developers easily deploy services to GKE cluster
if [ -z "${PROJECT_ID}" ]; then
    echo "[ERROR] PROJECT_ID environment variable must be set" >&2
    exit 1
fi

basedir=`pwd`

cd "${basedir}/kubernetes-manifests"

# Submit a kubectl apply for each deployment file
for service in profile-service matchmaking-service item-service tradepost-service; do
    echo "[INFO] Configuring ${service}"
    sed "s/\bPROJECT_ID\b/${PROJECT_ID}/" "${service}.yaml.tmpl" > "${service}.yaml"

    echo "[INFO] Deploying ${service}"
    kubectl apply -f "${service}.yaml"

    if [ $? -ne 0 ]; then
        echo "[ERROR] Deploy failed...stopping further deploys" >&2
        exit 1
    fi
done
