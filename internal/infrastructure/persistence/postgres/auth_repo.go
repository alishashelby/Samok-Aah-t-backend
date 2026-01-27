package postgres

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database/postgres"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DefaultAuthRepository struct {
	db *postgres.PostgresDb
}

func NewDefaultAuthRepository(db *postgres.PostgresDb) *DefaultAuthRepository {
	return &DefaultAuthRepository{
		db: db,
	}
}

func (d *DefaultAuthRepository) Save(ctx context.Context, auth *entity.Auth) error {
	query, args, err := sq.Insert("auth").
		Columns("email", "password_hash", "role").
		Values(auth.Email, auth.PasswordHash, auth.Role).
		Suffix("RETURNING auth_id, created_at, role").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&auth.ID, &auth.CreatedAt, &auth.Role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == persistence.UniqueViolationCode {
			return persistence.ErrDuplicateKey
		}

		return err
	}

	return nil
}

func (d *DefaultAuthRepository) GetByEmail(ctx context.Context, email string) (*entity.Auth, error) {
	query, args, err := sq.Select("auth_id", "email", "password_hash", "role", "created_at").
		From("auth").
		Where(sq.Eq{
			"email": email,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Auth
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.Email, &res.PasswordHash, &res.Role, &res.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultAuthRepository) getExecutor(ctx context.Context) postgres.Executor {
	if tx := postgres.TxFromContext(ctx); tx != nil {
		return tx
	}

	return d.db.Pool
}
