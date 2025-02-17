package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/Kopleman/gophermart/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBalanceService_GetUserBalanceDTO(t *testing.T) {
	logger := log.New()
	cfg := &config.Config{}
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful get balance", func(t *testing.T) {
		repo := new(mocks.BalanceRepo)
		expectedBalance := decimal.NewFromFloat(100.5)
		expectedWithdrawals := decimal.NewFromFloat(30.2)

		repo.On("GetUserBalance", ctx, userID).Return(expectedBalance, nil)
		repo.On("GetUserWithdrawalsSum", ctx, userID).Return(expectedWithdrawals, nil)

		service := NewBalanceService(logger, cfg, repo)
		result, err := service.GetUserBalanceDTO(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, 100.5, result.Current)
		assert.Equal(t, 30.2, result.Withdrawn)
		repo.AssertExpectations(t)
	})

	t.Run("error getting balance", func(t *testing.T) {
		repo := new(mocks.BalanceRepo)
		expectedErr := errors.New("database error")
		repo.On("GetUserBalance", ctx, userID).Return(decimal.Zero, expectedErr)

		service := NewBalanceService(logger, cfg, repo)
		_, err := service.GetUserBalanceDTO(ctx, userID)

		assert.ErrorContains(t, err, "cant get user balance")
		repo.AssertExpectations(t)
	})

	t.Run("error getting withdrawals", func(t *testing.T) {
		repo := new(mocks.BalanceRepo)
		expectedErr := errors.New("database error")
		repo.On("GetUserBalance", ctx, userID).Return(decimal.Zero, nil)
		repo.On("GetUserWithdrawalsSum", ctx, userID).Return(decimal.Zero, expectedErr)

		service := NewBalanceService(logger, cfg, repo)
		_, err := service.GetUserBalanceDTO(ctx, userID)

		assert.ErrorContains(t, err, "cant get user withdrawals")
		repo.AssertExpectations(t)
	})

	t.Run("zero values handling", func(t *testing.T) {
		repo := new(mocks.BalanceRepo)
		repo.On("GetUserBalance", ctx, userID).Return(decimal.Zero, nil)
		repo.On("GetUserWithdrawalsSum", ctx, userID).Return(decimal.Zero, nil)

		service := NewBalanceService(logger, cfg, repo)
		result, err := service.GetUserBalanceDTO(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, 0.0, result.Current)
		assert.Equal(t, 0.0, result.Withdrawn)
		repo.AssertExpectations(t)
	})
}

func TestBalanceService_MakeWithdraw(t *testing.T) {
	logger := log.New()
	cfg := &config.Config{}
	ctx := context.Background()
	userID := uuid.New()
	withdrawDTO := &dto.WithdrawDTO{
		UserID: userID,
		Order:  "123456",
		Amount: decimal.NewFromFloat(10.0),
	}

	t.Run("successful withdraw", func(t *testing.T) {
		repo := new(mocks.BalanceRepo)
		repo.On("MakeWithdraw", ctx, withdrawDTO).Return(nil)

		service := NewBalanceService(logger, cfg, repo)
		err := service.MakeWithdraw(ctx, withdrawDTO)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("not enough balance", func(t *testing.T) {
		repo := new(mocks.BalanceRepo)
		repo.On("MakeWithdraw", ctx, withdrawDTO).Return(pgxstore.ErrNotEnoughBalance)

		service := NewBalanceService(logger, cfg, repo)
		err := service.MakeWithdraw(ctx, withdrawDTO)

		assert.ErrorIs(t, err, ErrNotEnoughBalance)
		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(mocks.BalanceRepo)
		expectedErr := errors.New("database error")
		repo.On("MakeWithdraw", ctx, withdrawDTO).Return(expectedErr)

		service := NewBalanceService(logger, cfg, repo)
		err := service.MakeWithdraw(ctx, withdrawDTO)

		assert.ErrorContains(t, err, "cant make withdraw")
		repo.AssertExpectations(t)
	})

	t.Run("invalid withdraw amount", func(t *testing.T) {
		repo := new(mocks.BalanceRepo)
		invalidDTO := &dto.WithdrawDTO{
			UserID: userID,
			Order:  "123456",
			Amount: decimal.NewFromFloat(-10.0),
		}

		service := NewBalanceService(logger, cfg, repo)
		err := service.MakeWithdraw(ctx, invalidDTO)

		assert.ErrorContains(t, err, "amount must be greater than zero")
	})

	t.Run("nil dto", func(t *testing.T) {
		repo := new(mocks.BalanceRepo)

		service := NewBalanceService(logger, cfg, repo)
		err := service.MakeWithdraw(ctx, nil)

		assert.ErrorContains(t, err, "dto is nil")
	})
}
