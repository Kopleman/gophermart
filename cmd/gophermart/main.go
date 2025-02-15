package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	logger := log.New(
		log.WithAppVersion("local"),
		log.WithLogLevel(log.Info),
	)
	defer logger.Sync() //nolint:all // its safe

	onErrChan := make(chan error)
	defer close(onErrChan)
	srv := run(ctx, logger, onErrChan)

	// Wait system context done or onError
	for {
		select {
		case err := <-onErrChan:
			if err != nil {
				logger.Fatal(err)
			}
			cancel()
		case <-ctx.Done():
			srv.Shutdown()
			return
		}
	}
}

func run(ctx context.Context, logger log.Logger, onErrorChan chan<- error) *server.Server {
	srvConfig, err := config.GetServerConfig()
	if err != nil {
		logger.Fatalf("failed to parse config for server: %w", err)
	}

	srv := server.NewServer(logger, srvConfig)

	// Start server
	go func(ctx context.Context) {
		runTimeError := make(chan error, 1)
		defer close(runTimeError)

		if serverStartError := srv.Start(ctx, runTimeError); serverStartError != nil {
			onErrorChan <- fmt.Errorf("failed to start server: %w", serverStartError)
		}
		serverRunTimeError := <-runTimeError
		if serverRunTimeError != nil {
			onErrorChan <- fmt.Errorf("runtime error: %w", serverRunTimeError)
		}
	}(ctx)

	return srv
}
