package postgres

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database/postgres"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	"github.com/jackc/pgx/v5"
)

type DefaultAdminRepository struct {
	db *postgres.PostgresDb
}

func NewDefaultAdminRepository(db *postgres.PostgresDb) *DefaultAdminRepository {
	return &DefaultAdminRepository{
		db: db,
	}
}

func (d *DefaultAdminRepository) Save(ctx context.Context, admin *entity.Admin) error {
	query, args, err := sq.Insert("admins").
		Columns("auth_id", "permissions").
		Values(admin.AuthID, admin.Permissions).
		Suffix("RETURNING admin_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	return d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&admin.ID)
}

func (d *DefaultAdminRepository) GetByID(ctx context.Context, id int64) (*entity.Admin, error) {
	query, args, err := sq.Select("admin_id", "auth_id", "permissions").
		From("admins").
		Where(sq.Eq{
			"admin_id": id,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Admin
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.AuthID, &res.Permissions)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultAdminRepository) GetByAuthID(ctx context.Context, authID int64) (*entity.Admin, error) {
	query, args, err := sq.Select("admin_id", "auth_id", "permissions").
		From("admins").
		Where(sq.Eq{
			"auth_id": authID,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Admin
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.AuthID, &res.Permissions)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultAdminRepository) Update(ctx context.Context, admin *entity.Admin) (*entity.Admin, error) {
	query, args, err := sq.Update("admins").
		Set("permissions", admin.Permissions).
		Where(sq.Eq{
			"admin_id": admin.ID,
		}).
		Suffix("RETURNING admin_id, auth_id, permissions").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Admin
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.AuthID, &res.Permissions)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultAdminRepository) getExecutor(ctx context.Context) postgres.Executor {
	if tx := postgres.TxFromContext(ctx); tx != nil {
		return tx
	}

	return d.db.Pool
}
