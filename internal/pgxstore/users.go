package pgxstore

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (p *PGXStore) CreateNewUser(
	ctx context.Context,
	createParams CreateUserParams,
) (*User, error) {
	tx, err := p.startTx(ctx, &pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("pgxstore.CreateNewUser could not start transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:all // its safe

	newUser, createErr := p.CreateUser(ctx, createParams)
	if createErr != nil {
		return nil, fmt.Errorf("pgxstore.CreateUser could not create order: %w", createErr)
	}

	_, initTxErr := p.CreateInitUserTransaction(ctx, newUser.ID)
	if initTxErr != nil {
		return nil, fmt.Errorf("pgxstore.CreateNewUser could not create init user tx: %w", initTxErr)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, fmt.Errorf("pgxstore.CreateNewUser could not commit transaction: %w", commitErr)
	}

	return newUser, nil
}
