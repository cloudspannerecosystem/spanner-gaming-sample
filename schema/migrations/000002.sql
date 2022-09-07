/*
Copyright 2022 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

CREATE TABLE games (
	gameUUID string(36) NOT NULL,
	players ARRAY<STRING(36)> NOT NULL,
	winner STRING(36),
	created TIMESTAMP,
	finished TIMESTAMP
) PRIMARY KEY (gameUUID);

ALTER TABLE players
	ADD FOREIGN KEY (current_game) REFERENCES games (gameUUID);
