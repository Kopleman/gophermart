package dto

import (
	"github.com/google/uuid"
)

type OrderDTO struct {
	ID          uuid.UUID `json:"id" example:"e9fab143-025d-4b10-9865-ef17401fbb17"`
	UserID      uuid.UUID `json:"user_id" example:"e9fab143-025d-4b10-9865-ef17401fbb17"`
	OrderNumber string    `json:"order_number" example:"49927398716"`
	Status      string    `json:"status" example:"NEW"`
}

type CreateOrderDTO struct {
	UserID      uuid.UUID `json:"user_id" example:"e9fab143-025d-4b10-9865-ef17401fbb17"`
	OrderNumber string    `json:"order_number" example:"49927398716"`
}
