package repo

import (
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/pgxstore"
)

type Repos struct {
	userRepo    *UserRepo
	orderRepo   *OrderRepo
	balanceRepo *BalanceRepo
}

func NewRepository(logger log.Logger, store *pgxstore.PGXStore) *Repos {
	return &Repos{
		userRepo:    NewUserRepo(logger, store),
		orderRepo:   NewOrderRepo(logger, store),
		balanceRepo: NewBalanceRepo(logger, store),
	}
}

func (r Repos) User() *UserRepo {
	return r.userRepo
}

func (r Repos) Order() *OrderRepo { return r.orderRepo }

func (r Repos) Balance() *BalanceRepo { return r.balanceRepo }
