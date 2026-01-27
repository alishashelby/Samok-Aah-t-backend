package middleware

import (
	"net/http"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
)

func PanicMiddleware(next http.Handler, logger logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Info(r.Context(), "panic middleware",
					option.Any("recovered from:", err))

				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
