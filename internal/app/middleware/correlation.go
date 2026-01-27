package middleware

import (
	"context"
	"net/http"

	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/context"
	"github.com/google/uuid"
)

func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		w.Header().Set("Correlation-ID", correlationID)
		ctx := context.WithValue(r.Context(), pkg.CorrelationID, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
