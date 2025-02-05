package repo

import (
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/pgxstore"
)

type PageOptions struct {
	Page     *uint64
	PageSize *uint64
}

type Repos struct {
	userRepo *UserRepo
}

func NewRepository(logger log.Logger, store *pgxstore.PGXStore) *Repos {
	return &Repos{
		userRepo: NewUserRepo(logger, store),
	}
}

func (r Repos) User() *UserRepo {
	return r.userRepo
}
