package repo

import (
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/pgxstore"
)

type Repos struct {
	userRepo  *UserRepo
	orderRepo *OrderRepo
}

func NewRepository(logger log.Logger, store *pgxstore.PGXStore) *Repos {
	return &Repos{
		userRepo:  NewUserRepo(logger, store),
		orderRepo: NewOrderRepo(logger, store),
	}
}

func (r Repos) User() *UserRepo {
	return r.userRepo
}

func (r Repos) Order() *OrderRepo { return r.orderRepo }
