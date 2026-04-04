package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	ProjectID string `mapstructure:"google_cloud_project"`
	Location  string `mapstructure:"google_cloud_location"`
}

func Load() (*Config, error) {
	viper.SetEnvPrefix("VISUAL_AGENT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("google_cloud_location", "us-central1")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
