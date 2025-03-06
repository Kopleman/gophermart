package pgxstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *PGXStore) StartTx(ctx context.Context) (*PGXStore, error) {
	tx, err := p.startTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	newQ := p.WithTx(tx)

	return &PGXStore{
		logger:   p.logger,
		Queries:  newQ,
		db:       p.db,
		activeTX: tx,
	}, err
}

func (p *PGXStore) RollbackTx(ctx context.Context) error {
	if p.activeTX == nil {
		return nil
	}

	if err := p.activeTX.Rollback(ctx); err != nil {
		if errors.Is(err, pgx.ErrTxClosed) {
			return nil
		}
		return fmt.Errorf("pgxstore: failed to rollback transaction: %w", err)
	}

	return nil
}

func (p *PGXStore) CommitTx(ctx context.Context) error {
	if p.activeTX == nil {
		return nil
	}
	if err := p.activeTX.Commit(ctx); err != nil {
		return fmt.Errorf("pgxstore: failed to commit transaction: %w", err)
	}
	return nil
}

func (p *PGXStore) startTx(ctx context.Context, opts *pgx.TxOptions) (pgx.Tx, error) {
	txOpts := pgx.TxOptions{}
	if opts != nil {
		txOpts = *opts
	}
	tx, err := p.db.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	return tx, nil
}

func (p *PGXStore) WithTx(tx pgx.Tx) *Queries {
	return p.Queries.WithTx(tx)
}

type PgxPool interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Ping(context.Context) error
	Close()
}

type PGXStore struct {
	*Queries
	db       PgxPool
	logger   log.Logger
	activeTX pgx.Tx
}

func NewPGXStore(l log.Logger, db PgxPool) *PGXStore {
	return &PGXStore{
		Queries: New(db),
		db:      db,
		logger:  l,
	}
}
