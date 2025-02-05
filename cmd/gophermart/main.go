package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/server"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	logger := log.New(
		log.WithAppVersion("local"),
		log.WithLogLevel(log.INFO),
	)
	defer logger.Sync() //nolint:all // its safe

	go run(logger)

	// Wait system signals
	<-sig
}

func run(logger log.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	srvConfig, err := config.GetServerConfig()
	if err != nil {
		logger.Fatalf("failed to parse config for server: %w", err)
	}

	srv := server.NewServer(logger, srvConfig)

	// Start server
	go func(ctx context.Context) {
		if serverStartError := srv.Start(ctx); serverStartError != nil {
			logger.Fatalf("start server error: %v", serverStartError)
		}
		cancel()
	}(ctx)
}
