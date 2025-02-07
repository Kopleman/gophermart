package controller

import (
	"context"
	"errors"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/service"
	"github.com/go-playground/validator/v10"

	"github.com/gofiber/fiber/v2"
)

type UserService interface {
	CreateUser(ctx context.Context, createDto *dto.CreateUserRequestDTO) error
	AuthorizeUser(ctx context.Context, loginDto *dto.UserLoginRequestDTO) (string, error)
}

type UserController struct {
	logger      log.Logger
	validator   *validator.Validate
	cfg         *config.Config
	userService UserService
}

func NewUserController(
	logger log.Logger,
	validatorInstance *validator.Validate,
	cfg *config.Config,
	userService UserService,
) *UserController {
	return &UserController{logger, validatorInstance, cfg, userService}
}

type StatusResponseDto struct {
	Status string `json:"status"`
}

// RegisterNewUser register new user
//
//	@Summary		Register new user
//	@Description	Register new user
//	@Tags			auth
//	@Accept			json
//	@Produce		plain
//	@Param			data	body	dto.CreateUserRequestDTO	true	"Body params"
//	@Success		200		"OK"
//	@Failure		400		"Bad request"
//	@Failure		409		"Conflict"
//	@Failure		500		"Internal Server Error"
//	@Router			/api/user/register [post]
func (c *UserController) RegisterNewUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		data := new(dto.CreateUserRequestDTO)
		if err := ctx.BodyParser(data); err != nil {
			return fiber.ErrBadRequest
		}
		if err := c.validator.Struct(data); err != nil {
			return fiber.ErrBadRequest
		}

		if createError := c.userService.CreateUser(ctx.Context(), data); createError != nil {
			if errors.Is(createError, service.ErrAlreadyExists) {
				return fiber.ErrConflict
			}
			c.logger.Errorf("register new user error: %w", createError)
			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(fiber.StatusOK)
	}
}

type LoginResponseDto struct {
	Token string `json:"token" example:"some_token"`
}

// LoginUser login user
//
//	@Summary		Performs user login
//	@Description	Performs user login, returns jwt token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			data	body		dto.UserLoginRequestDTO	true	"Body params"
//	@Success		200		{object}	LoginResponseDto		"OK"
//	@Failure		400		"Bad request"
//	@Failure		401		"Unauthorized"
//	@Failure		500		"Internal Server Error"
//	@Router			/api/user/login [post]
func (c *UserController) LoginUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		data := new(dto.UserLoginRequestDTO)
		if err := ctx.BodyParser(data); err != nil {
			return fiber.ErrBadRequest
		}
		if err := c.validator.Struct(data); err != nil {
			return fiber.ErrBadRequest
		}

		token, err := c.userService.AuthorizeUser(ctx.Context(), data)
		if err != nil {
			if errors.Is(err, service.ErrInvalidArguments) || errors.Is(err, service.ErrNotFound) {
				return fiber.ErrUnauthorized
			}
			return fiber.ErrInternalServerError
		}
		return ctx.JSON(LoginResponseDto{
			Token: token,
		})
	}
}
