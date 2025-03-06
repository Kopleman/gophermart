package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BalanceDTO struct {
	Current   float64 `json:"current" example:"500.5"`
	Withdrawn float64 `json:"withdrawn" example:"42"`
}

type WithdrawRequestDTO struct {
	Order string  `json:"order" validate:"required" example:"49927398716"`
	Sum   float64 `json:"sum" validate:"required,numeric" example:"100.2"`
}

type WithdrawDTO struct {
	Order  string          `json:"order"`
	Amount decimal.Decimal `json:"amount"`
	UserID uuid.UUID       `json:"userID"`
}

type WithdrawalItemDTO struct {
	Order       string  `json:"order" validate:"required" example:"49927398716"`
	ProcessedAt string  `json:"processed_at" example:"2020-12-10T15:12:01+03:00"`
	Sum         float64 `json:"sum" validate:"required,numeric" example:"100.2"`
}
