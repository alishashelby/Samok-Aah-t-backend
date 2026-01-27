package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/metrics"
)

func MetricsMiddleware(next http.Handler, m *metrics.Metrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.RequestsInFlight.Inc()
		defer m.RequestsInFlight.Dec()

		rr := NewResponseRecorder(w)
		next.ServeHTTP(rr, r)

		m.ObserveRequest(r.Method, strconv.Itoa(rr.StatusCode), time.Since(start))
	})
}

type responseRecorder struct {
	http.ResponseWriter
	StatusCode int
}

func NewResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

func (r *responseRecorder) WriteHeader(code int) {
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}
