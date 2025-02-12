package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OrderRepo struct {
	store  *pgxstore.PGXStore
	logger log.Logger
}

func NewOrderRepo(l log.Logger, pgxStore *pgxstore.PGXStore) *OrderRepo {
	return &OrderRepo{
		logger: l,
		store:  pgxStore,
	}
}

func (r *OrderRepo) GetOrderByNumber(ctx context.Context, orderNumber string) (*pgxstore.Order, error) {
	order, err := r.store.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgxstore.ErrNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	return order, nil
}

func (r *OrderRepo) CreateOrder(
	ctx context.Context,
	createDTO *dto.CreateOrderDTO,
) (*pgxstore.Order, *pgxstore.OrdersToProcess, error) {
	order, orderToProcess, err := r.store.CreateNewOrder(ctx, pgxstore.CreateOrderParams{
		UserID:      createDTO.UserID,
		OrderNumber: createDTO.OrderNumber,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create order: %w", err)
	}
	return order, orderToProcess, nil
}

func (r *OrderRepo) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*pgxstore.Order, error) {
	orders, err := r.store.GetUserOrders(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgxstore.ErrNotFound
		}
		return nil, fmt.Errorf("get user orders: %w", err)
	}

	return orders, nil
}

func (r *OrderRepo) PickOrdersToProcess(ctx context.Context, limit int32) ([]*pgxstore.OrdersToProcess, error) {
	orders, err := r.store.PickOrdersToProcess(ctx, limit)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pick orders to process: %w", err)
	}

	return orders, nil
}

func (r *OrderRepo) GetRegisteredProcessingOrders(ctx context.Context, limit int32) ([]*pgxstore.OrdersToProcess, error) {
	orders, err := r.store.GetRegisteredProcessingOrders(ctx, limit)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pick orders to process: %w", err)
	}

	return orders, nil
}

func (r *OrderRepo) GetStartProcessingOrders(ctx context.Context) ([]*pgxstore.OrdersToProcess, error) {
	orders, err := r.store.GetStartProcessingOrders(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get processing orders: %w", err)
	}

	return orders, nil
}

func (r *OrderRepo) RegisterOrderProcessing(ctx context.Context, orderNumber string) error {
	if err := r.store.RegisterOrderProcessing(ctx, orderNumber); err != nil {
		return fmt.Errorf("register order processing: %w", err)
	}

	return nil
}

func (r *OrderRepo) StoreAccrualCalculation(
	ctx context.Context,
	params pgxstore.AccrualCalculationParams,
) error {
	if err := r.store.StoreAccrualCalculation(ctx, params); err != nil {
		return fmt.Errorf("store accrual calculation: %w", err)
	}

	return nil
}
