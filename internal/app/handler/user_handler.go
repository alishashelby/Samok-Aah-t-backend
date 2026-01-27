package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/authorized"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
	"github.com/go-playground/validator/v10"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

type UserService interface {
	Create(ctx context.Context, name string, birthDate time.Time) (*entity.User, error)
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByAuthID(ctx context.Context) (*entity.User, error)
	Update(ctx context.Context, name string) (*entity.User, error)
}

type UserHandler struct {
	userService UserService
	logger      pkg.Logger
	validate    *validator.Validate
}

func NewUserHandler(userService UserService, logger pkg.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
		validate:    validator.New(),
	}
}

func (h *UserHandler) CreateProfile(ctx context.Context,
	request authorized.PostUsersRequestObject) (authorized.PostUsersResponseObject, error) {

	h.logger.Info(ctx, "UserHandler.CreateProfile")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	birthDate, err := time.Parse("2006-01-02", request.Body.BirthDate)
	if err != nil {
		return nil, fmt.Errorf("invalid birth_date format: %w", err)
	}

	res, err := h.userService.Create(ctx, request.Body.Name, birthDate)
	if err != nil {
		return nil, err
	}

	return authorized.PostUsers201JSONResponse{
		Id:   res.ID,
		Name: res.Name,
		BirthDate: openapitypes.Date{
			Time: res.BirthDate,
		},
		IsVerified: res.IsVerified,
	}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context,
	request authorized.PatchUsersMeRequestObject) (authorized.PatchUsersMeResponseObject, error) {

	h.logger.Info(ctx, "UserHandler.UpdateProfile")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.userService.Update(ctx, request.Body.Name)
	if err != nil {
		return nil, err
	}

	return authorized.PatchUsersMe200JSONResponse{
		Id:   res.ID,
		Name: res.Name,
		BirthDate: openapitypes.Date{
			Time: res.BirthDate,
		},
		IsVerified: res.IsVerified,
	}, nil
}

func (h *UserHandler) GetOwnProfile(ctx context.Context,
	request authorized.GetUsersMeRequestObject) (authorized.GetUsersMeResponseObject, error) {

	h.logger.Info(ctx, "UserHandler.GetOwnProfile")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.userService.GetByAuthID(ctx)
	if err != nil {
		return nil, err
	}

	return authorized.GetUsersMe200JSONResponse{
		Id:   res.ID,
		Name: res.Name,
		BirthDate: openapitypes.Date{
			Time: res.BirthDate,
		},
		IsVerified: res.IsVerified,
	}, nil
}

func (h *UserHandler) GetSomeoneProfile(ctx context.Context,
	request authorized.GetUsersIdRequestObject) (authorized.GetUsersIdResponseObject, error) {

	h.logger.Info(ctx, "UserHandler.GetSomeoneProfile")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.userService.GetByID(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return authorized.GetUsersId200JSONResponse{
		Name:       res.Name,
		IsVerified: res.IsVerified,
	}, nil
}
