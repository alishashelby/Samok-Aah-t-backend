package interfaces

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
)

//go:generate mockgen -source=order_repo.go -destination=../mocks/order_repo_mock.go -package=mocks OrderRepository
type OrderRepository interface {
	Save(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id int64) (*entity.Order, error)
	GetByBookingID(ctx context.Context, bookingID int64) (*entity.Order, error)
	UpdateStatus(ctx context.Context, order *entity.Order) (*entity.Order, error)
	GetAllByModelID(ctx context.Context, modelID int64, opts *entity.Options) ([]*entity.Order, error)
	GetAllByClientID(ctx context.Context, clientID int64) ([]*entity.Order, error)
	GetAll(ctx context.Context, opts *entity.Options) ([]*entity.Order, error)
}
