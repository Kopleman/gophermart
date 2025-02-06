package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/pgxstore"
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
	order, orderToProcess, err := r.store.CreateNewOrder(ctx, createDTO)
	if err != nil {
		return nil, nil, fmt.Errorf("create order: %w", err)
	}
	return order, orderToProcess, nil
}
