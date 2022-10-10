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

package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Configurations exported
type Config struct {
	Server  ServerConfig
	Spanner SpannerConfig
}

// ServerConfigurations exported
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfigurations exported
type SpannerConfig struct {
	Project_id      string `mapstructure:"PROJECT_ID" yaml:"project_id,omitempty"`
	Instance_id     string `mapstructure:"INSTANCE_ID" yaml:"instance_id,omitempty"`
	Database_id     string `mapstructure:"DATABASE_ID" yaml:"database_id,omitempty"`
	CredentialsFile string `mapstructure:"CREDENTIALS_FILE" yaml:"credentials_file,omitempty"`
}

func NewConfig() (Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yml")

	viper.AutomaticEnv()

	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)

	// Bind environment variable override
	if err := viper.BindEnv("server.host", "SERVICE_HOST"); err != nil {
		return Config{}, fmt.Errorf("could not set environment variable 'server.host': %s", err)
	}
	if err := viper.BindEnv("server.port", "SERVICE_PORT"); err != nil {
		return Config{}, fmt.Errorf("could not set environment variable 'server.port': %s", err)
	}
	if err := viper.BindEnv("spanner.project_id", "SPANNER_PROJECT_ID"); err != nil {
		return Config{}, fmt.Errorf("could not set environment variable 'spanner.project_id': %s", err)
	}
	if err := viper.BindEnv("spanner.instance_id", "SPANNER_INSTANCE_ID"); err != nil {
		return Config{}, fmt.Errorf("could not set environment variable 'spanner.instance_id': %s", err)
	}
	if err := viper.BindEnv("spanner.database_id", "SPANNER_DATABASE_ID"); err != nil {
		return Config{}, fmt.Errorf("could not set environment variable 'spanner.database_id': %s", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("[WARNING] could not read config %s\n", err.Error())
	}

	var c Config

	if err := viper.Unmarshal(&c); err != nil {
		return Config{}, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return c, nil
}

func (c *SpannerConfig) DB() string {
	return fmt.Sprintf(
		"projects/%s/instances/%s/databases/%s",
		c.Project_id,
		c.Instance_id,
		c.Database_id,
	)
}

func (c *ServerConfig) URL() string {
	return fmt.Sprintf(
		"%s:%d",
		c.Host,
		c.Port,
	)
}
