package interfaces

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
)

//go:generate mockgen -source=booking_repo.go -destination=../mocks/booking_repo_mock.go -package=mocks BookingRepository
type BookingRepository interface {
	Save(ctx context.Context, booking *entity.Booking) error
	GetByID(ctx context.Context, id int64) (*entity.Booking, error)
	Update(ctx context.Context, b *entity.Booking) (*entity.Booking, error)
	GetAll(ctx context.Context, opts *entity.Options) ([]*entity.Booking, error)
}
