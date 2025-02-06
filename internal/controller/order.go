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
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OrderService interface {
	GetOrderByNumber(ctx context.Context, orderNumber string) (*dto.OrderDTO, error)
	CreateOrder(ctx context.Context, createDTO *dto.CreateOrderDTO) error
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*dto.OrderDTO, error)
}

type OrderController struct {
	logger       log.Logger
	validator    *validator.Validate
	cfg          *config.Config
	orderService OrderService
}

func NewOrderController(
	logger log.Logger,
	validatorInstance *validator.Validate,
	cfg *config.Config,
	orderService OrderService,
) *OrderController {
	return &OrderController{
		logger:       logger,
		validator:    validatorInstance,
		cfg:          cfg,
		orderService: orderService,
	}
}

// AddOrder Add new user order to system
//
//	@Summary		Add new order
//	@Description	Add new user order to system
//	@Tags			order
//	@Accept			plain
//	@Produce		plain
//	@Param			Authorization	header		string	true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Param			data			body		string	true	"Body params"
//	@Success		200				{string}	string	"OK"
//	@Success		202				{string}	string	"Accepted"
//	@Failure		400				{string}	string	"Bad request"
//	@Failure		401				{string}	string	"Unauthorized"
//	@Failure		409				{string}	string	"invalid order number"
//	@Failure		422				{string}	string	"invalid order number"
//	@Failure		500				{string}	string	"Internal Server Error"
//	@Router			/api/user/orders [post]
func (o *OrderController) AddOrder() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, err := middlerware.GetUserID(ctx)
		if err != nil {
			o.logger.Errorf("AddOrder get userID: %w", err)
			return fiber.ErrUnauthorized
		}

		orderNumber := string(ctx.Body())
		if orderNumber == "" {
			return fiber.ErrBadRequest
		}
		if !utils.IsValidOrderNumber(orderNumber) {
			return fiber.ErrUnprocessableEntity
		}

		order, getOrderErr := o.orderService.GetOrderByNumber(ctx.Context(), orderNumber)
		if getOrderErr != nil {
			if !errors.Is(getOrderErr, service.ErrNotFound) {
				o.logger.Errorf("AddOrder get order: %w", getOrderErr)
				return fiber.ErrInternalServerError
			}
		}

		if order != nil {
			if order.UserID != userID {
				return fiber.ErrConflict
			}

			if order.UserID == userID {
				return ctx.SendStatus(fiber.StatusOK)
			}
		}

		createDto := dto.CreateOrderDTO{
			UserID:      userID,
			OrderNumber: orderNumber,
		}
		if err = o.orderService.CreateOrder(ctx.Context(), &createDto); err != nil {
			o.logger.Errorf("AddOrder create order: %w", err)
			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(fiber.StatusAccepted)
	}
}

// GetOrders Fetch list of users orders
//
//	@Summary		Fetch list of users orders
//	@Description	Fetch list of users orders
//	@Tags			order
//	@Accept			plain
//	@Produce		plain
//	@Param			Authorization	header		string	true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Success		200				{array}		dto.OrderInfoDTO
//	@Success		204				{string}	string	"Accepted"
//	@Failure		401				{string}	string	"Unauthorized"
//	@Failure		500				{string}	string	"Internal Server Error"
//	@Router			/api/user/orders [get]
func (o *OrderController) GetOrders() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, err := middlerware.GetUserID(ctx)
		if err != nil {
			o.logger.Errorf("GetOrders get userID: %w", err)
			return fiber.ErrUnauthorized
		}

		orders, err := o.orderService.GetUserOrders(ctx.Context(), userID)
		if err != nil {
			o.logger.Errorf("GetOrders get orders: %w", err)
			return fiber.ErrInternalServerError
		}

		if len(orders) == 0 {
			return ctx.SendStatus(fiber.StatusNoContent)
		}

		dtos := make([]*dto.OrderInfoDTO, len(orders))
		for i, order := range orders {
			dtos[i] = order.ToInfoDTO()
		}
		return ctx.JSON(dtos)
	}
}
