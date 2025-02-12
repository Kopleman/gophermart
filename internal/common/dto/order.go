package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderDTO struct {
	OrderNumber string          `json:"order_number" example:"49927398716"`
	Status      string          `json:"status" example:"NEW"`
	CreatedAt   string          `json:"created_at" example:"2022-01-01T00:00:00+00:00"`
	Accrual     decimal.Decimal `json:"accrual" example:"0.1"`
	ID          uuid.UUID       `json:"id" example:"e9fab143-025d-4b10-9865-ef17401fbb17"`
	UserID      uuid.UUID       `json:"user_id" example:"e9fab143-025d-4b10-9865-ef17401fbb17"`
}

type CreateOrderDTO struct {
	OrderNumber string    `json:"order_number" example:"49927398716"`
	UserID      uuid.UUID `json:"user_id" example:"e9fab143-025d-4b10-9865-ef17401fbb17"`
}

type OrderInfoDTO struct {
	Number     string   `json:"number" example:"49927398716"`
	Status     string   `json:"status" example:"NEW"`
	Accrual    *float64 `json:"accrual,omitempty" example:"500"`
	UploadedAt string   `json:"uploaded_at" example:"2020-12-10T15:12:01+03:00"`
}

func (o *OrderDTO) ToInfoDTO() *OrderInfoDTO {
	dto := OrderInfoDTO{
		Number:     o.OrderNumber,
		Status:     o.Status,
		UploadedAt: o.CreatedAt,
		Accrual:    nil,
	}
	if o.Accrual.GreaterThan(decimal.Zero) {
		value, _ := o.Accrual.Float64()
		dto.Accrual = &value
	}
	return &dto
}
