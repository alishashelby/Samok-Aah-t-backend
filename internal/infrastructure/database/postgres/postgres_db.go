package postgres

import (
	"context"
	"fmt"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

const dsnTemplate = "postgres://%s:%s@%s:%s/%s"

type PostgresDb struct {
	Pool PgxPool
}

func New(ctx context.Context, envConfig *config.EnvConfig) (*PostgresDb, error) {
	dsn := fmt.Sprintf(
		dsnTemplate,
		envConfig.PostgresUser,
		envConfig.PostgresPassword,
		envConfig.PostgresHost,
		envConfig.PostgresPort,
		envConfig.PostgresDB,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &PostgresDb{
		Pool: pool,
	}, nil
}

func (p *PostgresDb) Close() {
	p.Pool.Close()
}
