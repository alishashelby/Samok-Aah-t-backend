package service

import (
	"context"
	"errors"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/interfaces"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
	"golang.org/x/crypto/bcrypt"
)

type DefaultAuthService struct {
	authRepo   interfaces.AuthRepository
	jwtService *JWTService
	txManager  database.TxManager
	logger     pkg.Logger
}

func NewDefaultAuthService(authRepo interfaces.AuthRepository, service *JWTService,
	txManager database.TxManager, logger pkg.Logger) *DefaultAuthService {
	return &DefaultAuthService{
		authRepo:   authRepo,
		jwtService: service,
		txManager:  txManager,
		logger:     logger,
	}
}

func (d *DefaultAuthService) Register(ctx context.Context, email, password string, role string) (*string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		d.logger.Error(ctx, "error generating hash from password",
			option.Any("email", email),
			option.Error(err))

		return nil, err
	}

	auth := entity.NewAuth(email, string(hash), entity.Role(role))
	if err = d.authRepo.Save(ctx, auth); err != nil {
		if errors.Is(err, persistence.ErrDuplicateKey) {
			d.logger.Error(ctx, "error inserting new auth",
				option.Any("email", email),
				option.Error(service_errors.ErrEmailExists))

			return nil, service_errors.ErrEmailExists
		}

		d.logger.Error(ctx, "failed to insert new auth",
			option.Any("email", email),
			option.Error(err))

		return nil, err
	}

	token, err := d.jwtService.GenerateToken(auth)
	if err != nil {
		d.logger.Error(ctx, "error generating token",
			option.Any("email", email),
			option.Error(err))

		return nil, err
	}

	return token, nil
}

func (d *DefaultAuthService) Login(ctx context.Context, email, password string) (*string, error) {
	auth, err := d.authRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "error getting auth by email",
				option.Any("email", email),
				option.Error(service_errors.ErrAuthWithEmailDoesNotExists))

			return nil, service_errors.ErrAuthWithEmailDoesNotExists
		}

		d.logger.Error(ctx, "failed to get auth by email",
			option.Any("email", email),
			option.Error(err))

		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(password)); err != nil {
		d.logger.Error(ctx, "wrong password",
			option.Any("email", email),
			option.Error(service_errors.ErrInvalidPasswordOrEmail))

		return nil, service_errors.ErrInvalidPasswordOrEmail
	}

	token, err := d.jwtService.GenerateToken(auth)
	if err != nil {
		d.logger.Error(ctx, "error generating token",
			option.Any("email", email),
			option.Error(err))

		return nil, err
	}

	return token, nil
}
