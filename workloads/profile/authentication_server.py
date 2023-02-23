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

"""Emulate authentication server workload"""

import string
import json
import random

from locust import HttpUser, task
from locust.exception import RescheduleTask

class PlayerLoad(HttpUser):
    """
    Generate player load by adding new users and retrieving those uuids
    to simulate 5:1 read and write traffic against the profile-service
    """

    # Stores a list of player_uuids that were added during the run to
    # be used to make requests for player load
    player_uuids = []

    def generate_player_name(self):
        """Generate a random player name 32 characters long"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=32))

    def generate_password(self):
        """Generate a random password 32 characters long"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=32))

    def generate_email(self):
        """Generate a random email for a subset of domains"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=32) + ['@'] +
            random.choices(['gmail', 'yahoo', 'microsoft']) + ['.com'])

    @task
    def create_player(self):
        """Task to add a player"""

        headers = {"Content-Type": "application/json"}
        data = {"player_name": self.generate_player_name(), "email": self.generate_email(),
                "password": self.generate_password()}

        with self.client.post("/players", data=json.dumps(data), headers=headers,
                                catch_response=True) as response:
            try:
                self.player_uuids.append(response.json())
            except json.JSONDecodeError:
                response.failure("Response could not be decoded as JSON")

    @task(5)
    def get_player(self):
        """Task to get a player by their uuid"""

        # No player UUIDs are in memory, reschedule task to run again later.
        if len(self.player_uuids) == 0:
            raise RescheduleTask()

        # Get first player in our list, removing it to avoid contention from concurrent requests
        player_uuid = self.player_uuids[0]
        del self.player_uuids[0]

        headers = {"Content-Type": "application/json"}

        self.client.get(f"/players/{player_uuid}", headers=headers, name="/players/[playerUUID]")
