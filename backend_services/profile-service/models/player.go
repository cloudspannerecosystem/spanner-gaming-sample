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
// data for the profile service.
package models

import (
	"context"
	"errors"
	"math/big"
	"time"

	spanner "cloud.google.com/go/spanner"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var validate *validator.Validate

// PlayerStats provides various statistics for a player
type PlayerStats struct {
	Games_played spanner.NullInt64 `json:"games_played"`
	Games_won    spanner.NullInt64 `json:"games_won"`
}

// Player maps to the fields stored for the backend database
type Player struct {
	PlayerUUID      string           `json:"playerUUID" validate:"omitempty,uuid4"`
	Player_name     string           `json:"player_name" validate:"required_with=Password Email"`
	Email           string           `json:"email" validate:"required_with=Player_name Password,email"`
	Password        string           `json:"password" validate:"required_with=Player_name Email"` // not stored in DB
	Password_hash   []byte           `json:"password_hash"`                                       // stored in DB
	created         time.Time        //lint:ignore U1000 Field is present to map to database schema
	updated         time.Time        //lint:ignore U1000 Field is present to map to database schema
	Stats           spanner.NullJSON `json:"stats"`
	Account_balance big.Rat          `json:"account_balance"`
	last_login      time.Time        //lint:ignore U1000 Field is present to map to database schema
	is_logged_in    bool             //lint:ignore U1000 Field is present to map to database schema
	valid_email     bool             //lint:ignore U1000 Field is present to map to database schema
	Current_game    string           `json:"current_game" validate:"omitempty,uuid4"`
}

func init() {
	validate = validator.New()
}

// hashPassword is a private helper to encrypte a password using the bcrypt library
func hashPassword(pwd string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	return hash, nil
}

// validatePassword is a private helper to ensure a supplied hash matches the stored encrypted
// password.
func validatePassword(pwd string, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(pwd))
}

// generateUUID is a private helper to create and returns a v4 UUID string.
func generateUUID() string {
	return uuid.NewString()
}

// Validate that the player has the required information based on the type's validation rules.
func (p *Player) Validate() error {
	validate = validator.New()
	err := validate.Struct(p)
	if err != nil {
		return err
	}

	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err
	}

	return nil
}

// AddPlayer provides functionality to insert a player into the backend.
// Provide with the required fields from the API call, the password is hashed and
// a UUID is generated. This is then inserted, along with empty stats, into
// the Spanner database.
func (p *Player) AddPlayer(ctx context.Context, client spanner.Client) error {
	// Validate based on struct validation rules
	err := p.Validate()
	if err != nil {
		return err
	}

	// take supplied password+salt, hash. Store in user_password
	passHash, err := hashPassword(p.Password)

	if err != nil {
		return errors.New("unable to hash password")
	}

	p.Password_hash = passHash

	// Generate UUIDv4
	p.PlayerUUID = generateUUID()

	// Initialize player stats
	emptyStats := spanner.NullJSON{Value: PlayerStats{
		Games_played: spanner.NullInt64{Int64: 0, Valid: true},
		Games_won:    spanner.NullInt64{Int64: 0, Valid: true},
	}, Valid: true}

	// insert into spanner
	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL: `INSERT players (playerUUID, player_name, email, password_hash, created, stats) VALUES
					(@playerUUID, @playerName, @email, @passwordHash, CURRENT_TIMESTAMP(), @pStats)
			`,
			Params: map[string]interface{}{
				"playerUUID":   p.PlayerUUID,
				"playerName":   p.Player_name,
				"email":        p.Email,
				"passwordHash": p.Password_hash,
				"pStats":       emptyStats,
			},
		}

		_, err := txn.Update(ctx, stmt)
		return err
	})

	// TODO: Handle 'AlreadyExists' errors
	if err != nil {
		return err
	}

	// return empty error on success
	return nil
}

// GetPlayerByUUID returns a Player based on a provided uuid. In the event of an error
// retrieving the player, an empty Player is returned with the error.
func GetPlayerByUUID(ctx context.Context, client spanner.Client, uuid string) (Player, error) {
	row, err := client.Single().ReadRow(ctx, "players",
		spanner.Key{uuid}, []string{"playerUUID", "player_name", "email", "stats"})
	if err != nil {
		return Player{}, err
	}

	player := Player{}
	err = row.ToStruct(&player)

	if err != nil {
		return Player{}, err
	}
	return player, nil
}

// Getting player by login information
// Uses player name and password. Should return an error if no player was found
// func GetPlayerByLogin(name string, password string) (Player, error) {

// }
