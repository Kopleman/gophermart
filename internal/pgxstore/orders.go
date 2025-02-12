package pgxstore

import (
	"context"
	"fmt"
	"time"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

var (
	orderStatusTypeName = map[OrderStatusType]string{
		OrderStatusTypeNEW:        "NEW",
		OrderStatusTypePROCESSING: "PROCESSING",
		OrderStatusTypeINVALID:    "INVALID",
		OrderStatusTypePROCESSED:  "PROCESSED",
	}
	orderStatusTypeValue = map[string]OrderStatusType{
		"NEW":        OrderStatusTypeNEW,
		"PROCESSING": OrderStatusTypePROCESSING,
		"INVALID":    OrderStatusTypeINVALID,
		"PROCESSED":  OrderStatusTypePROCESSED,
	}
)

func (os OrderStatusType) String() string {
	res, ok := orderStatusTypeName[os]
	if !ok {
		return ""
	}
	return res
}

func StringToOrderStatusType(s string) (OrderStatusType, bool) {
	res, ok := orderStatusTypeValue[s]
	return res, ok
}

func (o *Order) ToDTO() *dto.OrderDTO {
	return &dto.OrderDTO{
		ID:          o.ID,
		UserID:      o.UserID,
		OrderNumber: o.OrderNumber,
		Status:      o.Status.String(),
		Accrual:     o.Accrual,
		CreatedAt:   o.CreatedAt.Format(time.RFC3339),
	}
}

func (p *PGXStore) CreateNewOrder(
	ctx context.Context,
	createParams CreateOrderParams,
) (*Order, *OrdersToProcess, error) {
	tx, err := p.startTx(ctx, &pgx.TxOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("pgxstore.CreateNewOrder could not start transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:all // its safe

	newOrder, createErr := p.CreateOrder(ctx, createParams)
	if createErr != nil {
		return nil, nil, fmt.Errorf("pgxstore.CreateNewOrder could not create order: %w", createErr)
	}

	orderToProcess, putErr := p.PutOrderForProcessing(ctx, createParams.OrderNumber)
	if putErr != nil {
		return nil, nil, fmt.Errorf("pgxstore.CreateNewOrder could not create task for proccesing: %w", putErr)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, nil, fmt.Errorf("pgxstore.CreateNewOrder could not commit transaction: %w", commitErr)
	}

	return newOrder, orderToProcess, nil
}

func (p *PGXStore) RegisterOrderProcessing(ctx context.Context, orderNumber string) error {
	tx, err := p.startTx(ctx, &pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("pgxstore.RegisterOrderProcessing could not start transaction: %w", err)
	}

	_, updateErr := p.UpdateOrderToProcess(ctx, UpdateOrderToProcessParams{
		OrderNumber:   orderNumber,
		ProcessStatus: ProcessStatusTypeREGISTERED,
	})
	if updateErr != nil {
		return fmt.Errorf("pgxstore.RegisterOrderProcessing could not update order process status: %w", updateErr)
	}
	_, err = p.UpdateOrder(ctx, UpdateOrderParams{
		OrderNumber: orderNumber,
		Status:      OrderStatusTypePROCESSING,
	})
	if err != nil {
		return fmt.Errorf("pgxstore.RegisterOrderProcessing could not update order: %w", err)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("pgxstore.RegisterOrderProcessing could not commit transaction: %w", commitErr)
	}

	return nil
}

type AccrualCalculationParams struct {
	OrderNumber string
	Status      OrderStatusType
	Amount      decimal.Decimal
}

func (p *PGXStore) StoreAccrualCalculation(
	ctx context.Context,
	params AccrualCalculationParams,
) error {
	tx, err := p.startTx(ctx, &pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("pgxstore.ProcessAccrualCalculation could not start transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:all // its safe

	_, err = p.UpdateOrderToProcess(ctx, UpdateOrderToProcessParams{
		OrderNumber:   params.OrderNumber,
		ProcessStatus: ProcessStatusTypePROCESSED,
	})
	if err != nil {
		return fmt.Errorf("pgxstore.ProcessAccrualCalculation could not update order to process status: %w", err)
	}

	order, updateErr := p.UpdateOrder(ctx, UpdateOrderParams{
		OrderNumber: params.OrderNumber,
		Status:      params.Status,
		Accrual:     params.Amount,
	})
	if updateErr != nil {
		return fmt.Errorf("pgxstore.ProcessAccrualCalculation could not update order: %w", updateErr)
	}

	lastUserTx, lastTxErr := p.GetLastUserTransaction(ctx, order.UserID)
	if lastTxErr != nil {
		return fmt.Errorf("pgxstore.ProcessAccrualCalculation could not get last user transaction: %w", lastTxErr)
	}

	depositParams := CreateTransactionParams{
		OrderNumber: &params.OrderNumber,
		UserID:      order.UserID,
		OldBalance:  lastUserTx.NewBalance,
		Change:      params.Amount,
		NewBalance:  lastUserTx.NewBalance.Add(params.Amount),
		Type:        TransactionTypeDEPOSIT,
	}
	_, err = p.CreateTransaction(ctx, depositParams)
	if err != nil {
		return fmt.Errorf("pgxstore.ProcessAccrualCalculation could not create deposit transaction: %w", err)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("pgxstore.ProcessAccrualCalculation could not commit transaction: %w", commitErr)
	}

	return nil
}
