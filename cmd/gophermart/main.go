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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	logger := log.New(
		log.WithAppVersion("local"),
		log.WithLogLevel(log.INFO),
	)
	defer logger.Sync() //nolint:all // its safe

	onErrChan := make(chan struct{})
	defer close(onErrChan)
	srv := run(ctx, logger, onErrChan)

	// Wait system context done or onError
	for {
		select {
		case <-onErrChan:
			cancel()
		case <-ctx.Done():
			srv.Shutdown()
			return
		}
	}
}

func run(ctx context.Context, logger log.Logger, onErrorChan chan<- struct{}) *server.Server {
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
			logger.Fatalf("server startup error: %v", serverStartError)
			onErrorChan <- struct{}{}
		}
		serverRunTimeError := <-runTimeError
		if serverRunTimeError != nil {
			logger.Fatalf("server runtime error: %v", serverRunTimeError)
			onErrorChan <- struct{}{}
		}
	}(ctx)

	return srv
}
