package interfaces

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
)

//go:generate mockgen -source=../../../infrastructure/database/interfaces.go -destination=../mocks/tx_manager_mock.go -package=mocks
//go:generate mockgen -source=auth_repo.go -destination=../mocks/auth_repo_mock.go -package=mocks AuthRepository
type AuthRepository interface {
	Save(ctx context.Context, auth *entity.Auth) error
	GetByEmail(ctx context.Context, email string) (*entity.Auth, error)
}
