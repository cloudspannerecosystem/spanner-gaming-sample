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

import json

# Generate games
# A game consists of 100 players. Only 1 winner randomly selected from those players
#
# Matchmaking is random list of players that are not playing. This is sufficient for testing
# purposes, but is too simple for real use-cases. In real scenarios, something like OpenMatch
# should be used for matchmaking. https://github.com/googleforgames/open-match
#
# To achieve this
# A locust user 'GameMatch' will start off by creating a "game"
# Then, pre-selecting a subset of users, and set a current_game attribute for those players.
# Once done, after a period of time, a winner is randomly selected.

# Create and close game matches
class GameMatch(HttpUser):
    def on_start(self):
        global openGames
        # TODO: prepopulate list of exiting open games
        openGames = []

    @task(1)
    def createGame(self):
        headers = {"Content-Type": "application/json"}

        # Create the game, then store the response in memory of list of open games.
        with self.client.post("/games/create", headers=headers, catch_response=True) as response:
            try:
                openGames.append({"gameUUID": response.json()})
            except json.JSONDecodeError:
                response.failure("Response could not be decoded as JSON")
            except KeyError:
                response.failure("Response did not contain expected key 'gameUUID'")

    @task(1)
    def closeGame(self):
        # No open games are in memory, reschedule task to run again later.
        if len(openGames) == 0:
            raise RescheduleTask()

        headers = {"Content-Type": "application/json"}

        # Close the first open game in our list, removing it to avoid contention from concurrent requests
        game = openGames[0]
        del openGames[0]

        data = {"gameUUID": game["gameUUID"]}
        self.client.put("/games/close", data=json.dumps(data), headers=headers)
        # with self.client.get("/games/open", headers=headers, catch_response=True) as response:
        #     try:
        #         data = {"gameUUID": response.json()["gameUUID"]}
        #         self.client.put("/games/close", data=json.dumps(data), headers=headers)
        #     except json.JSONDecodeError:
        #         response.failure("Response could not be decoded as JSON")
        #     except KeyError:
        #         response.failure("Response did not contain expected key 'playerUUID'")


