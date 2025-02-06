package server

import (
	"github.com/Kopleman/gophermart/docs"
	"github.com/Kopleman/gophermart/internal/controller"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func (s *Server) applyRoutes(
	authMiddleware fiber.Handler,
	userController *controller.UserController,
	orderController *controller.OrderController,
) {
	docs.SwaggerInfo.Host = s.config.EndPoint

	apiRouter := s.app.Group("/api")

	apiRouter.Get("/api-docs/*", swagger.HandlerDefault)

	userGroup := apiRouter.Group("/user")
	userGroup.Post("/register", userController.RegisterNewUser())
	userGroup.Post("/login", userController.LoginUser())
	orderGroup := userGroup.Group("/", authMiddleware)
	orderGroup.Post("/orders", orderController.AddOrder())
	orderGroup.Get("/orders", orderController.GetOrders())

	s.app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotFound) // => 404 "Not Found"
	})
}
