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

// Package main exposes the REST endpoints for the item-service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	spanner "cloud.google.com/go/spanner"
	"github.com/cloudspannerecosystem/spanner-gaming-sample/gaming-item-service/config"
	"github.com/cloudspannerecosystem/spanner-gaming-sample/gaming-item-service/models"
	"github.com/gin-gonic/gin"
)

// setSpannerConnection is a mutator to create spanner context and client, and set them in gin
func setSpannerConnection(c config.Config) gin.HandlerFunc {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, c.Spanner.DB())

	if err != nil {
		log.Fatal(err)
	}

	return func(c *gin.Context) {
		c.Set("spanner_client", *client)
		c.Set("spanner_context", ctx)
		c.Next()
	}
}

// getSpannerConnection is a helper function to retrieve spanner client and context
func getSpannerConnection(c *gin.Context) (context.Context, spanner.Client) {
	return c.MustGet("spanner_context").(context.Context),
		c.MustGet("spanner_client").(spanner.Client)
}

// createItem responds to the POST /items endpoint
// Creates a new game_item and returns the information as a response
func createItem(c *gin.Context) {
	var item models.GameItem

	if err := c.BindJSON(&item); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	ctx, client := getSpannerConnection(c)
	if err := item.Create(ctx, client); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	c.IndentedJSON(http.StatusCreated, item.ItemUUID)
}

// getItemUUIDs responds to the GET /items endpoint
// Returns an unfiltered list of game_item UUIDs.
// TODO: used by game server to generate load. Should not be called by other entities,
//  so restrictions should be implemented
func getItemUUIDs(c *gin.Context) {
	ctx, client := getSpannerConnection(c)

	items, err := models.GetItemUUIDs(ctx, client)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "No items exist"})
		return
	}

	c.IndentedJSON(http.StatusOK, items)
}

// getItem responds to the GET /items/:id endpoint
// Returns information about a specific game_item when provided a valid itemUUID
func getItem(c *gin.Context) {
	var itemUUID = c.Param("id")

	ctx, client := getSpannerConnection(c)

	item, err := models.GetItemByUUID(ctx, client, itemUUID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "item not found"})
		return
	}

	type ReturnGameItem struct {
		ItemUUID, Item_name, Item_value string
		Available_time                  time.Time
		duration                        int64
	}

	gi := ReturnGameItem{ItemUUID: item.ItemUUID, Item_name: item.Item_name, Item_value: item.Item_value.FloatString(2),
		Available_time: item.Available_time, duration: item.Duration}

	c.IndentedJSON(http.StatusOK, gi)
}

// updatePlayerBalance responds to the PUT /players/balance endpoint
// Update a player balance with a provided amount. Result is a JSON object that contains PlayerUUID and AccountBalance
// TODO: fix code to update a player's balance, not a ledger balance
func updatePlayerBalance(c *gin.Context) {
	var player models.Player
	var ledger models.PlayerLedger

	if err := c.BindJSON(&ledger); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	ctx, client := getSpannerConnection(c)
	if err := ledger.UpdateBalance(ctx, client, &player); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	type PlayerBalance struct {
		PlayerUUID, AccountBalance string
	}

	balance := PlayerBalance{PlayerUUID: player.PlayerUUID, AccountBalance: player.Account_balance.FloatString(2)}
	c.IndentedJSON(http.StatusOK, balance)
}

// getPlayer responds to the GET /players endpoint
// Returns information about a random player that is currently playing a game
func getPlayer(c *gin.Context) {
	ctx, client := getSpannerConnection(c)
	player, err := models.GetPlayer(ctx, client)
	if err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	c.IndentedJSON(http.StatusOK, player)
}

// addPlayerItem responds to the POST /players/items endpoint
// Adds an item to the player's list of items when provided a valid game itemUUID.
// TODO: ensure only private access from valid game servers
func addPlayerItem(c *gin.Context) {
	var playerItem models.PlayerItem

	if err := c.BindJSON(&playerItem); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	ctx, client := getSpannerConnection(c)
	if err := playerItem.Add(ctx, client); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	c.IndentedJSON(http.StatusCreated, playerItem)
}

// main initializes the gin router and configures the endpoints
func main() {
	configuration, _ := config.NewConfig()

	router := gin.Default()
	// TODO: Better configuration of trusted proxy
	if err := router.SetTrustedProxies(nil); err != nil {
		fmt.Printf("could not set trusted proxies: %s", err)
		return
	}

	router.Use(setSpannerConnection(configuration))

	router.GET("/items", getItemUUIDs)
	router.POST("/items", createItem)
	router.GET("/items/:id", getItem)
	router.PUT("/players/balance", updatePlayerBalance) // TODO: leverage profile service instead
	router.GET("/players", getPlayer)
	router.POST("/players/items", addPlayerItem)

	if err := router.Run(configuration.Server.URL()); err != nil {
		fmt.Printf("could not run gin router: %s", err)
		return
	}
}
