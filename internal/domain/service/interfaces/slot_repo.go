package interfaces

import (
	"context"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
)

//go:generate mockgen -source=slot_repo.go -destination=../mocks/slot_repo_mock.go -package=mocks SlotRepository
type SlotRepository interface {
	Save(ctx context.Context, slot *entity.Slot) error
	GetByID(ctx context.Context, id int64) (*entity.Slot, error)
	GetByModelID(ctx context.Context, modelID int64) ([]*entity.Slot, error)
	GetOverlappingSlots(ctx context.Context, modelID int64, start, end time.Time) ([]*entity.Slot, error)
	Update(ctx context.Context, slot *entity.Slot) (*entity.Slot, error)
}
