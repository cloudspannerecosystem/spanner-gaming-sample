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

"""Emulate creation of new game items"""

import string
import json
import random
import decimal

from locust import HttpUser, task

class ItemLoad(HttpUser):
    """Seed the game with random items using the item-service APIs"""

    def generate_item_name(self):
        """Generate a random item name, 32-characters long"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=32))

    def generate_item_value(self):
        """Generate a random item decimal value, between 1 and 100"""
        return str(decimal.Decimal(random.randrange(100, 10000))/100)

    @task
    def create_item(self):
        """Task to create a new game item"""
        headers = {"Content-Type": "application/json"}
        data = {"item_name": self.generate_item_name(), "item_value": self.generate_item_value()}

        self.client.post("/items", data=json.dumps(data), headers=headers)
