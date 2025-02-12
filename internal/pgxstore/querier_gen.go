// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package pgxstore

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Querier interface {
	CreateInitUserTransaction(ctx context.Context, userID uuid.UUID) (*Transaction, error)
	CreateOrder(ctx context.Context, arg CreateOrderParams) (*Order, error)
	CreateTransaction(ctx context.Context, arg CreateTransactionParams) (*Transaction, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
	GetLastUserTransaction(ctx context.Context, userID uuid.UUID) (*Transaction, error)
	GetOrderByNumber(ctx context.Context, orderNumber string) (*Order, error)
	GetRegisteredProcessingOrders(ctx context.Context, limit int32) ([]*OrdersToProcess, error)
	GetStartProcessingOrders(ctx context.Context) ([]*OrdersToProcess, error)
	GetTransactions(ctx context.Context, arg GetTransactionsParams) ([]*Transaction, error)
	GetUserById(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByLogin(ctx context.Context, login string) (*User, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*Order, error)
	GetUserWithdrawalsSum(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	PickOrdersToProcess(ctx context.Context, limit int32) ([]*OrdersToProcess, error)
	PutOrderForProcessing(ctx context.Context, orderNumber string) (*OrdersToProcess, error)
	UpdateOrder(ctx context.Context, arg UpdateOrderParams) (*Order, error)
	UpdateOrderToProcess(ctx context.Context, arg UpdateOrderToProcessParams) (*OrdersToProcess, error)
}

var _ Querier = (*Queries)(nil)
