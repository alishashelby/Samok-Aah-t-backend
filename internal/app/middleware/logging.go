package middleware

import (
	"net/http"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
)

func LoggingMiddleware(next http.Handler, logger logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info(r.Context(), "incoming request",
			option.Any("method", r.Method),
			option.Any("url", r.URL.Path),
			option.Any("remote_addr", r.RemoteAddr),
		)

		next.ServeHTTP(w, r)
	})
}
