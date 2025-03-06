package postgres

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	// In seconds. Default 5 sec
	PingInterval         int   `json:"POSTGRES_PING_INTERVAL" default:"5"`
	MaxConns             int32 `json:"POSTGRES_MAX_CONNS"`
	MinConns             int32 `json:"POSTGRES_MIN_CONNS"`
	PreferSimpleProtocol bool  `json:"POSTGRES_PREFER_SIMPLE_PROTOCOL"`
}

func (c Config) Validate() error {
	err := validation.ValidateStruct(
		&c,
		validation.Field(&c.PingInterval, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	return nil
}
