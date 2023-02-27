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
# A convenience script to help developers easily submit service builds to Cloud Build
if [ -z "${PROJECT_ID}" ]; then
    echo "[ERROR] PROJECT_ID environment variable must be set" >&2
    exit 1
fi

basedir=`pwd`

# Submit a build command to
for service in profile-service matchmaking-service item-service tradepost-service; do
    cd "${basedir}/backend_services/${service}"
    echo "[INFO] Building ${service}"
    gcloud builds submit --tag gcr.io/$PROJECT_ID/$service .

    if [ $? -ne 0 ]; then
        echo "[ERROR] Build failed...stopping further builds" >&2
        exit 1
    fi
done
