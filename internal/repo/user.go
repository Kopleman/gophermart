package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	store  *pgxstore.PGXStore
	logger log.Logger
}

func NewUserRepo(l log.Logger, pgxStore *pgxstore.PGXStore) *UserRepo {
	return &UserRepo{
		logger: l,
		store:  pgxStore,
	}
}
func (r *UserRepo) CreateNewUser(ctx context.Context, createDto *dto.CreateUserDTO) (*pgxstore.User, error) {
	newUser, err := r.store.CreateNewUser(ctx, pgxstore.CreateUserParams{
		Login:        createDto.Login,
		PasswordHash: createDto.PasswordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return newUser, nil
}

func (r *UserRepo) GetUser(ctx context.Context, login string) (*pgxstore.User, error) {
	user, err := r.store.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgxstore.ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]*pgxstore.Transaction, error) {
	withdrawals, err := r.store.GetTransactions(ctx, pgxstore.GetTransactionsParams{
		UserID:   userID,
		IsUserID: true,
		Type:     pgxstore.TransactionTypeWITHDRAW,
		IsType:   true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user withdrawals: %w", err)
	}

	return withdrawals, nil
}
