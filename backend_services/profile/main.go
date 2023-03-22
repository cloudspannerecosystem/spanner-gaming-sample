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

// Package main exposes the REST endpoints for the profile-service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	spanner "cloud.google.com/go/spanner"
	"github.com/cloudspannerecosystem/spanner-gaming-sample/gaming-profile-service/config"
	"github.com/cloudspannerecosystem/spanner-gaming-sample/gaming-profile-service/models"

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

// getPlayerID responds to the GET /players/:id endpoint
// Returns a player's information when provided a valid player uuid
func getPlayerByID(c *gin.Context) {
	var playerUUID = c.Param("id")

	ctx, client := getSpannerConnection(c)

	player, err := models.GetPlayerByUUID(ctx, client, playerUUID)
	if err != nil {
		fmt.Printf("Error: %s", err)
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "player not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, player)
}

// playerLogin responds to the PUT /players/login endpoint
// Login requires 'email' and 'password'
// Returns player's information on successful login. Returns 404 on failed login.
func playerLogin(c *gin.Context) {
	type PlayerLogin struct {
		Email    string `json:"email" validate:"required_with=Password"`
		Password string `json:"password" validate:"required_with=Email"`
	}
	var pLogin PlayerLogin

	// Bind the request with the player
	if err := c.BindJSON(&pLogin); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	// Try to login
	ctx, client := getSpannerConnection(c)
	playerUUID, err := models.PlayerLogin(ctx, client, pLogin.Email, pLogin.Password)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "player not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, playerUUID)
}

// playerLogout responds to the PUT /players/logout endpoint
// Return an empty response with a 200 code.
func playerLogout(c *gin.Context) {
	var player models.Player

	// Bind the request with the player
	if err := c.BindJSON(&player); err != nil {
		if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
			fmt.Printf("could not abort: %s", err)
		}
		return
	}

	// Try to logout
	ctx, client := getSpannerConnection(c)
	err := player.PlayerLogout(ctx, client)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "could not log player out"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "player logged out"})
}

// createPlayer responds to the POST /players endpoint
// When provided the required fields of player_name, email and password, creates a player.
func createPlayer(c *gin.Context) {
	var player models.Player

	if err := c.BindJSON(&player); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ctx, client := getSpannerConnection(c)
	err := player.AddPlayer(ctx, client)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, player.PlayerUUID)
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

	router.POST("/players", createPlayer)
	router.GET("/players/:id", getPlayerByID)
	router.PUT("/players/login", playerLogin)
	router.PUT("/players/logout", playerLogout)

	if err := router.Run(configuration.Server.URL()); err != nil {
		fmt.Printf("could not run gin router: %s", err)
		return
	}
}
