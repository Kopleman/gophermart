package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/Kopleman/gophermart/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestOrderService_GetOrderByNumber(t *testing.T) {
	logger := log.New()
	cfg := &config.Config{}
	ctx := context.Background()
	orderNumber := "1234567890"
	userID := uuid.New()

	t.Run("successful get order", func(t *testing.T) {
		repo := new(mocks.OrderRepo)
		expectedOrder := &pgxstore.Order{
			ID:          uuid.New(),
			UserID:      userID,
			OrderNumber: orderNumber,
			Status:      "PROCESSED",
			Accrual:     decimal.NewFromFloat(100.5),
			CreatedAt:   time.Now(),
		}
		expectedOrderDto := &dto.OrderDTO{
			OrderNumber: expectedOrder.OrderNumber,
			Status:      expectedOrder.Status.String(),
			CreatedAt:   expectedOrder.CreatedAt.Format(time.RFC3339),
			Accrual:     expectedOrder.Accrual,
			ID:          expectedOrder.ID,
			UserID:      expectedOrder.UserID,
		}
		repo.On("GetOrderByNumber", ctx, orderNumber).Return(expectedOrder, nil)

		service := NewOrderService(logger, cfg, repo)
		result, err := service.GetOrderByNumber(ctx, orderNumber)

		assert.NoError(t, err)
		assert.Equal(t, expectedOrderDto, result)
		repo.AssertExpectations(t)
	})

	t.Run("order not found", func(t *testing.T) {
		repo := new(mocks.OrderRepo)
		repo.On("GetOrderByNumber", ctx, orderNumber).Return(nil, pgxstore.ErrNotFound)

		service := NewOrderService(logger, cfg, repo)
		_, err := service.GetOrderByNumber(ctx, orderNumber)

		assert.ErrorIs(t, err, ErrNotFound)
		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(mocks.OrderRepo)
		expectedErr := errors.New("database error")
		repo.On("GetOrderByNumber", ctx, orderNumber).Return(nil, expectedErr)

		service := NewOrderService(logger, cfg, repo)
		_, err := service.GetOrderByNumber(ctx, orderNumber)

		assert.ErrorContains(t, err, "error getting order by number")
		repo.AssertExpectations(t)
	})
}

func TestOrderService_CreateOrder(t *testing.T) {
	logger := log.New()
	cfg := &config.Config{}
	ctx := context.Background()
	userID := uuid.New()
	orderNumber := "1234567890"

	t.Run("successful order creation", func(t *testing.T) {
		repo := new(mocks.OrderRepo)
		createDTO := &dto.CreateOrderDTO{
			UserID:      userID,
			OrderNumber: orderNumber,
		}
		repo.On("CreateOrder", ctx, createDTO).Return(
			&pgxstore.Order{},
			&pgxstore.OrdersToProcess{},
			nil,
		)

		service := NewOrderService(logger, cfg, repo)
		err := service.CreateOrder(ctx, createDTO)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(mocks.OrderRepo)
		createDTO := &dto.CreateOrderDTO{
			UserID:      userID,
			OrderNumber: orderNumber,
		}
		expectedErr := errors.New("duplicate order")
		repo.On("CreateOrder", ctx, createDTO).Return(nil, nil, expectedErr)

		service := NewOrderService(logger, cfg, repo)
		err := service.CreateOrder(ctx, createDTO)

		assert.ErrorIs(t, err, expectedErr)
		repo.AssertExpectations(t)
	})
}

func TestOrderService_GetUserOrders(t *testing.T) {
	logger := log.New()
	cfg := &config.Config{}
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful get orders", func(t *testing.T) {
		repo := new(mocks.OrderRepo)
		expectedOrders := []*pgxstore.Order{
			{
				ID:          uuid.New(),
				UserID:      userID,
				OrderNumber: "1234567890",
				Status:      "PROCESSED",
				Accrual:     decimal.NewFromFloat(100.5),
				CreatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				UserID:      userID,
				OrderNumber: "1234567890",
				Status:      "49927398716",
				Accrual:     decimal.NewFromFloat(100.5),
				CreatedAt:   time.Now(),
			},
		}
		expectedOrderDtos := []*dto.OrderDTO{
			{
				OrderNumber: expectedOrders[0].OrderNumber,
				Status:      expectedOrders[0].Status.String(),
				CreatedAt:   expectedOrders[0].CreatedAt.Format(time.RFC3339),
				Accrual:     expectedOrders[0].Accrual,
				ID:          expectedOrders[0].ID,
				UserID:      expectedOrders[0].UserID,
			},
			{
				OrderNumber: expectedOrders[1].OrderNumber,
				Status:      expectedOrders[1].Status.String(),
				CreatedAt:   expectedOrders[1].CreatedAt.Format(time.RFC3339),
				Accrual:     expectedOrders[1].Accrual,
				ID:          expectedOrders[1].ID,
				UserID:      expectedOrders[1].UserID,
			},
		}

		repo.On("GetUserOrders", ctx, userID).Return(expectedOrders, nil)

		service := NewOrderService(logger, cfg, repo)
		result, err := service.GetUserOrders(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedOrderDtos[0], result[0])
		assert.Equal(t, expectedOrderDtos[1], result[1])
		repo.AssertExpectations(t)
	})

	t.Run("no orders found", func(t *testing.T) {
		repo := new(mocks.OrderRepo)
		repo.On("GetUserOrders", ctx, userID).Return([]*pgxstore.Order{}, nil)

		service := NewOrderService(logger, cfg, repo)
		result, err := service.GetUserOrders(ctx, userID)

		assert.NoError(t, err)
		assert.Empty(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(mocks.OrderRepo)
		expectedErr := errors.New("database error")
		repo.On("GetUserOrders", ctx, userID).Return(nil, expectedErr)

		service := NewOrderService(logger, cfg, repo)
		_, err := service.GetUserOrders(ctx, userID)

		assert.ErrorContains(t, err, "error getting user")
		repo.AssertExpectations(t)
	})
}
