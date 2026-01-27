package config

import (
	"fmt"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/pkg/config"
)

const (
	defaultPostgresPort    = "5432"
	defaultAppPort         = "8080"
	defaultShutdownTimeout = "10s"
	defaultMetricsInterval = "30s"
)

type EnvConfig struct {
	Port             string
	PostgresUser     string
	PostgresPassword string
	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
	ShutdownTimeout  time.Duration
	MetricsInterval  time.Duration
}

func LoadEnv() (*EnvConfig, error) {
	port := config.GetEnvVariableOrDefault("PORT", defaultAppPort)

	postgresUser, err := config.GetEnvVariable("POSTGRES_USER")
	if err != nil {
		return nil, err
	}

	postgresPassword, err := config.GetEnvVariable("POSTGRES_PASSWORD")
	if err != nil {
		return nil, err
	}

	postgresHost, err := config.GetEnvVariable("POSTGRES_HOST")
	if err != nil {
		return nil, err
	}

	postgresPort := config.GetEnvVariableOrDefault("POSTGRES_PORT", defaultPostgresPort)

	postgresDB, err := config.GetEnvVariable("POSTGRES_DB")
	if err != nil {
		return nil, err
	}

	shutdownTimeoutStr := config.GetEnvVariableOrDefault("SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	shutdownTimeout, err := time.ParseDuration(shutdownTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid value for SHUTDOWN_TIMEOUT: %w", err)
	}

	metricsIntervalStr := config.GetEnvVariableOrDefault("METRICS_INTERVAL", defaultMetricsInterval)
	metricsInterval, err := time.ParseDuration(metricsIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid value for SHUTDOWN_TIMEOUT: %w", err)
	}

	return &EnvConfig{
		Port:             port,
		PostgresUser:     postgresUser,
		PostgresPassword: postgresPassword,
		PostgresHost:     postgresHost,
		PostgresPort:     postgresPort,
		PostgresDB:       postgresDB,
		ShutdownTimeout:  shutdownTimeout,
		MetricsInterval:  metricsInterval,
	}, nil
}
