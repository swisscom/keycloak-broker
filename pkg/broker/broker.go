package broker

import (
	"context"
	"errors"
	"net/http"

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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation", "description": err.Error()})
	}

	// read in request parameters
	var req struct {
		ServiceID  string         `json:"service_id"`
		PlanID     string         `json:"plan_id"`
		Context    map[string]any `json:"context"`
		Parameters struct {
			RedirectURIs        []string `json:"redirect_uris"`
			ImplicitFlowEnabled bool     `json:"implicit_flow_enabled"`
			ConsentRequired     bool     `json:"consent_required"`
		} `json:"parameters"`
	}
	if err := c.Bind(&req); err != nil {
		logger.Error("failed to parse provision request for instance_id [%s]: %v", instanceId, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "payload", "description": err.Error()})
	}

	// validate service and plan IDs
	if err := validation.ValidateServiceID(req.ServiceID); err != nil {
		logger.Warn("invalid service_id [%s] for %s: %v", req.ServiceID, instanceId, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation", "description": err.Error()})
	}
	if err := validation.ValidatePlanID(req.ServiceID, req.PlanID); err != nil {
		logger.Warn("invalid plan_id [%s] for %s: %v", req.PlanID, instanceId, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation", "description": err.Error()})
	}

	// check first if instance_id already exists
	logger.Debug("checking if instance_id [%s] exists", instanceId)
	client, err := b.client.GetClient(context.Background(), instanceId)
	if err != nil {
		if !errors.Is(err, keycloak.ErrNotFound) {
			logger.Error("failed to get instance_id [%s]: %v", instanceId, err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fetch", "description": err.Error()})
		}
	}
	if client != nil && client.ClientId == instanceId {
		// it already exists, return data with HTTP 200
		logger.Info("instance_id [%s] already exists", instanceId)
		return c.JSON(http.StatusOK, keycloakClientToOSB(client))
	}

	client, err = b.client.CreateClient(context.Background(),
		instanceId, req.ServiceID, req.PlanID,
		&keycloak.OIDCClientParameters{
			RedirectURIs:        req.Parameters.RedirectURIs,
			PublicClient:        catalog.IsPublicClient(req.ServiceID, req.PlanID),
			ImplicitFlowEnabled: req.Parameters.ImplicitFlowEnabled,
			ConsentRequired:     req.Parameters.ConsentRequired,
		})
	if err != nil {
		logger.Error("failed to provision instance_id [%s]: %v", instanceId, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "provision", "description": err.Error()})
	}

	// newly created, return with HTTP 201
	logger.Info("instance_id [%s] provisioned", instanceId)
	return c.JSON(http.StatusCreated, keycloakClientToOSB(client))
}

func (b *Broker) GetInstance(c echo.Context) error {
	instanceId := c.Param("instance_id")
	if err := validation.ValidateInstanceID(instanceId); err != nil {
		logger.Warn("invalid instance_id: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation", "description": err.Error()})
	}

	logger.Debug("checking if instance_id [%s] exists", instanceId)
	client, err := b.client.GetClient(context.Background(), instanceId)
	if err != nil {
		if errors.Is(err, keycloak.ErrNotFound) {
			logger.Debug("instance_id [%s] not found", instanceId)
			return c.JSON(http.StatusNotFound, map[string]any{})
		}
		logger.Error("failed to get instance_id [%s]: %v", instanceId, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fetch", "description": err.Error()})
	}
	return c.JSON(http.StatusOK, client)
}

func (b *Broker) DeprovisionInstance(c echo.Context) error {
	instanceId := c.Param("instance_id")
	if err := validation.ValidateInstanceID(instanceId); err != nil {
		logger.Warn("invalid instance_id: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation", "description": err.Error()})
	}

	logger.Info("deprovision instance_id [%s]", instanceId)
	err := b.client.DeleteClient(context.Background(), instanceId)
	if err != nil {
		if errors.Is(err, keycloak.ErrNotFound) {
			logger.Debug("instance_id [%s] not found, already gone", instanceId)
			return c.JSON(http.StatusGone, map[string]any{})
		}
		logger.Error("failed to deprovision instance_id [%s]: %v", instanceId, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "deprovision", "description": err.Error()})
	}

	logger.Info("deprovisioned instance_id [%s]", instanceId)
	return c.JSON(http.StatusOK, map[string]any{})
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
