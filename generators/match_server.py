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

"""Emulate matchmaking server workload"""

import json

from locust import HttpUser, task
from locust.exception import RescheduleTask

class GameMatch(HttpUser):
    """Create and close games to simulate players joining and finishing games
    leveraging the matchmaking-service
    """

    @task(2)
    def create_game(self):
        """Task to create a new game"""

        headers = {"Content-Type": "application/json"}

        # Create the game, then store the response in memory of list of open games.
        self.client.post("/games/create", headers=headers)

    @task(1)
    def close_game(self):
        """Task to close a previously opened game"""
        headers = {"Content-Type": "application/json"}

        # Get an open game to be closed
        with self.client.get("/games/open", headers=headers, catch_response=True) as response:
            try:
                game_uuid = response.json()["gameUUID"]

                # Reschedule task when game_uuid is empty
                if game_uuid == "":
                    raise RescheduleTask()

                data = {"gameUUID": game_uuid}
                self.client.put("/games/close", data=json.dumps(data), headers=headers)
            except json.JSONDecodeError:
                response.failure("Response could not be decoded as JSON")
            except KeyError:
                response.failure("Response did not contain expected key 'gameUUID'")
