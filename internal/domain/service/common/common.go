package common

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_const"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
)

func GetAuthIDFromContext(ctx context.Context) (*int64, error) {
	authID, ok := ctx.Value(service_const.AuthIDKey).(int64)
	if !ok {
		return nil, service_errors.ErrUnauthorized
	}

	return &authID, nil
}

func GetRoleFromContext(ctx context.Context) (*string, error) {
	res, ok := ctx.Value(service_const.RoleKey).(string)
	if !ok {
		return nil, service_errors.ErrNoRole
	}

	return &res, nil
}

func CheckPagination(page, limit *int64) (resPage int64, resLimit int64) {
	if page != nil {
		resPage = *page
	}

	if limit != nil {
		resLimit = *limit
	}

	return
}
