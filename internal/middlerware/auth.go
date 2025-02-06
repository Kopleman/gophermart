package middlerware

import (
	"errors"

	"github.com/Kopleman/gophermart/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	jwtware "github.com/gofiber/contrib/jwt"
)

func NewAuthMiddleWare(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return jwtware.New(jwtware.Config{
			SigningKey: jwtware.SigningKey{Key: []byte(cfg.JWTSecret)},
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				return c.SendStatus(fiber.StatusUnauthorized)
			},
		})(c)
	}
}

func GetUserId(ctx *fiber.Ctx) (uuid.UUID, error) {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userIdString := claims["userId"].(string)
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		return uuid.Nil, errors.New("middleware.GetUserId: cannot convert userId from locals to uuid.UUID")
	}
	return userId, nil
}
