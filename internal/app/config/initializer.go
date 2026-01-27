package config

import (
	"context"
	"net/http"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/adapter"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/config/http_handler"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/handler"
	metrics2 "github.com/alishashelby/Samok-Aah-t/backend/internal/app/metrics"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/metrics"
	service2 "github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database/postgres"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/logger/config"
	persistence "github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence/postgres"
	env "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/config"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
)

type Initializer struct {
	server         *http.Server
	Logger         pkg.Logger
	metricsUpdater *metrics2.MetricsUpdater
}

func New(envConfig *env.EnvConfig, db *postgres.PostgresDb,
	txManager database.TxManager) (*Initializer, error) {
	cfg, err := config.LoadLogConfig()
	if err != nil {
		return nil, err
	}

	log, err := logger.NewDualLogger(cfg)
	if err != nil {
		return nil, err
	}

	adminRepo := persistence.NewDefaultAdminRepository(db)
	authRepo := persistence.NewDefaultAuthRepository(db)
	bookingRepo := persistence.NewDefaultBookingRepository(db)
	modelServiceRepo := persistence.NewDefaultModelServiceRepository(db)
	orderRepo := persistence.NewDefaultOrderRepository(db)
	slotRepo := persistence.NewDefaultSlotRepository(db)
	userRepo := persistence.NewDefaultUserRepository(db)

	jwtService, err := service2.NewJWTService()
	if err != nil {
		return nil, err
	}

	adminService := service2.NewDefaultAdminService(
		adminRepo, userRepo, bookingRepo, orderRepo, txManager, log)
	authService := service2.NewDefaultAuthService(
		authRepo, jwtService, txManager, log)

	bookingService, err := service2.NewDefaultBookingService(
		bookingRepo, slotRepo, userRepo, modelServiceRepo, orderRepo, txManager, log)
	if err != nil {
		return nil, err
	}

	m := metrics.NewMetrics()
	metricsUpdater := metrics2.NewMetricsUpdater(
		userRepo, m, envConfig.MetricsInterval, log)

	modelServiceService := service2.NewDefaultModelServiceService(
		modelServiceRepo, userRepo, txManager, log)
	orderService := service2.NewDefaultOrderService(
		orderRepo, bookingRepo, slotRepo, userRepo, modelServiceRepo, txManager, log, m)
	slotService := service2.NewDefaultSlotService(
		slotRepo, bookingRepo, userRepo, txManager, log)
	userService := service2.NewDefaultUserService(userRepo, txManager, log)

	adminHandler := handler.NewAdminHandler(adminService, log)
	authHandler := handler.NewAuthHandler(authService, log)
	bookingHandler := handler.NewBookingHandler(bookingService, log)
	orderHandler := handler.NewOrderHandler(orderService, log)
	modelServiceHandler := handler.NewModelServiceHandler(modelServiceService, log)
	slotHandler := handler.NewSlotHandler(slotService, log)
	userHandler := handler.NewUserHandler(userService, log)

	publicAdapter := adapter.NewPublicAdapter(authHandler)
	authorizedAdapter := adapter.NewAuthorizedAdapter(
		userHandler, modelServiceHandler, slotHandler, bookingHandler, &orderHandler, adminHandler)
	r := http_handler.BuildHTTPHandler(publicAdapter, authorizedAdapter, jwtService, m, log)

	return &Initializer{
		server: &http.Server{
			Addr:    ":" + envConfig.Port,
			Handler: r,
		},
		Logger:         log,
		metricsUpdater: metricsUpdater,
	}, nil
}

func (i *Initializer) Run(ctx context.Context) error {
	go i.metricsUpdater.Start(ctx)

	if err := i.server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (i *Initializer) Shutdown(ctx context.Context) error {
	return i.server.Shutdown(ctx)
}
