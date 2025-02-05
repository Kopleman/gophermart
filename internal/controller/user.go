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

//type UserController interface {
//	Register() fiber.Handler
//	Logout() fiber.Handler
//	Authorize() fiber.Handler
//	RefreshAcessToken() fiber.Handler
//}

type UserService interface {
	CreateUser(ctx context.Context, createDto *dto.CreateUserRequestDTO) error
	AuthorizeUser(ctx context.Context, loginDto *dto.UserLoginRequestDTO) (string, error)
}

type UserController struct {
	logger      log.Logger
	validator   *validator.Validate
	config      *config.Config
	userService UserService
}

func NewUserController(logger log.Logger, validator *validator.Validate, config *config.Config, userService UserService) *UserController {
	return &UserController{logger, validator, config, userService}
}

type StatusResponseDto struct {
	Status string `json:"status"`
}

// RegisterNewUser
//
//	@Summary		Register new user
//	@Description	Register new user
//	@Tags			user
//	@Accept			json
//	@Produce		plain
//	@Param			Authorization	header		string						true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Param			data			body		dto.CreateUserRequestDTO	true	"Body params"
//	@Success		200				{string}	string						"OK"
//	@Failure		400				{string}	string						"Bad request"
//	@Failure		409				{string}	string						"User already exists"
//	@Failure		500				{string}	string						"Internal Server Error"
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
				return fiber.NewError(fiber.StatusConflict, "User already exists")
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

// LoginUser
//
//	@Summary		Performs user login
//	@Description	Performs user login, returns jwt token
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			data	body		dto.UserLoginRequestDTO	true	"Body params"
//	@Success		200		{object}	LoginResponseDto		"OK"
//	@Failure		400		{string}	string					"Bad request"
//	@Failure		401		{string}	string					"Unauthorized"
//	@Failure		500		{string}	string					"Internal Server Error"
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
