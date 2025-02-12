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
	"github.com/shopspring/decimal"
)

type BalanceRepo interface {
	GetUserBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	GetUserWithdrawalsSum(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	MakeWithdraw(ctx context.Context, dto *dto.WithdrawDTO) error
}

type BalanceService struct {
	logger      log.Logger
	cfg         *config.Config
	balanceRepo BalanceRepo
}

func NewBalanceService(
	logger log.Logger,
	cfg *config.Config,
	balanceRepo BalanceRepo,
) *BalanceService {
	return &BalanceService{
		logger,
		cfg,
		balanceRepo,
	}
}

func (b *BalanceService) GetUserBalanceDTO(ctx context.Context, userID uuid.UUID) (*dto.BalanceDTO, error) {
	balance, err := b.balanceRepo.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserBalanceDTO cant get user balance: %w", err)
	}
	withdrawals, err := b.balanceRepo.GetUserWithdrawalsSum(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserBalanceDTO cant get user withdrawals: %w", err)
	}

	balanceValue, _ := balance.Float64()
	withdrawalsValue, _ := withdrawals.Float64()
	return &dto.BalanceDTO{
		Current:   balanceValue,
		Withdrawn: withdrawalsValue,
	}, nil
}

func (b *BalanceService) MakeWithdraw(ctx context.Context, dto *dto.WithdrawDTO) error {
	err := b.balanceRepo.MakeWithdraw(ctx, dto)
	if err != nil {
		if errors.Is(err, pgxstore.ErrNotEnoughBalance) {
			return ErrNotEnoughBalance
		}
		return fmt.Errorf("MakeWithdraw cant make withdraw: %w", err)
	}

	return nil
}
