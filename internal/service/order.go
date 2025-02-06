package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/pgxstore"
)

type OrderRepo interface {
	GetOrderByNumber(ctx context.Context, orderNumber string) (*pgxstore.Order, error)
	CreateOrder(ctx context.Context, createDTO *dto.CreateOrderDTO) (*pgxstore.Order, *pgxstore.OrdersToProcess, error)
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
