package config

import (
	"sync"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	WASENDER_API_BASE_URL     string `envconfig:"WASENDER_API_BASE_URL" required:"true"`
	DNET_SOPORTE_API_BASE_URL string `envconfig:"DNET_SOPORTE_API_BASE_URL" required:"true"`
	AppEnv                    string `envconfig:"APP_ENV" default:"development"`
	Debug                     bool   `envconfig:"DEBUG" default:"false"`
	APIKey                    string `envconfig:"API_KEY" required:"true"`
}

var (
	cfg  Config
	once sync.Once
	err  error
)

func Load() *Config {
	once.Do(func() {
		err = godotenv.Load()
		err = envconfig.Process("", &cfg)
		panic(err)
	})
	return &cfg
}
