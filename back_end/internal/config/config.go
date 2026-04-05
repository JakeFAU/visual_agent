package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	ProjectID   string `mapstructure:"google_cloud_project"`
	Location    string `mapstructure:"google_cloud_location"`
	RuntimeType string `mapstructure:"runtime_type"` // "vertex" or "local"
	ServerAddr  string `mapstructure:"server_addr"`
}

// Load reads Visual Agent configuration from environment variables.
//
// Supported variables use the VISUAL_AGENT_ prefix for app-owned settings, for
// example VISUAL_AGENT_RUNTIME_TYPE and VISUAL_AGENT_SERVER_ADDR.
func Load() (*Config, error) {
	viper.SetEnvPrefix("VISUAL_AGENT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("google_cloud_location", "us-central1")
	viper.SetDefault("runtime_type", "local")
	viper.SetDefault("server_addr", "127.0.0.1:8080")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
