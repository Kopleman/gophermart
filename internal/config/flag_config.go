package config

import (
	"flag"

	"github.com/Kopleman/gophermart/internal/common/flags"
)

type flagConfig struct {
	EndPoint        *flags.NetAddress
	AccrualEndPoint *flags.NetAddress
	DataBaseURI     string
	JWTSecret       string
}

func getFlagConfig() *flagConfig {
	config := new(flagConfig)
	endpoint := new(flags.NetAddress)
	endpoint.Host = ""
	endpoint.Port = ""
	config.EndPoint = endpoint
	accrualEndpoint := new(flags.NetAddress)
	accrualEndpoint.Host = ""
	accrualEndpoint.Port = ""
	config.AccrualEndPoint = accrualEndpoint

	endpointValue := flag.Value(endpoint)
	flag.Var(endpointValue, "a", "address and port to run server")

	accrualEndpointValue := flag.Value(endpoint)
	flag.Var(accrualEndpointValue, "r", "address and port to for accrual endpoint")

	flag.StringVar(&config.DataBaseURI, "d", "", "database DSN")

	flag.StringVar(&config.JWTSecret, "j", "", "JWT secret phrase")

	flag.Parse()

	return config
}
