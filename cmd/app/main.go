package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	initializer "github.com/alishashelby/Samok-Aah-t/backend/internal/app/config"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database/postgres"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/config"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	envConfig, err := config.LoadEnv()
	if err != nil {
		log.Printf("error loading env vars: %v", err)
		return
	}

	db, err := postgres.New(ctx, envConfig)
	if err != nil {
		log.Printf("error connecting to database: %v", err)
		return
	}
	defer db.Close()

	txManager := postgres.NewPostgresTxManager(db.Pool)
	app, err := initializer.New(envConfig, db, txManager)
	if err != nil {
		log.Printf("error initializing app: %v", err)
		return
	}

	app.Logger.Info(ctx, "start server running",
		option.Any("port", envConfig.Port))

	serverError := make(chan error, 1)

	go func() {
		serverError <- app.Run(ctx)
	}()

	select {
	case err = <-serverError:
		if err != nil {
			app.Logger.Error(ctx, "failed to start server",
				option.Error(err))
		}
	case <-ctx.Done():
		app.Logger.Info(ctx, "shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(), envConfig.ShutdownTimeout)
	defer cancel()

	if err = app.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error(ctx, "failed to shutdown server",
			option.Error(err))
	} else {
		app.Logger.Info(ctx, "server stopped successfully")
	}
}
