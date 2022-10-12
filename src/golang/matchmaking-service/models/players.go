// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import spanner "cloud.google.com/go/spanner"

// PlayerStats provides various statistics for a player
type PlayerStats struct {
	Games_played int `json:"games_played"`
	Games_won    int `json:"games_won"`
}

// Player maps to the fields required by a game's players
type Player struct {
	PlayerUUID   string           `json:"playerUUID"`
	Stats        spanner.NullJSON `json:"stats"`
	Current_game string           `json:"current_game"`
}
