package config

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	WASENDER_API_BASE_URL     string `envconfig:"WASENDER_API_BASE_URL" required:"true"`
	DNET_SOPORTE_API_BASE_URL string `envconfig:"DNET_SOPORTE_API_BASE_URL" required:"true"`
	APP_ENV                   string `envconfig:"APP_ENV" default:"development"`
	DEBUG                     bool   `envconfig:"DEBUG" default:"false"`
	API_KEY                   string `envconfig:"API_KEY" required:"true"`
	WASENDER_API_KEY          string `envconfig:"WASENDER_API_KEY" required:"true"`
	WEBHOOK_SECRET            string `envconfig:"WEBHOOK_SECRET" required:"true"`
}

var (
	cfg  Config
	once sync.Once
	err  error
)

func Load() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil && os.Getenv("APP_ENV") == "local" {
			panic(err)
		}
		if err := envconfig.Process("", &cfg); err != nil {
			panic(err)
		}
	})
	return &cfg
}
