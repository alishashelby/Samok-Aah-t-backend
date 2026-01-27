package service

import (
	"context"
	"errors"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/common"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/interfaces"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
)

type DefaultUserService struct {
	userRepo  interfaces.UserRepository
	txManager database.TxManager
	logger    pkg.Logger
}

func NewDefaultUserService(repo interfaces.UserRepository, txManager database.TxManager,
	logger pkg.Logger) *DefaultUserService {
	return &DefaultUserService{
		userRepo:  repo,
		txManager: txManager,
		logger:    logger,
	}
}

func (d *DefaultUserService) Create(ctx context.Context,
	name string, birthDate time.Time) (*entity.User, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	res := entity.NewUser(*authID, name, birthDate)
	err = d.userRepo.Save(ctx, res)
	if err != nil {
		d.logger.Error(ctx, "failed to save user",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultUserService) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	user, err := d.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "user not found",
				option.Any("id", id),
				option.Error(service_errors.ErrUserNotFound))

			return nil, service_errors.ErrUserNotFound
		}

		d.logger.Error(ctx, "failed to get user",
			option.Any("id", id),
			option.Error(err))

		return nil, err
	}

	return user, nil
}

func (d *DefaultUserService) GetByAuthID(ctx context.Context) (*entity.User, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := d.userRepo.GetByAuthID(ctx, *authID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "user not found",
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrUserNotFound))

			return nil, service_errors.ErrUserNotFound
		}

		d.logger.Error(ctx, "failed to get user",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	return user, nil
}

func (d *DefaultUserService) Update(ctx context.Context, name string) (*entity.User, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := d.userRepo.GetByAuthID(ctx, *authID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "user not found by auth id",
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrUserNotFound))

			return nil, service_errors.ErrUserNotFound
		}

		d.logger.Error(ctx, "failed to get user",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	user.Name = name

	res, err := d.userRepo.Update(ctx, user)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "user not found by auth id",
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrUserNotFound))

			return nil, service_errors.ErrUserNotFound
		}

		d.logger.Error(ctx, "failed to update user",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	return res, nil
}
