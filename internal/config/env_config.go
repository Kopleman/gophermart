package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type configFromEnv struct {
	EndPoint        string `env:"RUN_ADDRESS"`
	AccrualEndPoint string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DataBaseURI     string `env:"DATABASE_URI"`
	JWTSecret       string `env:"JWT_SECRET"`
}

func getConfigFromEnv() (*configFromEnv, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfgFromEnv := new(configFromEnv)
	if err := env.Parse(cfgFromEnv); err != nil {
		return nil, fmt.Errorf("failed to parse envs: %w", err)
	}

	return cfgFromEnv, nil
}
