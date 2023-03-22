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

// Package main exposes the REST endpoints for the tradepost-service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	spanner "cloud.google.com/go/spanner"
	"github.com/cloudspannerecosystem/spanner-gaming-sample/gaming-tradepost-service/config"
	"github.com/cloudspannerecosystem/spanner-gaming-sample/gaming-tradepost-service/models"
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

// getPlayerItem responds to the GET /trades/player_items endpoint
// Returns information about a random item that can be listed
func getPlayerItem(c *gin.Context) {
	ctx, client := getSpannerConnection(c)

	item, err := models.GetRandomPlayerItem(ctx, client)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "item not found"})
		return
	}

	type RandomItem struct {
		PlayerUUID, PlayerItemUUID, Price string
	}

	ri := RandomItem{PlayerUUID: item.PlayerUUID, PlayerItemUUID: item.PlayerItemUUID, Price: item.Price.FloatString(2)}

	c.IndentedJSON(http.StatusOK, ri)
}

// getOpenOrder responds to the GET /trades/open
// Returns a random open order with a random buyer. Used in trade simulation
func getOpenOrder(c *gin.Context) {
	ctx, client := getSpannerConnection(c)

	// Get an order
	order, err := models.GetRandomOpenOrder(ctx, client)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "order not found"})
		return
	}

	var buyer models.Player

	// Get a buyer; can't be the same player as the trade order's lister.
	// Only do this if an order exists, otherwise avoid a DB call
	if order.OrderUUID != "" {
		buyer, err = models.GetRandomPlayer(ctx, client, order.Lister, order.ListPrice)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "order not found"})
			return
		}
	}

	type RandomOrder struct {
		OrderUUID, ListPrice, BuyerUUID, AccountBalance string
	}

	ro := RandomOrder{OrderUUID: order.OrderUUID, BuyerUUID: buyer.PlayerUUID, ListPrice: order.ListPrice.FloatString(2), AccountBalance: buyer.AccountBalance.FloatString(2)}

	c.IndentedJSON(http.StatusOK, ro)
}

// createOrder responds to the POST /trades/sell endpoint
// Creates a sell order and returns information about the created order
func createOrder(c *gin.Context) {
	var order models.TradeOrder

	if err := c.BindJSON(&order); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	ctx, client := getSpannerConnection(c)
	if err := order.Create(ctx, client); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	c.IndentedJSON(http.StatusCreated, order.OrderUUID)
}

// purchaseOrder responds to the PUT /trades/buy endpoint
// Closes out a trade order as 'buy' and updates item and account balance information
func purchaseOrder(c *gin.Context) {
	var order models.TradeOrder

	if err := c.BindJSON(&order); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	ctx, client := getSpannerConnection(c)
	if err := order.Buy(ctx, client); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	c.IndentedJSON(http.StatusCreated, order.OrderUUID)
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

	router.GET("/trades/player_items", getPlayerItem)
	router.POST("/trades/sell", createOrder)
	router.GET("/trades/open", getOpenOrder)
	router.PUT("/trades/buy", purchaseOrder)

	if err := router.Run(configuration.Server.URL()); err != nil {
		fmt.Printf("could not run gin router: %s", err)
		return
	}
}
