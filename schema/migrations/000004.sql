-- Copyright 2022 Google LLC
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     https://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--

CREATE TABLE trade_orders
(
  orderUUID STRING(36)  NOT NULL,
  lister STRING(36) NOT NULL,
  buyer STRING(36),
  playerItemUUID STRING(36) NOT NULL,
  trade_type STRING(5) NOT NULL,
  list_price NUMERIC NOT NULL,
  created TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP()),
  ended TIMESTAMP,
  expires TIMESTAMP NOT NULL DEFAULT (TIMESTAMP_ADD(CURRENT_TIMESTAMP(), interval 24 HOUR)),
  active BOOL NOT NULL DEFAULT (true),
  cancelled BOOL NOT NULL DEFAULT (false),
  filled BOOL NOT NULL DEFAULT (false),
  expired BOOL NOT NULL DEFAULT (false),
  FOREIGN KEY (playerItemUUID) REFERENCES player_items (playerItemUUID)
) PRIMARY KEY (orderUUID);

CREATE INDEX TradeItem ON trade_orders(playerItemUUID, active);
