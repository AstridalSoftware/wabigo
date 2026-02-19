package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	API_BASE_URL string        `envconfig:"API_BASE_URL" required:"true"`
	AppEnv      string        `envconfig:"APP_ENV" default:"development"`
	Debug       bool          `envconfig:"DEBUG" default:"false"`
}

func Load() (*Config, error) {
	godotenv.Load()
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
