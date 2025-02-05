package server

import (
	"github.com/Kopleman/gophermart/docs"
	_ "github.com/Kopleman/gophermart/docs"
	"github.com/Kopleman/gophermart/internal/controller"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func (s *Server) applyRoutes(
	userController *controller.UserController,
) {

	docs.SwaggerInfo.Host = s.config.EndPoint

	apiRouter := s.app.Group("/api")

	apiRouter.Get("/api-docs/*", swagger.HandlerDefault)

	userGroup := apiRouter.Group("/user")
	userGroup.Post("/register", userController.RegisterNewUser())

	s.app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404) // => 404 "Not Found"
	})
}
