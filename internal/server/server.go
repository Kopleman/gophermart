package server

import (
	"context"
	"fmt"

	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/controller"
	"github.com/Kopleman/gophermart/internal/middlerware"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/Kopleman/gophermart/internal/postgres"
	"github.com/Kopleman/gophermart/internal/repo"
	"github.com/Kopleman/gophermart/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	logger   log.Logger
	config   *config.Config
	db       *postgres.PostgreSQL
	pgxStore *pgxstore.PGXStore
	app      *fiber.App
	repos    *repo.Repos
}

func NewServer(logger log.Logger, cfg *config.Config) *Server {
	s := &Server{
		logger: logger,
		config: cfg,
	}

	return s
}

func (s *Server) prepareStore(ctx context.Context) error {
	if err := postgres.RunMigrations(s.config.DataBaseURI); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	pg, err := postgres.NewPostgresSQL(ctx, s.logger, s.config.DataBaseURI)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	s.db = pg
	s.pgxStore = pgxstore.NewPGXStore(s.logger, s.db)
	return nil
}

func (s *Server) Start(ctx context.Context) error {
	defer s.Shutdown()
	if err := s.prepareStore(ctx); err != nil {
		return fmt.Errorf("failed to prepare store: %w", err)
	}

	s.repos = repo.NewRepository(s.logger, s.pgxStore)
	validatorInstance := validator.New()

	userService := service.NewUserService(s.logger, s.config, s.repos.User())
	orderService := service.NewOrderService(s.logger, s.config, s.repos.Order())
	balanceService := service.NewBalanceService(s.logger, s.config, s.repos.Balance())

	userController := controller.NewUserController(s.logger, validatorInstance, s.config, userService)
	orderController := controller.NewOrderController(s.logger, validatorInstance, s.config, orderService)
	balanceController := controller.NewBalanceController(s.logger, validatorInstance, s.config, balanceService)

	runTimeError := make(chan error, 1)

	app := fiber.New()
	app.Use(fiberLogger.New())

	s.app = app

	s.applyRoutes(
		middlerware.NewAuthMiddleWare(s.config),
		userController,
		orderController,
		balanceController,
	)

	go func() {
		if listenAndServeErr := s.app.Listen(s.config.EndPoint); listenAndServeErr != nil {
			runTimeError <- fmt.Errorf("internal server error: %w", listenAndServeErr)
		}
	}()
	s.logger.Infof("Server started on: %s", s.config.EndPoint)

	serverError := <-runTimeError
	if serverError != nil {
		return fmt.Errorf("server error: %w", serverError)
	}

	<-ctx.Done()

	return nil
}

func (s *Server) Shutdown() {
	if s.db != nil {
		s.db.Close()
	}

	s.logger.Infof("Server shut down")
}
