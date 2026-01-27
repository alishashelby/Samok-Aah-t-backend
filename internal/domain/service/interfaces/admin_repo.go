package interfaces

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
)

//go:generate mockgen -source=admin_repo.go -destination=../mocks/admin_repo_mock.go -package=mocks AdminRepository
type AdminRepository interface {
	Save(ctx context.Context, admin *entity.Admin) error
	GetByID(ctx context.Context, id int64) (*entity.Admin, error)
	GetByAuthID(ctx context.Context, authID int64) (*entity.Admin, error)
	Update(ctx context.Context, admin *entity.Admin) (*entity.Admin, error)
}
