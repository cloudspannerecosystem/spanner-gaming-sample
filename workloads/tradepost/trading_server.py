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

"""Emulate tradingpost server workload"""
import json

from locust import HttpUser, task
from locust.exception import RescheduleTask

class TradeLoad(HttpUser):
    """Players can sell and buy items leveraging the tradepost-service"""

    def item_markup(self, value):
        """Return the 150% the value of the item"""
        float_value = float(value)
        return str(float_value*1.5)

    @task
    def sell_item(self):
        """Task to sell an item on the trading post"""

        headers = {"Content-Type": "application/json"}

        # Get a random item
        with self.client.get("/trades/player_items", headers=headers,
                            catch_response=True) as response:
            try:
                player_uuid = response.json()["PlayerUUID"]
                player_item_uuid = response.json()["PlayerItemUUID"]
                list_price = self.item_markup(response.json()["Price"])

                # Currently don't have any items that can be sold, retry
                if player_item_uuid == "":
                    raise RescheduleTask()

                data = {"lister": player_uuid, "playerItemUUID": player_item_uuid,
                        "list_price": list_price}
                self.client.post("/trades/sell", data=json.dumps(data), headers=headers)
            except json.JSONDecodeError:
                response.failure("Response could not be decoded as JSON")
            except KeyError:
                response.failure("Response did not contain expected keys 'PlayerUUID'"
                                  + "or 'PlayerItemUUID")

    @task
    def buy_item(self):
        """Task to buy an item off the trading post"""
        headers = {"Content-Type": "application/json"}

        # Get a random item
        with self.client.get("/trades/open", headers=headers, catch_response=True) as response:
            try:
                order_uuid = response.json()["OrderUUID"]
                buyer_uuid = response.json()["BuyerUUID"]

                # Currently don't have any buyers that can fill the order, retry
                if buyer_uuid == "":
                    raise RescheduleTask()

                data = {"orderUUID": order_uuid, "buyer": buyer_uuid}
                self.client.put("/trades/buy", data=json.dumps(data), headers=headers)
            except json.JSONDecodeError:
                response.failure("Response could not be decoded as JSON")
            except KeyError:
                response.failure("Response did not contain expected key 'OrderUUID' or 'BuyerUUID'")
