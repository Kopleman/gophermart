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

func TestPGXStore_CreateNewOrder(t *testing.T) {
	t.Run("successful order create", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		if err != nil {
			t.Fatal(err)
		}
		defer mock.Close()

		ctx := context.Background()
		store := NewPGXStore(&log.MockLogger{}, mock)

		inputParams := CreateOrderParams{
			OrderNumber: "49927398716",
			UserID:      uuid.Nil,
		}
		newOrderUUID := uuid.New()

		mock.ExpectBegin()

		orderRows := pgxmock.
			NewRows([]string{"id", "user_id", "order_number", "status", "accrual", "created_at", "updated_at", "deleted_at"}).
			AddRow(
				newOrderUUID,
				inputParams.UserID,
				inputParams.OrderNumber,
				OrderStatusTypeNEW,
				decimal.Zero,
				time.Now(),
				nil,
				nil,
			)
		mock.ExpectQuery(CreateOrder).
			WithArgs(inputParams.UserID, inputParams.OrderNumber).
			WillReturnRows(orderRows)

		processingOrder := pgxmock.
			NewRows([]string{"order_number", "process_status", "created_at", "updated_at", "deleted_at"}).
			AddRow(
				inputParams.OrderNumber,
				ProcessStatusTypeNEW,
				time.Now(),
				nil,
				nil,
			)

		mock.ExpectQuery(PutOrderForProcessing).
			WithArgs(inputParams.OrderNumber).
			WillReturnRows(processingOrder)

		mock.ExpectCommit()

		_, _, err = store.CreateNewOrder(ctx, inputParams)
		assert.NoError(t, err)
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestPGXStore_RegisterOrderProcessing(t *testing.T) {
	t.Run("successful register", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		if err != nil {
			t.Fatal(err)
		}
		defer mock.Close()

		ctx := context.Background()
		store := NewPGXStore(&log.MockLogger{}, mock)
		orderNumber := "49927398716"

		orderUUID := uuid.New()

		mock.ExpectBegin()

		processingOrder := pgxmock.
			NewRows([]string{"order_number", "process_status", "created_at", "updated_at", "deleted_at"}).
			AddRow(
				orderNumber,
				ProcessStatusTypeREGISTERED,
				time.Now(),
				nil,
				nil,
			)

		mock.ExpectQuery(UpdateOrderToProcess).
			WithArgs(ProcessStatusTypeREGISTERED, orderNumber).
			WillReturnRows(processingOrder)

		orderRows := pgxmock.
			NewRows([]string{"id", "user_id", "order_number", "status", "accrual", "created_at", "updated_at", "deleted_at"}).
			AddRow(
				orderUUID,
				uuid.Nil,
				orderNumber,
				OrderStatusTypePROCESSING,
				decimal.Zero,
				time.Now(),
				nil,
				nil,
			)
		mock.ExpectQuery(UpdateOrder).
			WithArgs(OrderStatusTypePROCESSING, TestDecimal{expected: decimal.Zero}, orderNumber).
			WillReturnRows(orderRows)

		mock.ExpectCommit()

		err = store.RegisterOrderProcessing(ctx, orderNumber)
		assert.NoError(t, err)
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestPGXStore_StoreAccrualCalculation(t *testing.T) {
	t.Run("successful data update", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		if err != nil {
			t.Fatal(err)
		}
		defer mock.Close()

		ctx := context.Background()
		store := NewPGXStore(&log.MockLogger{}, mock)
		orderNumber := "49927398716"
		orderUUID := uuid.New()
		params := AccrualCalculationParams{
			OrderNumber: orderNumber,
			Status:      OrderStatusTypePROCESSED,
			Amount:      decimal.NewFromFloat(1000),
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
		newDeposit := CreateTransactionParams{
			OrderNumber: &params.OrderNumber,
			OldBalance:  lastTx.NewBalance,
			Change:      params.Amount,
			NewBalance:  lastTx.NewBalance.Add(params.Amount),
			Type:        TransactionTypeDEPOSIT,
			UserID:      lastTx.UserID,
		}

		mock.ExpectBegin()

		processingOrder := pgxmock.
			NewRows([]string{"order_number", "process_status", "created_at", "updated_at", "deleted_at"}).
			AddRow(
				orderNumber,
				ProcessStatusTypePROCESSED,
				time.Now(),
				nil,
				nil,
			)

		mock.ExpectQuery(UpdateOrderToProcess).
			WithArgs(ProcessStatusTypePROCESSED, orderNumber).
			WillReturnRows(processingOrder)

		orderRows := pgxmock.
			NewRows([]string{"id", "user_id", "order_number", "status", "accrual", "created_at", "updated_at", "deleted_at"}).
			AddRow(
				orderUUID,
				uuid.Nil,
				orderNumber,
				OrderStatusTypePROCESSING,
				decimal.Zero,
				time.Now(),
				nil,
				nil,
			)
		mock.ExpectQuery(UpdateOrder).
			WithArgs(params.Status, TestDecimal{expected: params.Amount}, params.OrderNumber).
			WillReturnRows(orderRows)

		lastTxRows := pgxmock.
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
			WillReturnRows(lastTxRows)

		newTXRows := pgxmock.
			NewRows([]string{"id", "order_number", "user_id", "old_balance", "change", "new_balance", "type", "created_at"}).
			AddRow(
				newDeposit.UserID,
				newDeposit.OrderNumber,
				newDeposit.UserID,
				newDeposit.OldBalance.String(),
				newDeposit.Change.String(),
				newDeposit.NewBalance.String(),
				newDeposit.Type,
				time.Now(),
			)
		mock.ExpectQuery(CreateTransaction).
			WithArgs(
				newDeposit.OrderNumber,
				newDeposit.UserID,
				TestDecimal{expected: newDeposit.OldBalance},
				TestDecimal{expected: newDeposit.Change},
				TestDecimal{expected: newDeposit.NewBalance},
				newDeposit.Type,
			).WillReturnRows(newTXRows)

		mock.ExpectCommit()

		err = store.StoreAccrualCalculation(ctx, params)
		assert.NoError(t, err)
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
