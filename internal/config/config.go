package config

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var defaultJWTSecret = "secret_key" // TODO replace with rand string?

type Config struct {
	EndPoint        string
	AccrualEndPoint string
	DataBaseURI     string
	JWTSecret       string
}

func (c *Config) Validate() error {
	return validation.ValidateStruct( //nolint:all // self explanatory.
		c,
		validation.Field(&c.EndPoint, validation.Required),
		validation.Field(&c.AccrualEndPoint, validation.Required),
		validation.Field(&c.DataBaseURI, validation.Required),
		validation.Field(&c.JWTSecret, validation.Required),
	)
}

func GetServerConfig() (*Config, error) {
	cfgFromEnv, err := getConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("config parse: %w", err)
	}
	config := new(Config)
	config.DataBaseURI = cfgFromEnv.DataBaseURI
	config.EndPoint = cfgFromEnv.EndPoint
	config.AccrualEndPoint = cfgFromEnv.AccrualEndPoint

	configFromFlags := getFlagConfig()

	if configFromFlags.EndPoint.String() != "" {
		config.EndPoint = configFromFlags.EndPoint.String()
	}

	if configFromFlags.AccrualEndPoint.String() != "" {
		config.AccrualEndPoint = configFromFlags.AccrualEndPoint.String()
	}

	if configFromFlags.DataBaseURI != "" {
		config.DataBaseURI = configFromFlags.DataBaseURI
	}

	if configFromFlags.JWTSecret != "" {
		config.JWTSecret = configFromFlags.JWTSecret
	}

	if config.JWTSecret == "" {
		config.JWTSecret = defaultJWTSecret
	}

	if err = config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}
