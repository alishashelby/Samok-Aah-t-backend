package service

import (
	"context"
	"errors"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/common"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/interfaces"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
)

type DefaultModelServiceService struct {
	modelServiceRepo interfaces.ModelServiceRepository
	userRepo         interfaces.UserRepository
	txManager        database.TxManager
	logger           pkg.Logger
}

func NewDefaultModelServiceService(modelServiceRepo interfaces.ModelServiceRepository, userRepo interfaces.UserRepository,
	txManager database.TxManager, logger pkg.Logger) *DefaultModelServiceService {
	return &DefaultModelServiceService{
		modelServiceRepo: modelServiceRepo,
		userRepo:         userRepo,
		txManager:        txManager,
		logger:           logger,
	}
}

func (d *DefaultModelServiceService) CreateService(ctx context.Context,
	title string, description string, price float32) (*entity.ModelService, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := d.checkPayloadRestrictions(price, description); err != nil {
		d.logger.Error(ctx, "check payload failed",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	service := entity.NewModelService(model.ID, title, description, price)
	if err = d.modelServiceRepo.Save(ctx, service); err != nil {
		d.logger.Error(ctx, "save model service failed",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	return service, nil
}

func (d *DefaultModelServiceService) GetServiceByID(ctx context.Context,
	serviceID int64) (*entity.ModelService, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkClientRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	service, err := d.modelServiceRepo.GetByID(ctx, serviceID, false)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "model service not found by id",
				option.Any("auth_id", authID),
				option.Any("service_id", serviceID),
				option.Error(service_errors.ErrServiceIsNotFound))

			return nil, service_errors.ErrServiceIsNotFound
		}

		d.logger.Error(ctx, "get model service failed",
			option.Any("auth_id", authID),
			option.Any("service_id", serviceID),
			option.Error(err))

		return nil, err
	}

	return service, nil
}

func (d *DefaultModelServiceService) GetAllServices(ctx context.Context,
	page, limit *int64) ([]*entity.ModelService, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkClientRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	res, err := d.modelServiceRepo.GetAll(ctx,
		entity.NewOptions(common.CheckPagination(page, limit)), false)
	if err != nil {
		d.logger.Error(ctx, "get all model services failed",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultModelServiceService) GetAllServicesByModelID(ctx context.Context,
	page, limit *int64) ([]*entity.ModelService, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	res, err := d.modelServiceRepo.GetByModelID(ctx, model.ID,
		entity.NewOptions(common.CheckPagination(page, limit)), true)
	if err != nil {
		d.logger.Error(ctx, "get all model services failed",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultModelServiceService) UpdateService(ctx context.Context, serviceID int64,
	title, description *string, price *float32) (*entity.ModelService, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	service, err := d.modelServiceRepo.GetByID(ctx, serviceID, true)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "model service not found by id",
				option.Any("auth_id", authID),
				option.Any("service_id", serviceID),
				option.Error(service_errors.ErrServiceIsNotFound))

			return nil, service_errors.ErrServiceIsNotFound
		}

		d.logger.Error(ctx, "get model service failed",
			option.Any("auth_id", authID),
			option.Any("service_id", serviceID),
			option.Error(err))

		return nil, err
	}

	newTitle := service.Title
	newDescription := service.Description
	newPrice := service.Price

	if title != nil {
		newTitle = *title
	}
	if description != nil {
		newDescription = *description
	}
	if price != nil {
		newPrice = *price
	}

	if err := d.checkPayloadRestrictions(newPrice, newDescription); err != nil {
		d.logger.Error(ctx, "check payload failed",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	if service.ModelID != model.ID {
		d.logger.Error(ctx, "model service is not owned by this model",
			option.Any("auth_id", authID),
			option.Any("service_id", serviceID),
			option.Error(service_errors.ErrModelIsNotAnOwnerOfService))

		return nil, service_errors.ErrModelIsNotAnOwnerOfService
	}

	if !service.IsActive {
		d.logger.Error(ctx, "model service is not active",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrServiceIsNotActive))

		return nil, service_errors.ErrServiceIsNotActive
	}

	hasBookings, err := d.modelServiceRepo.HasBookings(ctx, serviceID)
	if err != nil {
		d.logger.Error(ctx, "check if model service has bookings failed",
			option.Any("service_id", serviceID),
			option.Error(err))

		return nil, err
	}

	d.logger.Debug(ctx, "model service has bookings",
		option.Any("auth_id", authID),
		option.Any("service_id", serviceID))

	if !hasBookings {
		service.Title = newTitle
		service.Description = newDescription
		service.Price = newPrice

		res, err := d.modelServiceRepo.Update(ctx, service)
		if err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "model service not found by id",
					option.Any("service_id", serviceID),
					option.Error(service_errors.ErrServiceIsNotFound))

				return nil, service_errors.ErrServiceIsNotFound
			}

			d.logger.Error(ctx, "update model service failed",
				option.Any("auth_id", authID),
				option.Any("service_id", serviceID),
				option.Error(err))

			return nil, err
		}

		return res, nil
	}

	var newService *entity.ModelService
	err = d.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		err := d.modelServiceRepo.Deactivate(ctx, serviceID)
		if err != nil {
			if errors.Is(err, persistence.ErrNoRowsAffected) {
				d.logger.Error(ctx, "model service has not been deactivated",
					option.Any("service_id", serviceID),
					option.Error(service_errors.ErrServiceIsNotFound))

				return service_errors.ErrServiceIsNotFound
			}

			d.logger.Error(ctx, "deactivate model service failed",
				option.Any("auth_id", authID),
				option.Any("service_id", serviceID),
				option.Error(err))

			return err
		}

		newService = entity.NewModelService(model.ID, newTitle, newDescription, newPrice)
		if err = d.modelServiceRepo.Save(ctx, newService); err != nil {
			d.logger.Error(ctx, "save model service failed",
				option.Any("auth_id", authID),
				option.Any("service_id", serviceID),
				option.Error(err))

			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newService, nil
}

func (d *DefaultModelServiceService) DeactivateService(ctx context.Context, serviceID int64) error {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return err
	}

	service, err := d.modelServiceRepo.GetByID(ctx, serviceID, true)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "model service not found by id",
				option.Any("auth_id", authID),
				option.Any("service_id", serviceID),
				option.Error(service_errors.ErrServiceIsNotFound))

			return service_errors.ErrServiceIsNotFound
		}

		d.logger.Error(ctx, "get model service failed",
			option.Any("auth_id", authID),
			option.Any("service_id", serviceID),
			option.Error(err))

		return err
	}

	if service.ModelID != model.ID {
		d.logger.Error(ctx, "model service is not owned by this model",
			option.Any("auth_id", authID),
			option.Any("service_id", serviceID),
			option.Error(service_errors.ErrModelIsNotAnOwnerOfService))

		return service_errors.ErrModelIsNotAnOwnerOfService
	}

	if !service.IsActive {
		d.logger.Error(ctx, "model service is not active",
			option.Any("service_id", serviceID),
			option.Error(service_errors.ErrServiceIsNotActive))

		return service_errors.ErrServiceIsNotActive
	}

	err = d.modelServiceRepo.Deactivate(ctx, serviceID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsAffected) {
			d.logger.Error(ctx, "model service has not been deactivated",
				option.Any("auth_id", authID),
				option.Any("service_id", serviceID),
				option.Error(service_errors.ErrServiceIsNotFound))

			return service_errors.ErrServiceIsNotFound
		}

		d.logger.Error(ctx, "deactivate model service failed",
			option.Any("auth_id", authID),
			option.Any("service_id", serviceID),
			option.Error(err))

		return err
	}

	return nil
}

func (d *DefaultModelServiceService) checkPayloadRestrictions(
	price float32, description string) error {

	if price <= 0 {
		return service_errors.ErrInvalidPrice
	}

	if len(description) > 1000 {
		return service_errors.ErrDescriptionTooLong
	}

	return nil
}

func (d *DefaultModelServiceService) checkClientRestrictions(ctx context.Context, authID *int64) error {
	role, err := common.GetRoleFromContext(ctx)
	if err != nil {
		return err
	}

	if *role != entity.RoleClient.String() {
		d.logger.Error(ctx, "access denied",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotClient))

		return service_errors.ErrNotClient
	}

	client, err := d.userRepo.GetByAuthID(ctx, *authID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "client is not found by authID",
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrNotClient))

			return service_errors.ErrNotClient
		}

		d.logger.Error(ctx, "check client restrictions failed",
			option.Any("auth_id", authID),
			option.Error(err))

		return err
	}

	if !client.IsUserVerified() {
		d.logger.Error(ctx, "client is not verified",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotVerifiedClient))

		return service_errors.ErrNotVerifiedClient
	}

	return nil
}

func (d *DefaultModelServiceService) checkModelRestrictions(ctx context.Context, authID *int64) (*entity.User, error) {
	role, err := common.GetRoleFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if *role != entity.RoleModel.String() {
		d.logger.Error(ctx, "access denied",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotAdmin))

		return nil, service_errors.ErrNotAModel
	}

	model, err := d.userRepo.GetByAuthID(ctx, *authID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "model is not found by authID",
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrNotAModel))

			return nil, service_errors.ErrNotAModel
		}

		d.logger.Error(ctx, "check model restrictions failed",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if !model.IsUserVerified() {
		d.logger.Error(ctx, "model is not verified",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotVerifiedModel))

		return nil, service_errors.ErrNotVerifiedModel
	}

	return model, nil
}
