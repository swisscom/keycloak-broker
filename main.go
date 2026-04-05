package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/keycloak-broker/pkg/config"
	"github.com/keycloak-broker/pkg/logger"
	"github.com/keycloak-broker/pkg/router"
)

func main() {
	cfg := config.Get()
	logger.Init()
	logger.Info("starting keycloak-broker on port %d", cfg.Port)

	r := router.New()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := r.Start(cfg.Port); err != nil {
			logger.Info("shutting down: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Info("received shutdown signal")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := r.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("forced shutdown: %v", err)
	}
	logger.Info("shutdown complete")
}
