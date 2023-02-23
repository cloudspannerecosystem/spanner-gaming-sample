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

"""Emulate game server workload"""
import json
import random

from locust import HttpUser, task

import requests

class GameLoad(HttpUser):
    """
    Leverage the item-service APIs to allow players to generate items and money
    at a 1:2 ratio to ensure they have enough money to buy items later.
    """
    item_uuids = {}

    def on_start(self):
        """When starting load generator, initialize items"""
        self.get_items()

    def get_items(self):
        """Initialize list of items from endpoint"""
        headers = {"Content-Type": "application/json"}
        req = requests.get(f"{self.host}/items", headers=headers, timeout=10)
        self.item_uuids = json.loads(req.text)

    def generate_amount(self):
        """Generate a random monetary amount between 1 and 50"""
        return str(round(random.uniform(1.01, 49.99), 2))

    @task(2)
    def acquire_money(self):
        """Task for random player to acquire money"""
        headers = {"Content-Type": "application/json"}

        # Get a random player that's part of a game, and update balance
        with self.client.get("/players", headers=headers, catch_response=True) as response:
            try:
                data = {"playerUUID": response.json()["playerUUID"],
                        "amount": self.generate_amount(), "source": "loot"}
                self.client.put("/players/balance", data=json.dumps(data), headers=headers)
            except json.JSONDecodeError:
                response.failure("Response could not be decoded as JSON")
            except KeyError:
                response.failure("Response did not contain expected key 'playerUUID'")

    @task(1)
    def acquire_item(self):
        """Task for random player to acquire an item"""
        headers = {"Content-Type": "application/json"}

        # Get a random player that's part of a game, and add an item
        with self.client.get("/players", headers=headers, catch_response=True) as response:
            try:
                item_uuid = self.item_uuids[random.randint(0, len(self.item_uuids)-1)]
                data = {"playerUUID": response.json()["playerUUID"],
                        "itemUUID": item_uuid, "source": "loot"}
                self.client.post("/players/items", data=json.dumps(data), headers=headers)
            except json.JSONDecodeError:
                response.failure("Response could not be decoded as JSON")
            except KeyError:
                response.failure("Response did not contain expected key 'playerUUID'")
