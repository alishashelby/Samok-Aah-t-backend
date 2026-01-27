package middleware

import (
	"context"
	"net/http"
	"strings"

	service2 "github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_const"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
)

func AuthMiddleware(next http.Handler, jwtService *service2.JWTService, logger logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := strings.TrimSpace(
			strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "),
		)

		if tokenString == "" {
			logger.Error(r.Context(), "token is empty")
			http.Error(w, "no authorization header in request", http.StatusUnauthorized)
			return
		}

		jwtClaims, err := jwtService.ParseToken(tokenString)
		if err != nil {
			logger.Error(r.Context(), "token is invalid while parsing")
			http.Error(w, "failed to parse bearer token", http.StatusUnauthorized)
			return
		}

		authIdClaim, exists := jwtClaims[string(service_const.AuthIDKey)]
		if !exists {
			logger.Error(r.Context(), "token is invalid")
			http.Error(w, "missing auth key", http.StatusUnauthorized)
			return
		}

		floatID, ok := authIdClaim.(float64)
		if !ok {
			logger.Error(r.Context(), "token is not numeric")
			http.Error(w, "authID should be numeric", http.StatusUnauthorized)
			return
		}

		roleRaw, ok := jwtClaims[string(service_const.RoleKey)]
		if !ok {
			http.Error(w, "role missing in jwt", http.StatusUnauthorized)
			return
		}

		role, ok := roleRaw.(string)
		if !ok {
			http.Error(w, "role invalid type", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), service_const.AuthIDKey, int64(floatID))
		ctx = context.WithValue(ctx, service_const.RoleKey, role)

		logger.Debug(r.Context(), "token and role are now in context")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
