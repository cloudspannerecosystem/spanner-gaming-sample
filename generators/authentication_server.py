# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    https:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from locust import HttpUser, task
from locust.exception import RescheduleTask

import string
import json
import random

# Generate player load with 5:1 reads to write
class PlayerLoad(HttpUser):
    def on_start(self):
        global pUUIDs
        pUUIDs = []

    def generatePlayerName(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=32))

    def generatePassword(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=32))

    def generateEmail(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=32) + ['@'] +
            random.choices(['gmail', 'yahoo', 'microsoft']) + ['.com'])

    @task
    def createPlayer(self):
        headers = {"Content-Type": "application/json"}
        data = {"player_name": self.generatePlayerName(), "email": self.generateEmail(), "password": self.generatePassword()}

        with self.client.post("/players", data=json.dumps(data), headers=headers, catch_response=True) as response:
            try:
                pUUIDs.append(response.json())
            except json.JSONDecodeError:
                response.failure("Response could not be decoded as JSON")
            except KeyError:
                response.failure("Response did not contain expected key 'gameUUID'")

    @task(5)
    def getPlayer(self):
        # No player UUIDs are in memory, reschedule task to run again later.
        if len(pUUIDs) == 0:
            raise RescheduleTask()

        # Get first player in our list, removing it to avoid contention from concurrent requests
        pUUID = pUUIDs[0]
        del pUUIDs[0]

        headers = {"Content-Type": "application/json"}

        self.client.get(f"/players/{pUUID}", headers=headers, name="/players/[playerUUID]")
