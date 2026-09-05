package service

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	JWTSecret  string        `envconfig:"JWT_SECRET" required:"true"`
	AccessTTL  time.Duration `envconfig:"ACCESS_TTL" default:"15m"`
	RefreshTTL time.Duration `envconfig:"REFRESH_TTL" default:"720h"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("AUTH", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get auth config: %w", err))
	}

	return config
}
