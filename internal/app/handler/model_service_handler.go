package handler

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/authorized"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/models"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
	"github.com/go-playground/validator/v10"
)

type ModelServiceService interface {
	CreateService(ctx context.Context,
		title string, description string, price float32) (*entity.ModelService, error)
	GetServiceByID(ctx context.Context, serviceID int64) (*entity.ModelService, error)
	GetAllServices(ctx context.Context, page, limit *int64) ([]*entity.ModelService, error)
	GetAllServicesByModelID(ctx context.Context, page, limit *int64) ([]*entity.ModelService, error)
	UpdateService(ctx context.Context, serviceID int64,
		title, description *string, price *float32) (*entity.ModelService, error)
	DeactivateService(ctx context.Context, serviceID int64) error
}

type ModelServiceHandler struct {
	modelServiceService ModelServiceService
	logger              pkg.Logger
	validate            *validator.Validate
}

func NewModelServiceHandler(modelServiceService ModelServiceService,
	logger pkg.Logger) *ModelServiceHandler {
	return &ModelServiceHandler{
		modelServiceService: modelServiceService,
		logger:              logger,
		validate:            validator.New(),
	}
}

func (h *ModelServiceHandler) CreateService(ctx context.Context,
	request authorized.PostModelServicesRequestObject,
) (authorized.PostModelServicesResponseObject, error) {

	h.logger.Info(ctx, "ModelServiceHandler.CreateService")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.modelServiceService.CreateService(
		ctx, request.Body.Title, request.Body.Description, request.Body.Price)
	if err != nil {
		return nil, err
	}

	return authorized.PostModelServices201JSONResponse{
		Id:          res.ID,
		ModelId:     res.ModelID,
		Title:       res.Title,
		Description: res.Description,
		Price:       res.Price,
		IsActive:    res.IsActive,
		CreatedAt:   res.CreatedAt,
	}, nil
}

func (h *ModelServiceHandler) GetAllServices(ctx context.Context,
	request authorized.GetClientServicesRequestObject) (authorized.GetClientServicesResponseObject, error) {

	h.logger.Info(ctx, "ModelServiceHandler.GetAllServices")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	services, err := h.modelServiceService.GetAllServices(ctx,
		request.Params.Page, request.Params.Limit)
	if err != nil {
		return nil, err
	}

	res := make(authorized.GetClientServices200JSONResponse, len(services))
	for i, s := range services {
		res[i] = models.ModelServiceResponse{
			Id:          s.ID,
			ModelId:     s.ModelID,
			Title:       s.Title,
			Description: s.Description,
			Price:       s.Price,
			IsActive:    s.IsActive,
			CreatedAt:   s.CreatedAt,
		}
	}

	return res, nil
}

func (h *ModelServiceHandler) GetServiceByID(ctx context.Context,
	request authorized.GetClientServicesIdRequestObject) (authorized.GetClientServicesIdResponseObject, error) {

	h.logger.Info(ctx, "ModelServiceHandler.GetServiceByID")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.modelServiceService.GetServiceByID(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return authorized.GetClientServicesId200JSONResponse{
		Id:          res.ID,
		ModelId:     res.ModelID,
		Title:       res.Title,
		Description: res.Description,
		Price:       res.Price,
		IsActive:    res.IsActive,
		CreatedAt:   res.CreatedAt,
	}, nil
}

func (h *ModelServiceHandler) GetModelServices(ctx context.Context,
	request authorized.GetModelServicesRequestObject) (authorized.GetModelServicesResponseObject, error) {

	h.logger.Info(ctx, "ModelServiceHandler.GetModelServices")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	services, err := h.modelServiceService.GetAllServicesByModelID(ctx,
		request.Params.Page, request.Params.Limit)
	if err != nil {
		return nil, err
	}

	res := make(authorized.GetModelServices200JSONResponse, len(services))
	for i, s := range services {
		res[i] = models.ModelServiceResponse{
			Id:          s.ID,
			ModelId:     s.ModelID,
			Title:       s.Title,
			Description: s.Description,
			Price:       s.Price,
			IsActive:    s.IsActive,
			CreatedAt:   s.CreatedAt,
		}
	}

	return res, nil
}

func (h *ModelServiceHandler) UpdateService(ctx context.Context,
	request authorized.PatchModelServicesIdRequestObject) (authorized.PatchModelServicesIdResponseObject, error) {

	h.logger.Info(ctx, "ModelServiceHandler.UpdateService")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.modelServiceService.UpdateService(
		ctx, request.Id, request.Body.Title, request.Body.Description, request.Body.Price)
	if err != nil {
		return nil, err
	}

	return authorized.PatchModelServicesId200JSONResponse{
		Id:          res.ID,
		ModelId:     res.ModelID,
		Title:       res.Title,
		Description: res.Description,
		Price:       res.Price,
		IsActive:    res.IsActive,
		CreatedAt:   res.CreatedAt,
	}, nil
}

func (h *ModelServiceHandler) DeactivateService(ctx context.Context,
	request authorized.PatchModelServicesIdDeactivateRequestObject,
) (authorized.PatchModelServicesIdDeactivateResponseObject, error) {

	h.logger.Info(ctx, "ModelServiceHandler.DeactivateService")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	if err := h.modelServiceService.DeactivateService(ctx, request.Id); err != nil {
		return nil, err
	}

	return authorized.PatchModelServicesIdDeactivate200JSONResponse{
		Status: "deactivated",
	}, nil
}
