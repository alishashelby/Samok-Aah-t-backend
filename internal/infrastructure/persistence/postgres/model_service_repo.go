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

type DefaultModelServiceRepository struct {
	db *postgres.PostgresDb
}

func NewDefaultModelServiceRepository(db *postgres.PostgresDb) *DefaultModelServiceRepository {
	return &DefaultModelServiceRepository{
		db: db,
	}
}

func (d *DefaultModelServiceRepository) Save(ctx context.Context, service *entity.ModelService) error {
	query, args, err := sq.Insert("model_services").
		Columns("model_id", "title", "description", "is_active", "price").
		Values(service.ModelID, service.Title, service.Description, service.IsActive, service.Price).
		Suffix("RETURNING model_service_id, created_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	return d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&service.ID, &service.CreatedAt)
}

func (d *DefaultModelServiceRepository) GetByID(ctx context.Context, id int64,
	includeInactive bool) (*entity.ModelService, error) {

	builder := sq.Select("model_service_id", "model_id", "title", "description",
		"price", "is_active", "created_at").
		From("model_services").
		Where(sq.Eq{
			"model_service_id": id,
		})

	if !includeInactive {
		builder = builder.Where(sq.Eq{
			"is_active": true,
		})
	}

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.ModelService
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.ModelID, &res.Title, &res.Description, &res.Price, &res.IsActive, &res.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultModelServiceRepository) GetAll(ctx context.Context,
	opts *entity.Options, includeInactive bool) ([]*entity.ModelService, error) {

	builder := sq.Select("model_service_id", "model_id", "title",
		"description", "price", "is_active", "created_at").
		From("model_services").
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset))

	if !includeInactive {
		builder = builder.Where(sq.Eq{
			"is_active": true,
		})
	}

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := d.getExecutor(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*entity.ModelService, 0)
	for rows.Next() {
		var service entity.ModelService
		if err = rows.Scan(
			&service.ID, &service.ModelID, &service.Title, &service.Description,
			&service.Price, &service.IsActive, &service.CreatedAt,
		); err != nil {
			return nil, err
		}

		res = append(res, &service)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (d *DefaultModelServiceRepository) GetByModelID(ctx context.Context, modelID int64,
	opts *entity.Options, includeInactive bool) ([]*entity.ModelService, error) {

	builder := sq.Select("model_service_id", "model_id", "title",
		"description", "price", "is_active", "created_at").
		From("model_services").
		Where(sq.Eq{
			"model_id": modelID,
		}).
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset))

	if !includeInactive {
		builder = builder.Where(sq.Eq{
			"is_active": true,
		})
	}

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := d.getExecutor(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*entity.ModelService, 0)
	for rows.Next() {
		var service entity.ModelService
		if err = rows.Scan(
			&service.ID, &service.ModelID, &service.Title, &service.Description,
			&service.Price, &service.IsActive, &service.CreatedAt,
		); err != nil {
			return nil, err
		}

		res = append(res, &service)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (d *DefaultModelServiceRepository) Update(ctx context.Context, service *entity.ModelService) (*entity.ModelService, error) {
	query, args, err := sq.Update("model_services").
		Set("title", service.Title).
		Set("description", service.Description).
		Set("price", service.Price).
		Where(sq.Eq{
			"model_service_id": service.ID,
		}).
		Suffix("RETURNING model_service_id, model_id, title, description, price, is_active, created_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.ModelService
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.ModelID, &res.Title, &res.Description, &res.Price, &res.IsActive, &res.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultModelServiceRepository) HasBookings(ctx context.Context, serviceID int64) (bool, error) {
	query, args, err := sq.Select("1").
		From("bookings").
		Where(sq.Eq{
			"model_service_id": serviceID,
		}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, err
	}

	err = d.getExecutor(ctx).QueryRow(ctx, query, args...).Scan(new(int))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (d *DefaultModelServiceRepository) Deactivate(ctx context.Context, id int64) error {
	query, args, err := sq.Update("model_services").
		Set("is_active", false).
		Where(sq.Eq{
			"model_service_id": id,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	res, err := d.getExecutor(ctx).Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return persistence.ErrNoRowsAffected
	}

	return nil
}

func (d *DefaultModelServiceRepository) getExecutor(ctx context.Context) postgres.Executor {
	if tx := postgres.TxFromContext(ctx); tx != nil {
		return tx
	}

	return d.db.Pool
}
