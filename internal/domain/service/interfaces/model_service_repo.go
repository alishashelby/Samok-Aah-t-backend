package interfaces

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
)

//go:generate mockgen -source=model_service_repo.go -destination=../mocks/model_service_repo_mock.go -package=mocks ModelServiceRepository
type ModelServiceRepository interface {
	Save(ctx context.Context, service *entity.ModelService) error
	GetByID(ctx context.Context, id int64, includeInactive bool) (*entity.ModelService, error)
	GetAll(ctx context.Context, opts *entity.Options, includeInactive bool) ([]*entity.ModelService, error)
	GetByModelID(ctx context.Context, modelID int64, opts *entity.Options, includeInactive bool) ([]*entity.ModelService, error)
	HasBookings(ctx context.Context, serviceID int64) (bool, error)
	Update(ctx context.Context, service *entity.ModelService) (*entity.ModelService, error)
	Deactivate(ctx context.Context, id int64) error
}
