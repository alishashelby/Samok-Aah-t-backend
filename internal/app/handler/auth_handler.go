package handler

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/public"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
	"github.com/go-playground/validator/v10"
)

type AuthService interface {
	Register(ctx context.Context, email, password, role string) (*string, error)
	Login(ctx context.Context, email, password string) (*string, error)
}

type AuthHandler struct {
	authService AuthService
	logger      pkg.Logger
	validate    *validator.Validate
}

func NewAuthHandler(authService AuthService, logger pkg.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
		validate:    validator.New(),
	}
}

func (h *AuthHandler) Register(ctx context.Context,
	request public.PostAuthRegisterRequestObject) (public.PostAuthRegisterResponseObject, error) {

	h.logger.Info(ctx, "AuthHandler.Register")
	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	token, err := h.authService.Register(ctx, string(request.Body.Email),
		request.Body.Password, string(request.Body.Role))
	if err != nil {
		return nil, err
	}

	return public.PostAuthRegister201JSONResponse{
		AccessToken: *token,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context,
	request public.PostAuthLoginRequestObject) (public.PostAuthLoginResponseObject, error) {

	h.logger.Info(ctx, "AuthHandler.Login")
	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	token, err := h.authService.Login(ctx, string(request.Body.Email), request.Body.Password)
	if err != nil {
		return nil, err
	}

	return public.PostAuthLogin200JSONResponse{
		AccessToken: *token,
	}, nil
}
