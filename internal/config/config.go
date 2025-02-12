package config

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var defaultJWTSecret = "secret_key" // TODO replace with rand string?
const defaultPollInterval int64 = 2
const defaultWorkerLimit int64 = 10
const defaultMaxOrdersInWork int64 = 10

type Config struct {
	EndPoint        string
	AccrualEndPoint string
	DataBaseURI     string
	JWTSecret       string
	PollInterval    int64
	WorkerLimit     int64
	MaxOrdersInWork int64
}

func (c *Config) Validate() error {
	return validation.ValidateStruct( //nolint:all // self explanatory.
		c,
		validation.Field(&c.EndPoint, validation.Required),
		validation.Field(&c.AccrualEndPoint, validation.Required),
		validation.Field(&c.DataBaseURI, validation.Required),
		validation.Field(&c.JWTSecret, validation.Required),
		validation.Field(&c.PollInterval, validation.Required),
		validation.Field(&c.WorkerLimit, validation.Required),
		validation.Field(&c.MaxOrdersInWork, validation.Required),
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
	config.JWTSecret = cfgFromEnv.JWTSecret
	config.PollInterval = cfgFromEnv.PollInterval
	config.MaxOrdersInWork = cfgFromEnv.MaxOrdersInWork
	config.WorkerLimit = cfgFromEnv.WorkerLimit

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

	if config.WorkerLimit == 0 {
		config.WorkerLimit = defaultWorkerLimit
	}

	if config.PollInterval == 0 {
		config.PollInterval = defaultPollInterval
	}

	if config.MaxOrdersInWork == 0 {
		config.MaxOrdersInWork = defaultMaxOrdersInWork
	}

	if err = config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}
