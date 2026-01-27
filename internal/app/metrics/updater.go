package metrics

import (
	"context"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/interfaces"
	service2 "github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/metrics"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
)

type MetricsUpdater struct {
	userRepo interfaces.UserRepository
	metrics  *service2.Metrics
	interval time.Duration
	logger   pkg.Logger
}

func NewMetricsUpdater(userRepo interfaces.UserRepository, metrics *service2.Metrics,
	interval time.Duration, logger pkg.Logger) *MetricsUpdater {
	return &MetricsUpdater{
		userRepo: userRepo,
		metrics:  metrics,
		interval: interval,
		logger:   logger,
	}
}

func (m *MetricsUpdater) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				m.logger.Info(ctx, "metrics updater stopped")
				return

			case <-ticker.C:
				m.update(ctx)
			}
		}
	}()
}

func (m *MetricsUpdater) update(ctx context.Context) {
	clients, err := m.userRepo.CountByRole(ctx, entity.RoleClient)
	if err != nil {
		m.logger.Error(ctx, "failed to count clients", option.Error(err))
		return
	}

	models, err := m.userRepo.CountByRole(ctx, entity.RoleModel)
	if err != nil {
		m.logger.Error(ctx, "failed to count models", option.Error(err))
		return
	}

	m.metrics.SetClients(clients)
	m.metrics.SetModels(models)
}
