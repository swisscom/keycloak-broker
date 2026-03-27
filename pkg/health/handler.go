package health

import (
	"context"
	"net/http"

	"github.com/keycloak-broker/pkg/keycloak"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	client *keycloak.Client
}

func New() *Handler {
	return &Handler{
		client: keycloak.NewClient(),
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("")

	g.GET("/health", h.healthz)
	g.GET("/healthz", h.healthz)
}

func (h *Handler) healthz(c echo.Context) error {
	if h.client == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"status": "unhealthy",
		})
	}
	if err := h.client.HealthCheck(context.Background()); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"status": "unhealthy",
			"reason": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
	})
}
