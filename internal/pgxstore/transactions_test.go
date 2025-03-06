package pgxstore

import (
	"context"
	"testing"
	"time"

	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPGXStore_MakeWithdraw(t *testing.T) {
	t.Run("successful withdraw", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		if err != nil {
			t.Fatal(err)
		}
		defer mock.Close()

		ctx := context.Background()
		store := NewPGXStore(&log.MockLogger{}, mock)

		inputParams := MakeWithdrawParams{
			OrderNumber: "49927398716",
			Amount:      decimal.NewFromFloat(500),
			UserID:      uuid.Nil,
		}
		lastTx := Transaction{
			CreatedAt:  time.Now(),
			OldBalance: decimal.Zero,
			Change:     decimal.NewFromFloat(1000),
			NewBalance: decimal.NewFromFloat(1000),
			Type:       TransactionTypeINIT,
			ID:         uuid.New(),
			UserID:     uuid.Nil,
		}

		newWithdraw := CreateTransactionParams{
			OrderNumber: &inputParams.OrderNumber,
			OldBalance:  lastTx.NewBalance,
			Change:      inputParams.Amount,
			NewBalance:  lastTx.NewBalance.Sub(inputParams.Amount),
			Type:        TransactionTypeWITHDRAW,
			UserID:      lastTx.UserID,
		}

		mock.ExpectBegin()
		rows := pgxmock.
			NewRows([]string{"id", "order_number", "user_id", "old_balance", "change", "new_balance", "type", "created_at"}).
			AddRow(
				lastTx.ID,
				nil,
				lastTx.UserID,
				lastTx.OldBalance.String(),
				lastTx.Change.String(),
				lastTx.NewBalance.String(),
				lastTx.Type,
				lastTx.CreatedAt,
			)
		mock.ExpectQuery(GetLastUserTransaction).
			WithArgs(uuid.Nil).
			WillReturnRows(rows)

		newTXRows := pgxmock.
			NewRows([]string{"id", "order_number", "user_id", "old_balance", "change", "new_balance", "type", "created_at"}).
			AddRow(
				newWithdraw.UserID,
				newWithdraw.OrderNumber,
				newWithdraw.UserID,
				newWithdraw.OldBalance.String(),
				newWithdraw.Change.String(),
				newWithdraw.NewBalance.String(),
				newWithdraw.Type,
				time.Now(),
			)
		mock.ExpectQuery(CreateTransaction).
			WithArgs(
				newWithdraw.OrderNumber,
				newWithdraw.UserID,
				TestDecimal{expected: newWithdraw.OldBalance},
				TestDecimal{expected: newWithdraw.Change},
				TestDecimal{expected: newWithdraw.NewBalance},
				newWithdraw.Type,
			).WillReturnRows(newTXRows)

		mock.ExpectCommit()

		err = store.MakeWithdraw(ctx, inputParams)
		assert.NoError(t, err)
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("not enough balance", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		if err != nil {
			t.Fatal(err)
		}
		defer mock.Close()

		ctx := context.Background()
		store := NewPGXStore(&log.MockLogger{}, mock)

		inputParams := MakeWithdrawParams{
			OrderNumber: "49927398716",
			Amount:      decimal.NewFromFloat(10000),
			UserID:      uuid.Nil,
		}
		lastTx := Transaction{
			CreatedAt:  time.Now(),
			OldBalance: decimal.Zero,
			Change:     decimal.NewFromFloat(1000),
			NewBalance: decimal.NewFromFloat(1000),
			Type:       TransactionTypeINIT,
			ID:         uuid.New(),
			UserID:     uuid.Nil,
		}

		mock.ExpectBegin()
		rows := pgxmock.
			NewRows([]string{"id", "order_number", "user_id", "old_balance", "change", "new_balance", "type", "created_at"}).
			AddRow(
				lastTx.ID,
				nil,
				lastTx.UserID,
				lastTx.OldBalance.String(),
				lastTx.Change.String(),
				lastTx.NewBalance.String(),
				lastTx.Type,
				lastTx.CreatedAt,
			)
		mock.ExpectQuery(GetLastUserTransaction).
			WithArgs(uuid.Nil).
			WillReturnRows(rows)

		mock.ExpectRollback()

		err = store.MakeWithdraw(ctx, inputParams)
		assert.ErrorIs(t, err, ErrNotEnoughBalance)
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
