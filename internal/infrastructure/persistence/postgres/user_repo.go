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

type DefaultUserRepository struct {
	db *postgres.PostgresDb
}

func NewDefaultUserRepository(db *postgres.PostgresDb) *DefaultUserRepository {
	return &DefaultUserRepository{
		db: db,
	}
}

func (d *DefaultUserRepository) Save(ctx context.Context, user *entity.User) error {
	query, args, err := sq.Insert("users").
		Columns("auth_id", "name", "birth_date", "is_verified").
		Values(user.AuthID, user.Name, user.BirthDate, user.IsVerified).
		Suffix("RETURNING user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	return d.getExecutor(ctx).
		QueryRow(ctx, query, args...).Scan(&user.ID)
}

func (d *DefaultUserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	query, args, err := sq.Select("user_id", "auth_id", "name", "birth_date", "is_verified").
		From("users").
		Where(sq.Eq{
			"user_id": id,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.User
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.AuthID, &res.Name, &res.BirthDate, &res.IsVerified)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultUserRepository) GetByAuthID(ctx context.Context, authID int64) (*entity.User, error) {
	query, args, err := sq.Select("user_id", "auth_id", "name", "birth_date", "is_verified").
		From("users").
		Where(sq.Eq{
			"auth_id": authID,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.User
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.AuthID, &res.Name, &res.BirthDate, &res.IsVerified)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultUserRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	query, args, err := sq.Update("users").
		Set("name", user.Name).
		Set("is_verified", user.IsVerified).
		Where(sq.Eq{
			"user_id": user.ID,
		}).
		Suffix("RETURNING user_id, auth_id, name, birth_date, is_verified").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.User
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.AuthID, &res.Name, &res.BirthDate, &res.IsVerified)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultUserRepository) GetAll(
	ctx context.Context,
	opts *entity.Options,
) ([]*entity.User, error) {

	query, args, err := sq.Select("user_id", "name",
		"birth_date", "auth_id", "is_verified").
		From("users").
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := d.getExecutor(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*entity.User
	for rows.Next() {
		var user entity.User
		if err = rows.Scan(&user.ID, &user.Name, &user.BirthDate, &user.AuthID, &user.IsVerified); err != nil {
			return nil, err
		}

		res = append(res, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (d *DefaultUserRepository) CountByRole(ctx context.Context, role entity.Role) (int64, error) {
	query, args, err := sq.Select("COUNT(*)").
		From("users u").
		Join("auth a ON a.auth_id = u.auth_id").
		Where(sq.Eq{
			"a.role": role.String(),
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, err
	}

	var count int64
	err = d.getExecutor(ctx).QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (d *DefaultUserRepository) getExecutor(ctx context.Context) postgres.Executor {
	if tx := postgres.TxFromContext(ctx); tx != nil {
		return tx
	}

	return d.db.Pool
}
