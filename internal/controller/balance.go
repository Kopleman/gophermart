package controller

import (
	"context"
	"errors"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/common/utils"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/middlerware"
	"github.com/Kopleman/gophermart/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/gofiber/fiber/v2"
)

type BalanceService interface {
	GetUserBalanceDTO(ctx context.Context, userID uuid.UUID) (*dto.BalanceDTO, error)
	MakeWithdraw(ctx context.Context, requestDTO *dto.WithdrawDTO) error
}

type BalanceController struct {
	logger         log.Logger
	validator      *validator.Validate
	cfg            *config.Config
	balanceService BalanceService
}

func NewBalanceController(
	logger log.Logger,
	validatorInstance *validator.Validate,
	cfg *config.Config,
	balanceService BalanceService,
) *BalanceController {
	return &BalanceController{logger, validatorInstance, cfg, balanceService}
}

// GetUserBalance Fetch user's balance
//
//	@Summary		Fetch user's balance
//	@Description	Fetch user's balance
//	@Tags			user
//	@Accept			plain
//	@Produce		json
//	@Param			Authorization	header		string	true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Success		200				{object}	dto.BalanceDTO
//	@Failure		401				"Unauthorized"
//	@Failure		500				"Internal Server Error"
//	@Router			/api/user/balance [get]
func (b *BalanceController) GetUserBalance() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, err := middlerware.GetUserID(ctx)
		if err != nil {
			b.logger.Errorf("GetUserBalance get userID: %w", err)
			return fiber.ErrUnauthorized
		}

		balance, err := b.balanceService.GetUserBalanceDTO(ctx.Context(), userID)
		if err != nil {
			b.logger.Errorf("GetUserBalance get balance: %w", err)
			return fiber.ErrInternalServerError
		}

		return ctx.JSON(balance)
	}
}

// MakeWithdraw Make withdraw from users balance
//
//	@Summary		Make withdraw from users balance
//	@Description	Make withdraw from users balance
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string					true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Param			data			body	dto.WithdrawRequestDTO	true	"Body params"
//	@Success		200				"OK"
//	@Failure		400				"Bad request"
//	@Failure		401				"Unauthorized"
//	@Failure		402				"Payment Required"
//	@Failure		422				"Unprocessable Entity"
//	@Failure		500				"Internal Server Error"
//	@Router			/api/user/balance/withdraw [post]
func (b *BalanceController) MakeWithdraw() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, err := middlerware.GetUserID(ctx)
		if err != nil {
			b.logger.Errorf("MakeWithdraw get userID: %w", err)
			return fiber.ErrUnauthorized
		}

		data := new(dto.WithdrawRequestDTO)
		if parseErr := ctx.BodyParser(data); parseErr != nil {
			b.logger.Errorf("MakeWithdraw body parse error: %v", parseErr)
			return fiber.ErrUnprocessableEntity
		}
		if validateErr := b.validator.Struct(data); validateErr != nil {
			return fiber.ErrUnprocessableEntity
		}
		if !utils.IsValidOrderNumber(data.Order) {
			return fiber.ErrUnprocessableEntity
		}

		amount := decimal.NewFromFloat(data.Sum)

		withdrawDTO := &dto.WithdrawDTO{
			UserID: userID,
			Order:  data.Order,
			Amount: amount,
		}
		if withdrawErr := b.balanceService.MakeWithdraw(ctx.Context(), withdrawDTO); withdrawErr != nil {
			if errors.Is(withdrawErr, service.ErrNotEnoughBalance) {
				return fiber.ErrPaymentRequired
			}
			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(fiber.StatusOK)
	}
}
