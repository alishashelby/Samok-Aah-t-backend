package adapter

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/public"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/handler"
)

type PublicAdapter struct {
	Auth *handler.AuthHandler
}

func NewPublicAdapter(auth *handler.AuthHandler) *PublicAdapter {
	return &PublicAdapter{
		Auth: auth,
	}
}

func (p *PublicAdapter) PostAuthLogin(ctx context.Context,
	request public.PostAuthLoginRequestObject) (public.PostAuthLoginResponseObject, error) {
	return p.Auth.Login(ctx, request)
}

func (p *PublicAdapter) PostAuthRegister(ctx context.Context,
	request public.PostAuthRegisterRequestObject) (public.PostAuthRegisterResponseObject, error) {
	return p.Auth.Register(ctx, request)
}
