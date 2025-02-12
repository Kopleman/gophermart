package pgxstore

import (
	"context"
	"fmt"
	"time"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type MakeWithdrawParams struct {
	OrderNumber string          `db:"order_number" json:"order_number"`
	Amount      decimal.Decimal `db:"amount" json:"amount"`
	UserID      uuid.UUID       `db:"user_id" json:"user_id"`
}

func (p *PGXStore) MakeWithdraw(ctx context.Context, params MakeWithdrawParams) error {
	tx, err := p.startTx(ctx, &pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("pgxstore.MakeWithdraw could not start transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:all // its safe

	lastUserTx, lastTxErr := p.GetLastUserTransaction(ctx, params.UserID)
	if lastTxErr != nil {
		return fmt.Errorf("pgxstore.MakeWithdraw could not get last user transaction: %w", lastTxErr)
	}

	if params.Amount.GreaterThanOrEqual(lastUserTx.NewBalance) {
		return ErrNotEnoughBalance
	}
	withDrawParams := CreateTransactionParams{
		OrderNumber: &params.OrderNumber,
		UserID:      params.UserID,
		OldBalance:  lastUserTx.NewBalance,
		Change:      params.Amount,
		NewBalance:  lastUserTx.NewBalance.Sub(params.Amount),
		Type:        TransactionTypeWITHDRAW,
	}

	_, err = p.CreateTransaction(ctx, withDrawParams)
	if err != nil {
		return fmt.Errorf("pgxstore.MakeWithdraw could not create transaction: %w", err)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("pgxstore.MakeWithdraw could not commit transaction: %w", commitErr)
	}

	return nil
}

func (t *Transaction) ToWithdrawalItemDTO() *dto.WithdrawalItemDTO {
	order := ""
	if t.OrderNumber != nil {
		order = *t.OrderNumber
	}
	sum, _ := t.Change.Float64()
	item := &dto.WithdrawalItemDTO{
		Order:       order,
		Sum:         sum,
		ProcessedAt: t.CreatedAt.Format(time.RFC3339),
	}

	return item
}
