package postgres

import (
	"context"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database/postgres"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	"github.com/jackc/pgx/v5"
)

type DefaultSlotRepository struct {
	db *postgres.PostgresDb
}

func NewDefaultSlotRepository(db *postgres.PostgresDb) *DefaultSlotRepository {
	return &DefaultSlotRepository{
		db: db,
	}
}

func (d *DefaultSlotRepository) Save(ctx context.Context, slot *entity.Slot) error {
	query, args, err := sq.Insert("slots").
		Columns("model_id", "start_time", "end_time", "status").
		Values(slot.ModelID, slot.StartTime, slot.EndTime, slot.Status).
		Suffix("RETURNING slot_id, created_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	return d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&slot.ID, &slot.CreatedAt)
}

func (d *DefaultSlotRepository) GetByID(ctx context.Context, id int64) (*entity.Slot, error) {
	query, args, err := sq.Select("slot_id", "model_id", "start_time", "end_time", "status", "created_at").
		From("slots").
		Where(sq.Eq{
			"slot_id": id,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Slot
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.ModelID, &res.StartTime, &res.EndTime, &res.Status, &res.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultSlotRepository) GetByModelID(ctx context.Context, modelID int64) ([]*entity.Slot, error) {
	query, args, err := sq.Select("slot_id", "model_id", "start_time", "end_time", "status", "created_at").
		From("slots").
		Where(sq.Eq{
			"model_id": modelID,
		}).
		OrderBy("start_time ASC").
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

	var res []*entity.Slot
	for rows.Next() {
		var slot entity.Slot
		if err = rows.Scan(
			&slot.ID, &slot.ModelID, &slot.StartTime, &slot.EndTime,
			&slot.Status, &slot.CreatedAt); err != nil {
			return nil, err
		}

		res = append(res, &slot)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (d *DefaultSlotRepository) GetOverlappingSlots(ctx context.Context, modelID int64, start, end time.Time) ([]*entity.Slot, error) {
	query, args, err := sq.Select("slot_id", "model_id", "start_time", "end_time", "status", "created_at").
		From("slots").
		Where(sq.Eq{
			"model_id": modelID,
		}).
		Where(sq.And{
			sq.Expr("start_time < ?", end),
			sq.Expr("? < end_time", start),
		}).
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

	var res []*entity.Slot
	for rows.Next() {
		var slot entity.Slot
		if err = rows.Scan(
			&slot.ID,
			&slot.ModelID,
			&slot.StartTime,
			&slot.EndTime,
			&slot.Status,
			&slot.CreatedAt,
		); err != nil {
			return nil, err
		}

		res = append(res, &slot)
	}

	return res, nil
}

func (d *DefaultSlotRepository) Update(ctx context.Context, slot *entity.Slot) (*entity.Slot, error) {
	query, args, err := sq.Update("slots").
		Set("start_time", slot.StartTime).
		Set("end_time", slot.EndTime).
		Set("status", slot.Status).
		Where(sq.Eq{
			"slot_id": slot.ID,
		}).
		Suffix("RETURNING slot_id, model_id, start_time, end_time, status, created_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Slot
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.ModelID, &res.StartTime, &res.EndTime, &res.Status, &res.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, err
}

func (d *DefaultSlotRepository) getExecutor(ctx context.Context) postgres.Executor {
	if tx := postgres.TxFromContext(ctx); tx != nil {
		return tx
	}

	return d.db.Pool
}
