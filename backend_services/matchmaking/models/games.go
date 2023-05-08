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

// Package models interacts with the backend database to handle the stateful
// data for the matchmaking service.
//
// Provides models for games and players relevant to matchmaking
package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	spanner "cloud.google.com/go/spanner"
	"github.com/google/uuid"
	iterator "google.golang.org/api/iterator"
)

// Game represents information for a single game
type Game struct {
	GameUUID string           `json:"gameUUID"`
	Players  []string         `json:"players"`
	Winner   string           `json:"winner"`
	Created  time.Time        `json:"created"`
	Finished spanner.NullTime `json:"finished"`
}

// generateUUID is a private helper to create and returns a v4 UUID string.
func generateUUID() string {
	return uuid.NewString()
}

// readRows is a private helper function to read rows from Spanner.
func readRows(iter *spanner.RowIterator) ([]spanner.Row, error) {
	var rows []spanner.Row
	defer iter.Stop()

	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		rows = append(rows, *row)
	}

	return rows, nil
}

// determineWinner returns the uuid of the winning player
// Current implementation is a random player from the list of players assigned to the game
func determineWinner(playerUUIDs []string) string {
	if len(playerUUIDs) == 0 {
		return ""
	}

	var winnerUUID string

	rand.Seed(time.Now().UnixNano())
	offset := rand.Intn(len(playerUUIDs))
	winnerUUID = playerUUIDs[offset]
	return winnerUUID
}

// getGamePlayers returns player information for a specificied game
// We only care about the playerUUID and their stats, as this is intended to be used
// to modify players when a game is closed. We get the current_game to make sure later that the player is part of the game.
func (g Game) getGamePlayers(ctx context.Context, txn *spanner.ReadWriteTransaction) ([]string, []Player, error) {
	stmt := spanner.Statement{
		SQL: `SELECT PlayerUUID, Stats, Current_game FROM players
				INNER JOIN (
				SELECT pUUID FROM games g, UNNEST(g.Players) AS pUUID WHERE gameUUID=@game
				) AS gPlayers ON gPlayers.pUUID = players.PlayerUUID;`,
		Params: map[string]interface{}{
			"game": g.GameUUID,
		},
	}

	iter := txn.QueryWithOptions(ctx, stmt, spanner.QueryOptions{RequestTag: "app=matchmaking,action=GetGamePlayers"})
	playerRows, err := readRows(iter)
	if err != nil {
		return []string{}, []Player{}, err
	}

	var playerUUIDs []string
	var players []Player
	for _, row := range playerRows {
		var p Player

		if err := row.ToStruct(&p); err != nil {
			return []string{}, []Player{}, err
		}
		if p.Stats.IsNull() {
			// Initialize player stats
			p.Stats = spanner.NullJSON{Value: PlayerStats{
				Games_played: 0,
				Games_won:    0,
			}, Valid: true}
		}

		players = append(players, p)
		playerUUIDs = append(playerUUIDs, p.PlayerUUID)
	}

	return playerUUIDs, players, nil
}

// GetOpenGame returns the gameUUID of an open game.
func GetOpenGame(ctx context.Context, client spanner.Client) (Game, error) {
	var g Game

	// Get an open game
	// Retrieve a random unfinished from 10 of the oldest games to reduce concurrency contention when closing the same game
	query := fmt.Sprintf("SELECT gameUUID FROM (SELECT gameUUID FROM games WHERE finished IS NULL ORDER BY created DESC LIMIT 10) TABLESAMPLE RESERVOIR (%d ROWS)", 1)

	stmt := spanner.Statement{SQL: query}

	iter := client.Single().QueryWithOptions(ctx, stmt, spanner.QueryOptions{RequestTag: "app=matchmaking,action=GetOpenGame"})
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return Game{}, err
		}

		if err := row.ToStruct(&g); err != nil {
			return Game{}, err
		}
	}
	return g, nil
}

// updateGamePlayers updates a game's player statistics when closing out a game.
// Updating players involves closing out the game (current_game = NULL) and
// updating their game stats. Specifically, we are incrementing games_played.
// If the player is the determined winner, then their games_won stat is incremented.
func (g Game) updateGamePlayers(txn *spanner.ReadWriteTransaction, players []Player) error {
	for _, p := range players {
		// Modify stats
		var pStats PlayerStats
		if err := json.Unmarshal([]byte(p.Stats.String()), &pStats); err != nil {
			return fmt.Errorf("could not unmarshal json: %s", err)
		}

		pStats.Games_played = pStats.Games_played + 1

		if p.PlayerUUID == g.Winner {
			pStats.Games_won = pStats.Games_won + 1
		}
		updatedStats, _ := json.Marshal(pStats)
		if err := p.Stats.UnmarshalJSON(updatedStats); err != nil {
			return fmt.Errorf("could not unmarshal json: %s", err)
		}

		// Update player
		// If player's current game isn't the same as this game, that's an error
		if p.Current_game != g.GameUUID {
			return fmt.Errorf("player '%s' doesn't belong to game '%s'", p.PlayerUUID, g.GameUUID)
		}

		cols := []string{"playerUUID", "current_game", "stats"}
		newGame := spanner.NullString{
			StringVal: "",
			Valid:     false,
		}

		err := txn.BufferWrite([]*spanner.Mutation{
			spanner.Update("players", cols, []interface{}{p.PlayerUUID, newGame, p.Stats}),
		})

		if err != nil {
			return fmt.Errorf("could not buffer write: %s", err)
		}
	}

	return nil
}

// CreateGame starts a new game and assign players
// Players that are not currently playing a game are eligble to be selected for the new game
// Current implementation allows for less than numPlayers to be placed in a game
func (g *Game) CreateGame(ctx context.Context, client spanner.Client) error {
	// Initialize game values
	g.GameUUID = generateUUID()

	numPlayers := 10

	// Create and assign
	_, err := client.ReadWriteTransactionWithOptions(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		var m []*spanner.Mutation

		// get players
		query := fmt.Sprintf("SELECT playerUUID FROM (SELECT playerUUID FROM players WHERE current_game IS NULL LIMIT 10000) TABLESAMPLE RESERVOIR (%d ROWS)", numPlayers)
		stmt := spanner.Statement{SQL: query}
		iter := txn.QueryWithOptions(ctx, stmt, spanner.QueryOptions{RequestTag: "app=matchmaking,action=AssignPlayers"})

		playerRows, err := readRows(iter)
		if err != nil {
			return err
		}

		var playerUUIDs []string

		for _, row := range playerRows {
			var pUUID string
			if err := row.Columns(&pUUID); err != nil {
				return err
			}

			playerUUIDs = append(playerUUIDs, pUUID)
		}

		// Create the game
		gCols := []string{"gameUUID", "players", "created"}
		m = append(m, spanner.Insert("games", gCols, []interface{}{g.GameUUID, playerUUIDs, time.Now()}))

		// Update players to lock into this game

		for _, p := range playerUUIDs {
			pCols := []string{"playerUUID", "current_game"}
			m = append(m, spanner.Update("players", pCols, []interface{}{p, g.GameUUID}))
		}

		if err := txn.BufferWrite(m); err != nil {
			return fmt.Errorf("could not buffer write: %s", err)
		}

		return nil
	}, spanner.TransactionOptions{TransactionTag: "app=matchmaking,action=create_game"})

	if err != nil {
		return err
	}

	return nil
}

// CloseGame chooses a random winner and closes the game when provided a game UUID
// A game is closed by setting the winner and finished time.
// Additionally all players' game stats are updated, and the current_game is set to null to allow
// them to be chosen for a new game.
func (g *Game) CloseGame(ctx context.Context, client spanner.Client) error {
	// Close game
	_, err := client.ReadWriteTransactionWithOptions(ctx,
		func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			// Validate game finished time is null
			row, err := txn.ReadRow(ctx, "games", spanner.Key{g.GameUUID}, []string{"finished"})
			if err != nil {
				return err
			}

			if err := row.Column(0, &g.Finished); err != nil {
				return err
			}

			// If time is not null, then the game is already marked as finished. That's an error.
			if !g.Finished.IsNull() {
				errorMsg := fmt.Sprintf("Game '%s' is already finished.", g.GameUUID)
				return errors.New(errorMsg)
			}

			// Get game players
			playerUUIDs, players, err := g.getGamePlayers(ctx, txn)

			if err != nil {
				return err
			}

			// Might be an issue if there are no players!
			if len(playerUUIDs) == 0 {
				errorMsg := fmt.Sprintf("No players found for game '%s'", g.GameUUID)
				return errors.New(errorMsg)
			}

			// Get random winner
			g.Winner = determineWinner(playerUUIDs)

			cols := []string{"gameUUID", "finished", "winner"}
			err = txn.BufferWrite([]*spanner.Mutation{
				spanner.Update("games", cols, []interface{}{g.GameUUID, time.Now(), g.Winner}),
			})

			if err != nil {
				return fmt.Errorf("could not buffer write: %s", err)
			}

			// Update each player to increment stats.games_played (and stats.games_won if winner),
			// and set current_game to null so they can be chosen for a new game
			if err := g.updateGamePlayers(txn, players); err != nil {
				return err
			}

			return nil
		}, spanner.TransactionOptions{TransactionTag: "app=matchmaking,action=close_game"})

	if err != nil {
		return err
	}

	return nil
}
