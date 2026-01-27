package postgres

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/context"
	"github.com/jackc/pgx/v5"
)

type PostgresTxManager struct {
	pool PgxPool
}

func NewPostgresTxManager(pool PgxPool) database.TxManager {
	return &PostgresTxManager{
		pool: pool,
	}
}

func TxFromContext(ctx context.Context) pgx.Tx {
	tx, ok := ctx.Value(pkg.Tx).(pgx.Tx)
	if !ok {
		return nil
	}

	return tx
}

func (m *PostgresTxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if existing := TxFromContext(ctx); existing != nil {
		return fn(ctx)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, pkg.Tx, tx)

	err = fn(ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return rollbackErr
		}

		return err
	}

	return tx.Commit(ctx)
}
