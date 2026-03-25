package main

import (
	"github.com/keycloak-broker/pkg/config"
	"github.com/keycloak-broker/pkg/logger"
	"github.com/keycloak-broker/pkg/router"
)

func main() {
	cfg := config.Get()
	logger.Init()
	logger.Info("starting keycloak-broker on port %d", cfg.Port)

	r := router.New()
	logger.Fatal("failed to start HTTP router: %v", r.Start(cfg.Port))
}
