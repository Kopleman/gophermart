package repo

import (
	"context"
	"fmt"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BalanceRepo struct {
	store  *pgxstore.PGXStore
	logger log.Logger
}

func NewBalanceRepo(l log.Logger, pgxStore *pgxstore.PGXStore) *BalanceRepo {
	return &BalanceRepo{
		logger: l,
		store:  pgxStore,
	}
}

func (b *BalanceRepo) GetUserBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	tx, err := b.store.GetLastUserTransaction(ctx, userID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get user balance: %w", err)
	}

	return tx.NewBalance, nil
}

func (b *BalanceRepo) GetUserWithdrawalsSum(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	sum, err := b.store.GetUserWithdrawalsSum(ctx, userID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get user withdrawals: %w", err)
	}

	return sum, nil
}

func (b *BalanceRepo) MakeWithdraw(ctx context.Context, dto *dto.WithdrawDTO) error {
	if err := b.store.MakeWithdraw(ctx, pgxstore.MakeWithdrawParams{
		UserID:      dto.UserID,
		OrderNumber: dto.Order,
		Amount:      dto.Amount,
	}); err != nil {
		return fmt.Errorf("make withdraw: %w", err)
	}
	return nil
}
