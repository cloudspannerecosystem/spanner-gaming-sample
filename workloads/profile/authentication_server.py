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

    # Stores a list of players that were added during the run to
    # be used to make requests for player load
    new_players = []

    # Stores a list of players that have been logged in during the run
    logged_in_players = []

    def generate_player_name(self):
        """Generate a random player name 32 characters long"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=32))

    def generate_password(self, seed):
        """Reverse seed for password to allow known login credentials for load testing"""
        return seed[::-1]

    def generate_email(self):
        """Generate a random email for a subset of domains"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=32) + ['@'] +
            random.choices(['gmail', 'yahoo', 'microsoft']) + ['.com'])

    @task
    def create_player(self):
        """Task to add a player"""
        player_name = self.generate_player_name()
        headers = {"Content-Type": "application/json"}
        data = {"player_name": player_name, "email": self.generate_email(),
                "password": self.generate_password(player_name)}

        with self.client.post("/players", data=json.dumps(data), headers=headers,
                                catch_response=True) as response:
            try:
                # Add player UUID to the data and append to list of players created in this run
                data["player_uuid"] = response.json()
                self.new_players.append(data)
            except json.JSONDecodeError:
                response.failure("Response could not be decoded as JSON")

    @task(3)
    def login(self):
        """Task to login players"""

        # No new_players are in memory, reschedule task to run again later.
        if len(self.new_players) == 0:
            raise RescheduleTask()

        # Get first player in our list, removing it to avoid contention from concurrent requests
        player = self.new_players[0]
        del self.new_players[0]

        headers = {"Content-Type": "application/json"}

        data = { "email": player["email"], "password": player["password"]}

        self.client.put("/players/login", data=json.dumps(data), headers=headers)

        # Append player to 'logged in player' to be used later
        self.logged_in_players.append(player)

    @task(4)
    def get_player(self):
        """Task to get a logged_in player by their uuid"""

        # No logged_in players are in memory, reschedule task to run again later.
        if len(self.logged_in_players) == 0:
            raise RescheduleTask()

        # Get first player in our list, removing it to avoid contention from concurrent requests
        player = self.logged_in_players[0]
        del self.logged_in_players[0]

        headers = {"Content-Type": "application/json"}
        player_uuid = player["player_uuid"]
        self.client.get(f"/players/{player_uuid}", headers=headers, name="/players/[playerUUID]")

        # Add player to 'logged in player' to be re-used later
        self.logged_in_players.append(player)

    @task
    def logout(self):
        """Task to logout players"""

        # No logged_in_players are in memory, reschedule task to run again later.
        if len(self.logged_in_players) == 0:
            raise RescheduleTask()

        # Get first player in our list, removing it to avoid contention from concurrent requests
        player = self.logged_in_players[0]
        del self.logged_in_players[0]

        headers = {"Content-Type": "application/json"}

        data = { "playerUUID": player["player_uuid"] }

        self.client.put("/players/logout", data=json.dumps(data), headers=headers)
