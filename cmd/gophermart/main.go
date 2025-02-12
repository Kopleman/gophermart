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
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	logger := log.New(
		log.WithAppVersion("local"),
		log.WithLogLevel(log.INFO),
	)
	defer logger.Sync() //nolint:all // its safe

	srv := run(ctx, logger, cancel)

	// Wait system signals or context done
	for {
		select {
		case <-sig:
			cancel()
		case <-ctx.Done():
			srv.Shutdown()
			return
		}
	}
}

func run(ctx context.Context, logger log.Logger, cancel context.CancelFunc) *server.Server {
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
			cancel()
		}
		serverRunTimeError := <-runTimeError
		if serverRunTimeError != nil {
			logger.Fatalf("server runtime error: %v", serverRunTimeError)
			cancel()
		}
	}(ctx)

	return srv
}
