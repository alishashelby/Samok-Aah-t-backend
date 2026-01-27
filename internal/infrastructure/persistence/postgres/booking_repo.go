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

type DefaultBookingRepository struct {
	db *postgres.PostgresDb
}

func NewDefaultBookingRepository(db *postgres.PostgresDb) *DefaultBookingRepository {
	return &DefaultBookingRepository{
		db: db,
	}
}

func (d *DefaultBookingRepository) Save(ctx context.Context, b *entity.Booking) error {
	query, args, err := sq.Insert("bookings").
		Columns("client_id", "model_service_id", "slot_id", "address", "status", "expires_at").
		Values(b.ClientID, b.ModelServiceID, b.SlotID, b.Address, b.Status, b.ExpiresAt).
		Suffix("RETURNING booking_id, created_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	return d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(&b.ID, &b.CreatedAt)
}

func (d *DefaultBookingRepository) GetByID(ctx context.Context, id int64) (*entity.Booking, error) {
	query, args, err := sq.Select(
		"booking_id", "client_id", "model_service_id",
		"slot_id", "address", "status", "expires_at", "created_at").
		From("bookings").
		Where(sq.Eq{
			"booking_id": id,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Booking
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(
			&res.ID, &res.ClientID, &res.ModelServiceID, &res.SlotID,
			&res.Address, &res.Status, &res.ExpiresAt, &res.CreatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultBookingRepository) Update(ctx context.Context, b *entity.Booking) (*entity.Booking, error) {
	query, args, err := sq.Update("bookings").
		SetMap(map[string]interface{}{
			"client_id":        b.ClientID,
			"model_service_id": b.ModelServiceID,
			"slot_id":          b.SlotID,
			"address":          b.Address,
			"status":           b.Status,
			"expires_at":       b.ExpiresAt,
		}).
		Where(sq.Eq{
			"booking_id": b.ID,
		}).
		Suffix("RETURNING booking_id, client_id, model_service_id, " +
			"slot_id, address, status, expires_at, created_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var res entity.Booking
	err = d.getExecutor(ctx).
		QueryRow(ctx, query, args...).
		Scan(
			&res.ID, &res.ClientID, &res.ModelServiceID, &res.SlotID,
			&res.Address, &res.Status, &res.ExpiresAt, &res.CreatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNoRowsFound
		}

		return nil, err
	}

	return &res, nil
}

func (d *DefaultBookingRepository) GetAll(
	ctx context.Context,
	opts *entity.Options,
) ([]*entity.Booking, error) {

	query, args, err :=
		sq.Select(
			"booking_id", "client_id", "model_service_id", "slot_id",
			"address", "status", "expires_at", "created_at",
		).
			From("bookings").
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

	var res []*entity.Booking
	for rows.Next() {
		var booking entity.Booking
		if err = rows.Scan(
			&booking.ID, &booking.ClientID, &booking.ModelServiceID, &booking.SlotID,
			&booking.Address, &booking.Status, &booking.ExpiresAt, &booking.CreatedAt,
		); err != nil {
			return nil, err
		}

		res = append(res, &booking)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (d *DefaultBookingRepository) getExecutor(ctx context.Context) postgres.Executor {
	if tx := postgres.TxFromContext(ctx); tx != nil {
		return tx
	}

	return d.db.Pool
}
