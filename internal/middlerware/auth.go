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

func GetUserID(ctx *fiber.Ctx) (uuid.UUID, error) {
	user, ok := ctx.Locals("user").(*jwt.Token)
	if !ok {
		return uuid.Nil, errors.New("middleware.GetUserID: cannot get user from context")
	}
	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("middleware.GetUserID: cannot convert users claims")
	}
	userIDString, ok := claims["userID"].(string)
	if !ok {
		return uuid.Nil, errors.New("middleware.GetUserID: cannot convert userID to string")
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, errors.New("middleware.GetUserID: cannot convert userID from locals to uuid.UUID")
	}
	return userID, nil
}
