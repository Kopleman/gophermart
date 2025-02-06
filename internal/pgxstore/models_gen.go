// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package pgxstore

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatusType string

const (
	OrderStatusTypeNEW        OrderStatusType = "NEW"
	OrderStatusTypePROCESSING OrderStatusType = "PROCESSING"
	OrderStatusTypeINVALID    OrderStatusType = "INVALID"
	OrderStatusTypePROCESSED  OrderStatusType = "PROCESSED"
)

func (e *OrderStatusType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = OrderStatusType(s)
	case string:
		*e = OrderStatusType(s)
	default:
		return fmt.Errorf("unsupported scan type for OrderStatusType: %T", src)
	}
	return nil
}

type NullOrderStatusType struct {
	OrderStatusType OrderStatusType `json:"order_status_type"`
	Valid           bool            `json:"valid"` // Valid is true if OrderStatusType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullOrderStatusType) Scan(value interface{}) error {
	if value == nil {
		ns.OrderStatusType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.OrderStatusType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullOrderStatusType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.OrderStatusType), nil
}

func (e OrderStatusType) Valid() bool {
	switch e {
	case OrderStatusTypeNEW,
		OrderStatusTypePROCESSING,
		OrderStatusTypeINVALID,
		OrderStatusTypePROCESSED:
		return true
	}
	return false
}

func AllOrderStatusTypeValues() []OrderStatusType {
	return []OrderStatusType{
		OrderStatusTypeNEW,
		OrderStatusTypePROCESSING,
		OrderStatusTypeINVALID,
		OrderStatusTypePROCESSED,
	}
}

type ProcessStatusType string

const (
	ProcessStatusTypeNEW        ProcessStatusType = "NEW"
	ProcessStatusTypePROCESSING ProcessStatusType = "PROCESSING"
	ProcessStatusTypePROCESSED  ProcessStatusType = "PROCESSED"
)

func (e *ProcessStatusType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ProcessStatusType(s)
	case string:
		*e = ProcessStatusType(s)
	default:
		return fmt.Errorf("unsupported scan type for ProcessStatusType: %T", src)
	}
	return nil
}

type NullProcessStatusType struct {
	ProcessStatusType ProcessStatusType `json:"process_status_type"`
	Valid             bool              `json:"valid"` // Valid is true if ProcessStatusType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullProcessStatusType) Scan(value interface{}) error {
	if value == nil {
		ns.ProcessStatusType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ProcessStatusType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullProcessStatusType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ProcessStatusType), nil
}

func (e ProcessStatusType) Valid() bool {
	switch e {
	case ProcessStatusTypeNEW,
		ProcessStatusTypePROCESSING,
		ProcessStatusTypePROCESSED:
		return true
	}
	return false
}

func AllProcessStatusTypeValues() []ProcessStatusType {
	return []ProcessStatusType{
		ProcessStatusTypeNEW,
		ProcessStatusTypePROCESSING,
		ProcessStatusTypePROCESSED,
	}
}

type Order struct {
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time      `db:"deleted_at" json:"deleted_at"`
	OrderNumber string          `db:"order_number" json:"order_number"`
	Status      OrderStatusType `db:"status" json:"status"`
	Accrual     decimal.Decimal `db:"accrual" json:"accrual"`
	ID          uuid.UUID       `db:"id" json:"id"`
	UserID      uuid.UUID       `db:"user_id" json:"user_id"`
}

type OrdersToProcess struct {
	CreatedAt     time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt     *time.Time        `db:"updated_at" json:"updated_at"`
	DeletedAt     *time.Time        `db:"deleted_at" json:"deleted_at"`
	ProcessStatus ProcessStatusType `db:"process_status" json:"process_status"`
	OrderID       uuid.UUID         `db:"order_id" json:"order_id"`
}

type User struct {
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at"`
	Login        string     `db:"login" json:"login"`
	PasswordHash string     `db:"password_hash" json:"password_hash"`
	ID           uuid.UUID  `db:"id" json:"id"`
}
