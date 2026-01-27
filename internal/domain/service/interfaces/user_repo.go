package interfaces

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
)

//go:generate mockgen -source=user_repo.go -destination=../mocks/user_repo_mock.go -package=mocks UserRepository
type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByAuthID(ctx context.Context, authID int64) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	GetAll(ctx context.Context, opts *entity.Options) ([]*entity.User, error)
	CountByRole(ctx context.Context, role entity.Role) (int64, error)
}
