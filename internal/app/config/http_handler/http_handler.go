package http_handler

import (
	"net/http"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/adapter"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/authorized"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/public"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/config/server"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/handler"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/middleware"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/metrics"
	service2 "github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/gorilla/mux"
)

func BuildHTTPHandler(
	publicAdapter *adapter.PublicAdapter,
	authorizedAdapter *adapter.AuthorizedAdapter,
	jwtService *service2.JWTService,
	m *metrics.Metrics,
	logger pkg.Logger,
) http.Handler {
	mapper := handler.NewErrorMapper()

	r := mux.NewRouter()
	r.Handle("/metrics", m.Handler())
	r.Use(
		middleware.CorrelationMiddleware,
		func(next http.Handler) http.Handler { return middleware.MetricsMiddleware(next, m) },
		func(next http.Handler) http.Handler { return middleware.LoggingMiddleware(next, logger) },
		func(next http.Handler) http.Handler { return middleware.PanicMiddleware(next, logger) },
	)

	public.HandlerWithOptions(
		public.NewStrictHandlerWithOptions(publicAdapter, nil, public.StrictHTTPServerOptions{
			ResponseErrorHandlerFunc: server.NewResponseErrorHandler(mapper),
		}),
		public.GorillaServerOptions{BaseRouter: r},
	)

	authorizedRouter := r.PathPrefix("/").Subrouter()
	authorizedRouter.Use(func(next http.Handler) http.Handler {
		return middleware.AuthMiddleware(next, jwtService, logger)
	})
	authorized.HandlerWithOptions(
		authorized.NewStrictHandlerWithOptions(authorizedAdapter, nil, authorized.StrictHTTPServerOptions{
			ResponseErrorHandlerFunc: server.NewResponseErrorHandler(mapper),
		}),
		authorized.GorillaServerOptions{BaseRouter: authorizedRouter},
	)

	return r
}
