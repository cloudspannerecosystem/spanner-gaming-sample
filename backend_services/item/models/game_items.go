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
// data for the item service.
//
// Provides models for game_items, players and player_items
package models

import (
	"context"
	"math/big"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// GameItem represents information about a game_item
type GameItem struct {
	ItemUUID       string    `json:"itemUUID"`
	Item_name      string    `json:"item_name"`
	Item_value     big.Rat   `json:"item_value"`
	Available_time time.Time `json:"available_time"`
	Duration       int64     `json:"duration"`
}

// generateUUID is a private helper to create and returns a v4 UUID string.
func generateUUID() string {
	return uuid.NewString()
}

// readRows is a helper function to read rows from Spanner.
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

// GetItemUUIDs returns a list of item UUIDs
// TODO: Currently limits to 10k by default, but this should be configurable.
func GetItemUUIDs(ctx context.Context, client spanner.Client) ([]string, error) {
	ro := client.ReadOnlyTransaction()
	stmt := spanner.Statement{SQL: `SELECT itemUUID FROM game_items LIMIT 10000`}
	iter := ro.QueryWithOptions(ctx, stmt, spanner.QueryOptions{RequestTag: "app=item,action=GetGameItems"})
	defer iter.Stop()

	itemRows, err := readRows(iter)
	if err != nil {
		return nil, err
	}

	var itemUUIDs []string

	for _, row := range itemRows {
		var iUUID string
		if err := row.Columns(&iUUID); err != nil {
			return nil, err
		}

		itemUUIDs = append(itemUUIDs, iUUID)
	}

	return itemUUIDs, nil
}

// GetItemPrice returns an item's price when provided a valid item uuid
func GetItemPrice(ctx context.Context, txn *spanner.ReadWriteTransaction, itemUUID string) (big.Rat, error) {
	var price big.Rat

	row, err := txn.ReadRowWithOptions(ctx, "game_items", spanner.Key{itemUUID}, []string{"item_value"},
		&spanner.ReadOptions{RequestTag: "app=item,action=GetGameItemPrice"})
	if err != nil {
		return price, err
	}

	err = row.Columns(&price)
	if err != nil {
		return price, err
	}

	return price, nil
}

// Create adds a new game_item to the database
// A game_item uuid is generated, and the available_time is set if none is provided
func (i *GameItem) Create(ctx context.Context, client spanner.Client) error {
	// Initialize item values
	i.ItemUUID = generateUUID()

	if i.Available_time.IsZero() {
		i.Available_time = time.Now()
	}

	// insert into spanner
	_, err := client.ReadWriteTransactionWithOptions(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL: `INSERT game_items (itemUUID, item_name, item_value, available_time, duration) VALUES
					(@itemUUID, @itemName, @itemValue, @availableTime, @duration)
			`,
			Params: map[string]interface{}{
				"itemUUID":      i.ItemUUID,
				"itemName":      i.Item_name,
				"itemValue":     i.Item_value,
				"availableTime": i.Available_time,
				"duration":      i.Duration,
			},
		}

		_, err := txn.UpdateWithOptions(ctx, stmt, spanner.QueryOptions{RequestTag: "app=item,action=AddGameItem"})
		return err
	}, spanner.TransactionOptions{TransactionTag: "app=item,action=add_game_item"})

	if err != nil {
		return err
	}

	// return empty error on success
	return nil
}

// GetItemByUUID returns information about an item when provided a valid game_item UUID
func GetItemByUUID(ctx context.Context, client spanner.Client, itemUUID string) (GameItem, error) {
	row, err := client.Single().ReadRowWithOptions(ctx, "game_items",
		spanner.Key{itemUUID}, []string{"itemUUID", "item_name", "item_value", "available_time", "duration"},
		&spanner.ReadOptions{RequestTag: "app=item,action=GetGameItemByUuid"})
	if err != nil {
		return GameItem{}, err
	}

	item := GameItem{}
	err = row.ToStruct(&item)

	if err != nil {
		return GameItem{}, err
	}
	return item, nil
}
