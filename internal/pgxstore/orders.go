package pgxstore

import (
	"context"
	"fmt"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/jackc/pgx/v5"
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
	}
}

func (p *PGXStore) CreateNewOrder(ctx context.Context, createDTO *dto.CreateOrderDTO) (*Order, *OrdersToProcess, error) {
	tx, err := p.startTx(ctx, &pgx.TxOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("pgxstore.CreateNewOrder could not start transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:all // its safe

	createParams := CreateOrderParams{
		UserID:      createDTO.UserID,
		OrderNumber: createDTO.OrderNumber,
	}
	newOrder, createErr := p.CreateOrder(ctx, createParams)
	if createErr != nil {
		return nil, nil, fmt.Errorf("pgxstore.CreateNewOrder could not create order: %w", createErr)
	}

	orderToProcess, putErr := p.PutOrderForProcessing(ctx, newOrder.ID)
	if putErr != nil {
		return nil, nil, fmt.Errorf("pgxstore.CreateNewOrder could not create task for proccesing: %w", putErr)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, nil, fmt.Errorf("pgxstore.CreateNewOrder could not commit transaction: %w", commitErr)
	}

	return newOrder, orderToProcess, nil
}
