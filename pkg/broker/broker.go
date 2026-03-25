package broker

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/keycloak-broker/pkg/catalog"
	"github.com/keycloak-broker/pkg/keycloak"
	"github.com/keycloak-broker/pkg/logger"
	"github.com/keycloak-broker/pkg/validation"
	"github.com/labstack/echo/v4"
)

type Broker struct {
	client *keycloak.Client
}

func NewBroker() *Broker {
	return &Broker{
		client: keycloak.NewClient(),
	}
}

func (b *Broker) GetCatalog(c echo.Context) error {
	logger.Debug("catalog requested")
	return c.JSON(http.StatusOK, catalog.GetCatalog())
}

func (b *Broker) ProvisionInstance(c echo.Context) error {
	instanceId := c.Param("instance_id")
	if err := validation.ValidateInstanceID(instanceId); err != nil {
		logger.Warn("invalid instance_id: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var req struct {
		ServiceID string         `json:"service_id"`
		PlanID    string         `json:"plan_id"`
		Context   map[string]any `json:"context"`
	}
	if err := c.Bind(&req); err != nil {
		logger.Error("failed to parse provision request for instance_id [%s]: %v", instanceId, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := validation.ValidateServiceID(req.ServiceID); err != nil {
		logger.Warn("invalid service_id [%s] for %s: %v", req.ServiceID, instanceId, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := validation.ValidatePlanID(req.ServiceID, req.PlanID); err != nil {
		logger.Warn("invalid plan_id [%s] for %s: %v", req.PlanID, instanceId, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// TODO: provision OIDC client in keycloak

	logger.Info("instance_id [%s] provisioned", instanceId)
	return c.JSON(http.StatusAccepted, map[string]any{}) // TODO: verify return status according to OSB spec
}

func (b *Broker) GetInstance(c echo.Context) error {
	instanceId := c.Param("instance_id")
	if err := validation.ValidateInstanceID(instanceId); err != nil {
		logger.Warn("invalid instance_id: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	logger.Debug("checking instance_id [%s]", instanceId)
	client, err := b.client.GetClient(context.Background(), instanceId)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("\"%s\" not found", instanceId)) {
			logger.Debug("instance_id [%s] not found", instanceId)
			return c.JSON(http.StatusNotFound, map[string]any{})
		} else {
			logger.Error("failed to get instance %s: %v", instanceId, err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	return c.JSON(http.StatusOK, client)
}

func (b *Broker) DeprovisionInstance(c echo.Context) error {
	instanceId := c.Param("instance_id")
	if err := validation.ValidateInstanceID(instanceId); err != nil {
		logger.Warn("invalid instance_id: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	logger.Info("deprovision instance_id [%s]", instanceId)
	err := b.client.DeleteClient(context.Background(), instanceId)
	if err != nil {
		logger.Error("failed to deprovision instance_id [%s]: %v", instanceId, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	logger.Info("deprovisioned instance_id [%s]", instanceId)
	return c.JSON(http.StatusAccepted, map[string]any{}) // TODO: verify return status according to OSB spec
}

func (b *Broker) BindInstance(c echo.Context) error {
	return nil
}

func (b *Broker) GetBinding(c echo.Context) error {
	return nil
}

func (b *Broker) UnbindInstance(c echo.Context) error {
	return nil
}
