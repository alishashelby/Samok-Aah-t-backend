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

type DefaultOrderRepository struct {
	db *postgres.PostgresDb
}

func NewDefaultOrderRepository(db *postgres.PostgresDb) *DefaultOrderRepository {
	return &DefaultOrderRepository{
		db: db,
	}
}

func (d *DefaultOrderRepository) Save(ctx context.Context, order *entity.Order) error {
	query, args, err := sq.Insert("orders").
		Columns("booking_id", "status").
		Values(order.BookingID, order.Status).
		Suffix("RETURNING order_id, created_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	return d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&order.ID, &order.CreatedAt)
}

func (d *DefaultOrderRepository) GetByID(ctx context.Context, id int64) (*entity.Order, error) {
	query, args, err := sq.Select("order_id", "booking_id", "status", "created_at").
		From("orders").
		Where(sq.Eq{
			"order_id": id,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Order
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.BookingID, &res.Status, &res.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}
		return nil, err
	}

	return &res, nil
}

func (d *DefaultOrderRepository) GetByBookingID(ctx context.Context, bookingID int64) (*entity.Order, error) {
	query, args, err := sq.Select("order_id", "booking_id", "status", "created_at").
		From("orders").
		Where(sq.Eq{
			"booking_id": bookingID,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Order
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.BookingID, &res.Status, &res.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}
		return nil, err
	}

	return &res, nil
}

func (d *DefaultOrderRepository) UpdateStatus(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	query, args, err := sq.Update("orders").
		Set("status", order.Status).
		Where(sq.Eq{
			"order_id": order.ID,
		}).
		Suffix("RETURNING order_id, booking_id, status, created_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Order
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.BookingID, &res.Status, &res.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}
		return nil, err
	}

	return &res, nil
}

func (d *DefaultOrderRepository) GetAllByModelID(ctx context.Context, modelID int64,
	opts *entity.Options) ([]*entity.Order, error) {
	query, args, err := sq.Select(
		"o.order_id", "o.booking_id", "o.status", "o.created_at").
		From("orders o").
		Join("bookings b ON o.booking_id = b.booking_id").
		Join("model_services ms ON b.model_service_id = ms.model_service_id").
		Where(sq.Eq{
			"ms.model_id": modelID,
		}).
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

	var res []*entity.Order
	for rows.Next() {
		var order entity.Order
		if err = rows.Scan(&order.ID, &order.BookingID, &order.Status, &order.CreatedAt); err != nil {
			return nil, err
		}

		res = append(res, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (d *DefaultOrderRepository) GetAllByClientID(ctx context.Context, clientID int64) ([]*entity.Order, error) {
	query, args, err := sq.Select("o.order_id", "o.booking_id", "o.status", "o.created_at").
		From("orders o").
		Join("bookings b ON o.booking_id = b.booking_id").
		Where(sq.Eq{
			"b.client_id": clientID,
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

	var res []*entity.Order
	for rows.Next() {
		var order entity.Order
		if err = rows.Scan(&order.ID, &order.BookingID, &order.Status, &order.CreatedAt); err != nil {
			return nil, err
		}

		res = append(res, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (d *DefaultOrderRepository) GetAll(ctx context.Context, opts *entity.Options) ([]*entity.Order, error) {
	query, args, err := sq.Select("order_id", "booking_id", "status", "created_at").
		From("orders").
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

	var res []*entity.Order
	for rows.Next() {
		var order entity.Order
		if err = rows.Scan(&order.ID, &order.BookingID, &order.Status, &order.CreatedAt); err != nil {
			return nil, err
		}

		res = append(res, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (d *DefaultOrderRepository) getExecutor(ctx context.Context) postgres.Executor {
	if tx := postgres.TxFromContext(ctx); tx != nil {
		return tx
	}

	return d.db.Pool
}
