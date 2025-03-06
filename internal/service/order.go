package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/google/uuid"
)

type OrderRepo interface {
	GetOrderByNumber(ctx context.Context, orderNumber string) (*pgxstore.Order, error)
	CreateOrder(ctx context.Context, createDTO *dto.CreateOrderDTO) (*pgxstore.Order, *pgxstore.OrdersToProcess, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*pgxstore.Order, error)
}

type OrderService struct {
	logger    log.Logger
	cfg       *config.Config
	orderRepo OrderRepo
}

func NewOrderService(
	logger log.Logger,
	cfg *config.Config,
	orderRepo OrderRepo,
) *OrderService {
	return &OrderService{
		logger,
		cfg,
		orderRepo,
	}
}

func (os *OrderService) GetOrderByNumber(ctx context.Context, orderNumber string) (*dto.OrderDTO, error) {
	order, err := os.orderRepo.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, pgxstore.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error getting order by number %s: %w", orderNumber, err)
	}

	return order.ToDTO(), nil
}

func (os *OrderService) CreateOrder(ctx context.Context, createDTO *dto.CreateOrderDTO) error {
	_, _, err := os.orderRepo.CreateOrder(ctx, createDTO)
	if err != nil {
		return fmt.Errorf("error creating order %s: %w", createDTO.OrderNumber, err)
	}
	return nil
}

func (os *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*dto.OrderDTO, error) {
	orders, err := os.orderRepo.GetUserOrders(ctx, userID)
	if err != nil {
		if errors.Is(err, pgxstore.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user %s orders: %w", userID, err)
	}
	dtos := make([]*dto.OrderDTO, len(orders))
	for i, order := range orders {
		dtos[i] = order.ToDTO()
	}
	return dtos, nil
}
