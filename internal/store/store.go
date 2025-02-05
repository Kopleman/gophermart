package store

import (
	"context"

	"github.com/Kopleman/gophermart/internal/common/dto"
)

type Store interface {
	CreateUser(ctx context.Context, value *dto.MetricDTO) error
	GetUserByLogin(ctx context.Context, login string) (*User, error)
}
