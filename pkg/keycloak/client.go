package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/keycloak-broker/pkg/catalog"
	"github.com/keycloak-broker/pkg/config"
	"github.com/keycloak-broker/pkg/logger"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	realm      string
	admin      string
	password   string

	mu    sync.Mutex
	token string
	expAt time.Time

	discovery *OIDCDiscoveryResponse
}

func NewClient() *Client {
	cfg := config.Get()
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    strings.TrimRight(cfg.KeycloakURL, "/"),
		realm:      cfg.KeycloakRealm,
		admin:      cfg.KeycloakAdmin,
		password:   cfg.KeycloakPassword,
	}
}

// CreateClient creates an OIDC client in Keycloak and returns it.
func (c *Client) CreateClient(ctx context.Context, instanceId, serviceId, planId string, parameters *OIDCClientParameters) (*OIDCClientResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}

	// store service_id and plan_id inside client attributes in Keycloak
	attributes := make(map[string]string)
	attributes["service_id"] = serviceId
	attributes["plan_id"] = planId
	if parameters.PKCEEnabled == nil || *parameters.PKCEEnabled {
		attributes["pkce.code.challenge.method"] = "S256"
	}
	if parameters.RefreshTokenLifetime > 0 {
		attributes["client.session.max.lifespan"] = fmt.Sprintf("%d", parameters.RefreshTokenLifetime)
	}
	if parameters.AccessTokenLifetime > 0 {
		attributes["access.token.lifespan"] = fmt.Sprintf("%d", parameters.AccessTokenLifetime)
	}

	oidc := OIDCClientPayload{
		ClientId:                  instanceId,
		Name:                      instanceId,
		Description:               catalog.GetPlan(serviceId, planId).Description,
		Enabled:                   true,
		Protocol:                  "openid-connect",
		PublicClient:              parameters.PublicClient,
		RedirectURIs:              parameters.RedirectURIs,
		ConsentRequired:           parameters.ConsentRequired,
		StandardFlowEnabled:       parameters.StandardFlowEnabled == nil || *parameters.StandardFlowEnabled,
		ImplicitFlowEnabled:       parameters.ImplicitFlowEnabled,
		DirectAccessGrantsEnabled: parameters.DirectAccessGrantsEnabled,
		ServiceAccountsEnabled:    parameters.ServiceAccountsEnabled,
		Attributes:                attributes,
	}
	body, err := json.Marshal(oidc)
	if err != nil {
		return nil, fmt.Errorf("marshal client failure: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/admin/realms/%s/clients", c.baseURL, c.realm),
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build create request for client instance_id [%s]: %w", instanceId, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create client instance_id [%s]: %w", instanceId, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("client with instance_id [%s] already exists", instanceId)
	}
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create client instance_id [%s] failed (%d): %s", instanceId, resp.StatusCode, respBody)
	}

	logger.Info("created keycloak client with instance_id [%s]", instanceId)
	return c.GetClient(ctx, instanceId)
}

// GetClient looks up an OIDC client by instanceId and returns it.
func (c *Client) GetClient(ctx context.Context, instanceId string) (*OIDCClientResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/admin/realms/%s/clients?clientId=%s", c.baseURL, c.realm, url.QueryEscape(instanceId)),
		nil)
	if err != nil {
		return nil, fmt.Errorf("build get request for client instance_id [%s]: %w", instanceId, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get client instance_id [%s]: %w", instanceId, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get client instance_id [%s] failed (%d): %s", instanceId, resp.StatusCode, respBody)
	}

	var clients []OIDCClientResponse
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, fmt.Errorf("decode clients for instance_id [%s]: %w", instanceId, err)
	}

	// find client
	for _, cl := range clients {
		if cl.ClientId == instanceId {
			if len(cl.Issuer) == 0 {
				cl.Issuer = fmt.Sprintf("%s/realms/%s", c.baseURL, c.realm)
			}
			if len(cl.DiscoveryEndpoint) == 0 {
				cl.DiscoveryEndpoint = fmt.Sprintf("%s/realms/%s/.well-known/openid-configuration", c.baseURL, c.realm)
			}

			// get all other endpoints by calling Keycloaks ".well-known" discovery endpoint
			endpoints, err := c.getEndpoints(ctx)
			if err != nil {
				return nil, fmt.Errorf("discovery endpoint request failure: %w", err)
			}
			if len(cl.AuthorizationEndpoint) == 0 {
				cl.AuthorizationEndpoint = endpoints.AuthorizationEndpoint
			}
			if len(cl.TokenEndpoint) == 0 {
				cl.TokenEndpoint = endpoints.TokenEndpoint
			}
			if len(cl.IntrospectionEndpoint) == 0 {
				cl.IntrospectionEndpoint = endpoints.IntrospectionEndpoint
			}
			if len(cl.UserInfoEndpoint) == 0 {
				cl.UserInfoEndpoint = endpoints.UserInfoEndpoint
			}
			if len(cl.EndSessionEndpoint) == 0 {
				cl.EndSessionEndpoint = endpoints.EndSessionEndpoint
			}
			if len(cl.JWKSURI) == 0 {
				cl.JWKSURI = endpoints.JWKSURI
			}
			return &cl, nil
		}
	}
	return nil, fmt.Errorf("get client instance_id [%s]: %w", instanceId, ErrNotFound)
}

// UpdateClient updates an existing OIDC client's parameters in Keycloak.
func (c *Client) UpdateClient(ctx context.Context, instanceId string, update *OIDCClientUpdatePayload) (*OIDCClientResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := c.GetClient(ctx, instanceId)
	if err != nil {
		return nil, err
	}

	// merge update into existing client to avoid wiping fields on PUT
	merged := OIDCClientPayload{
		ClientId:                  existing.ClientId,
		Name:                      existing.Name,
		Description:               existing.Description,
		Enabled:                   existing.Enabled,
		Protocol:                  existing.Protocol,
		PublicClient:              existing.PublicClient,
		RedirectURIs:              existing.RedirectURIs,
		ConsentRequired:           existing.ConsentRequired,
		StandardFlowEnabled:       existing.StandardFlowEnabled,
		ImplicitFlowEnabled:       existing.ImplicitFlowEnabled,
		DirectAccessGrantsEnabled: existing.DirectAccessGrantsEnabled,
		ServiceAccountsEnabled:    existing.ServiceAccountsEnabled,
		Attributes:                existing.Attributes,
	}
	if update.RedirectURIs != nil {
		merged.RedirectURIs = update.RedirectURIs
	}
	if update.ConsentRequired != nil {
		merged.ConsentRequired = *update.ConsentRequired
	}
	if update.StandardFlowEnabled != nil {
		merged.StandardFlowEnabled = *update.StandardFlowEnabled
	}
	if update.ImplicitFlowEnabled != nil {
		merged.ImplicitFlowEnabled = *update.ImplicitFlowEnabled
	}
	if update.DirectAccessGrantsEnabled != nil {
		merged.DirectAccessGrantsEnabled = *update.DirectAccessGrantsEnabled
	}
	if update.ServiceAccountsEnabled != nil {
		merged.ServiceAccountsEnabled = *update.ServiceAccountsEnabled
	}
	if update.PKCEEnabled != nil {
		if merged.Attributes == nil {
			merged.Attributes = make(map[string]string)
		}
		if *update.PKCEEnabled {
			merged.Attributes["pkce.code.challenge.method"] = "S256"
		} else {
			delete(merged.Attributes, "pkce.code.challenge.method")
		}
	}
	if update.RefreshTokenLifetime != nil {
		if merged.Attributes == nil {
			merged.Attributes = make(map[string]string)
		}
		if *update.RefreshTokenLifetime > 0 {
			merged.Attributes["client.session.max.lifespan"] = fmt.Sprintf("%d", *update.RefreshTokenLifetime)
		} else {
			delete(merged.Attributes, "client.session.max.lifespan")
		}
	}
	if update.AccessTokenLifetime != nil {
		if merged.Attributes == nil {
			merged.Attributes = make(map[string]string)
		}
		if *update.AccessTokenLifetime > 0 {
			merged.Attributes["access.token.lifespan"] = fmt.Sprintf("%d", *update.AccessTokenLifetime)
		} else {
			delete(merged.Attributes, "access.token.lifespan")
		}
	}

	body, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal update payload failure: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		fmt.Sprintf("%s/admin/realms/%s/clients/%s", c.baseURL, c.realm, existing.Id),
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build update request for client instance_id [%s]: %w", instanceId, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("update client instance_id [%s]: %w", instanceId, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("update client instance_id [%s] failed (%d): %s", instanceId, resp.StatusCode, respBody)
	}

	logger.Info("updated keycloak client with instance_id [%s]", instanceId)
	return c.GetClient(ctx, instanceId)
}

// DeleteClient removes an OIDC client by clientId.
func (c *Client) DeleteClient(ctx context.Context, instanceId string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return err
	}

	// we need the internal UUID to delete the client
	oidc, err := c.GetClient(ctx, instanceId)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/admin/realms/%s/clients/%s", c.baseURL, c.realm, oidc.Id),
		nil)
	if err != nil {
		return fmt.Errorf("build delete request for client instance_id [%s]: %w", instanceId, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete client instance_id [%s]: %w", instanceId, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		logger.Info("deleted keycloak client instance_id [%s]", instanceId)
		return nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("delete client failed (%d): %s", resp.StatusCode, respBody)
}
